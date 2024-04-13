package container

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/dennishilgert/apollo/pkg/defers"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/utils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/api/types/strslice"
	docker "github.com/docker/docker/client"
	dockerArchive "github.com/docker/docker/pkg/archive"
)

var config = LoadConfig()

// GetDefaultClient returns a default instance of the Docker client.
func GetDefaultClient() (*docker.Client, error) {
	return docker.NewClientWithOpts(docker.WithAPIVersionNegotiation())
}

// ContainerStart creates and starts a container and adds its stop and removal to the cleanup defers.
func ContainerStart(ctx context.Context, client *docker.Client, log logger.Logger, cleanup defers.Defers, containerConfig container.Config, hostConfig container.HostConfig, name string, destroyAfter bool) (*string, error) {
	log.Debugf("creating container from image: %s", containerConfig.Image)
	_, err := FetchImageIdByTag(ctx, client, log, containerConfig.Image)
	if err != nil {
		log.Debugf("image for container is not available locally - pulling: %s", containerConfig.Image)
		if err := ImagePull(ctx, client, log, containerConfig.Image); err != nil {
			return nil, fmt.Errorf("failed to pull docker image: %v", err)
		}
	}
	containerCreateResponse, err := client.ContainerCreate(ctx, &containerConfig, &hostConfig, nil, nil, name)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker container: %v", err)
	}

	log = log.WithFields(map[string]any{"container-id": containerCreateResponse.ID[:12]})

	log.Debug("starting container")
	if err := client.ContainerStart(ctx, containerCreateResponse.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start docker container: %v", err)
	}
	if destroyAfter {
		cleanup.Add(func() {
			ContainerRemove(context.Background(), client, log, containerCreateResponse.ID)
		})
		cleanup.Add(func() {
			ContainerStop(context.Background(), client, log, containerCreateResponse.ID)
		})
	}
	return &containerCreateResponse.ID, nil
}

// ContainerStop stops a container gracefully or kills it after a timeout.
func ContainerStop(ctx context.Context, client *docker.Client, log logger.Logger, containerId string) {
	log.Debug("stopping container")
	go func() {
		if err := client.ContainerStop(ctx, containerId, container.StopOptions{Timeout: &config.ContainerStopTimeout}); err != nil {
			log.Warnf("failed to stop container gracefully, killing: %v", err)
			if err := client.ContainerKill(ctx, containerId, "SIGKILL"); err != nil {
				log.Warnf("failed to kill container: %v", err)
			}
		}
	}()

	log.Debug("waiting for container to stop")
	chanStopOk, chanStopErr := client.ContainerWait(ctx, containerId, container.WaitConditionNotRunning)
	select {
	case ok := <-chanStopOk:
		log.Debugf("container stopped with exit code: %d - error: %v", ok.StatusCode, ok.Error)
	case err := <-chanStopErr:
		log.Warnf("error while waiting for container to be stopped: %v", err)
	}
}

// ContainerRemove removes a Docker container instance.
func ContainerRemove(ctx context.Context, client *docker.Client, log logger.Logger, containerId string) {
	log.Debug("removing container")
	containerRemoveOptions := container.RemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}
	go func() {
		if err := client.ContainerRemove(ctx, containerId, containerRemoveOptions); err != nil {
			log.Warnf("failed to remove container: %v", err)
		}
	}()

	log.Debug("waiting for container to be removed")
	chanRemoveOk, chanRemoveErr := client.ContainerWait(ctx, containerId, container.WaitConditionRemoved)
	select {
	case ok := <-chanRemoveOk:
		log.Debugf("container removed with exit code: %d - reason: %v", ok.StatusCode, ok.Error)
	case err := <-chanRemoveErr:
		log.Warnf("error while waiting for container to be removed: %v", err)
	}
}

// FetchImageIdByTag fetchs a Docker image id by its tag name.
func FetchImageIdByTag(ctx context.Context, client *docker.Client, log logger.Logger, imageTag string) (*string, error) {
	images, err := client.ImageList(ctx, types.ImageListOptions{All: true})
	if err != nil {
		return nil, err
	}
	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == imageTag {
				return &image.ID, nil
			}
		}
	}
	return nil, fmt.Errorf("cannot find image: %s - reason: %v", imageTag, err)
}

// ImagePush pushes an image to the docker image registry.
func ImagePush(ctx context.Context, client *docker.Client, log logger.Logger, imageTag string) error {
	authConfig := registry.AuthConfig{
		Username: config.ImageRegistryUsername,
		Password: config.ImageRegistryPassword,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return err
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	response, err := client.ImagePush(ctx, imageTag, types.ImagePushOptions{
		All:          false,
		RegistryAuth: authStr,
	})
	if err != nil {
		return err
	}

	return processDockerOutput(log, response, dockerReaderStream())
}

// ImagePull pulls an image from the docker image registry.
func ImagePull(ctx context.Context, client *docker.Client, log logger.Logger, refStr string) error {
	authConfig := registry.AuthConfig{
		Username: config.ImageRegistryUsername,
		Password: config.ImageRegistryPassword,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return err
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	response, err := client.ImagePull(ctx, refStr, types.ImagePullOptions{
		All:          false,
		RegistryAuth: authStr,
	})
	if err != nil {
		return err
	}

	return processDockerOutput(log, response, dockerReaderStatus())
}

// ImageRemove removes an image from the Docker host.
func ImageRemove(ctx context.Context, client *docker.Client, log logger.Logger, imageTag string) error {
	log = log.WithFields(map[string]any{"image-tag": imageTag})
	imageId, err := FetchImageIdByTag(ctx, client, log, imageTag)
	if err != nil {
		return fmt.Errorf("failed to fetch image id by tag: %v", err)
	}
	responses, err := client.ImageRemove(ctx, *imageId, types.ImageRemoveOptions{Force: true})
	if err != nil {
		return fmt.Errorf("failed to remove image: %s - reason: %v", imageTag, err)
	}
	for _, response := range responses {
		log.Debugf("docker image removal status: %s - deleted: %s - untagged: %s", imageId, response.Deleted, response.Untagged)
	}
	return nil
}

// ImageBuild builds the rootfs image from a Dockerfile.
func ImageBuild(ctx context.Context, client *docker.Client, log logger.Logger, sourcePath string, dockerfilePath string, imageTag string) error {
	if !strings.HasSuffix(sourcePath, string(os.PathSeparator)) {
		sourcePath = fmt.Sprintf("%s%s", sourcePath, string(os.PathSeparator))
	}

	log = log.WithFields(map[string]any{"context-dir": sourcePath, "dockerfile": dockerfilePath, "image-tag": imageTag})

	buildContext, err := dockerArchive.TarWithOptions(sourcePath, &dockerArchive.TarOptions{})
	if err != nil {
		return fmt.Errorf("failed to create tar archive as Docker build context: %v", err)
	}
	defer buildContext.Close()

	buildResponse, err := client.ImageBuild(ctx, buildContext, types.ImageBuildOptions{
		Dockerfile:  dockerfilePath,
		Tags:        []string{imageTag},
		ForceRemove: true,
		Remove:      true,
		PullParent:  false, // TODO: This should be enabled in production to always use the latest version of the parent.
	})
	if err != nil {
		return fmt.Errorf("failed to build docker image: %v", err)
	}

	return processDockerOutput(log, buildResponse.Body, dockerReaderStream())
}

// ImageExport exports the rootfs from the container to the rootfs image file.
func ImageExport(ctx context.Context, client *docker.Client, log logger.Logger, destPath string, imageTag string, imageFileName string) error {
	cleanup := defers.NewDefers()
	defer cleanup.CallAll()

	imgFilePath := strings.Join([]string{config.ContainerDestMountTarget, imageFileName}, string(os.PathSeparator))

	log.Info("creating empty image ...")
	workerContainerId, err := createImage(ctx, client, log, destPath, imgFilePath)
	if err != nil {
		return fmt.Errorf("error while creating image file: %v", err)
	}
	log.Info("copying rootfs from container to image ...")
	if err := copyRootFsToImage(ctx, client, log, destPath, imgFilePath, imageTag); err != nil {
		return fmt.Errorf("error while copying rootfs to image file: %v", err)
	}
	log.Info("finalizing image ...")
	if err := finalizeImage(ctx, client, log, workerContainerId, imgFilePath); err != nil {
		return fmt.Errorf("error while finalizing image file: %v", err)
	}
	log.Info("resizing image to minimum size ...")
	if err := resizeImage(ctx, client, log, workerContainerId, imgFilePath); err != nil {
		return fmt.Errorf("error while resizing image file: %v", err)
	}
	cleanup.Add(func() {
		ContainerStop(ctx, client, log, *workerContainerId)
	})
	return nil
}

// ContainerExec executes a command in a given container and returns the output reader.
func ContainerExec(ctx context.Context, client *docker.Client, log logger.Logger, containerId string, execConfig types.ExecConfig) (*types.HijackedResponse, error) {
	createResponse, err := client.ContainerExecCreate(ctx, containerId, execConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create command exec: %v", err)
	}
	attachResponse, err := client.ContainerExecAttach(ctx, createResponse.ID, types.ExecStartCheck{})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to command exec: %v", err)
	}
	if err := client.ContainerExecStart(ctx, createResponse.ID, types.ExecStartCheck{}); err != nil {
		return nil, fmt.Errorf("failed to start command exec: %v", err)
	}
	return &attachResponse, nil
}

// ParseExecOutput parses the output of a command execution into separate lines.
func ParseExecOutput(reader *bufio.Reader) ([]string, error) {
	lines := []string{}
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return lines, err
		}
		lines = append(lines, strings.TrimSpace(line))
	}
	return lines, nil
}

// DebugOutput logs the lines of a reader as debug messages.
func DebugOutput(log logger.Logger, reader *bufio.Reader) {
	lines, err := ParseExecOutput(reader)
	if err != nil {
		log.Error("error while parsing reader")
	}
	for _, line := range lines {
		log.Debug(line)
	}
}

// ContainerCopy copies a file inside a container.
func ContainerCopy(ctx context.Context, client *docker.Client, log logger.Logger, containerId string, srcPath string, dstPath string) error {
	log.Debugf("copying %s to %s", srcPath, dstPath)
	execConfig := types.ExecConfig{
		Cmd:          []string{"cp", "-rP", srcPath, dstPath},
		AttachStdout: true,
		AttachStderr: true,
	}
	execResponse, err := ContainerExec(ctx, client, log, containerId, execConfig)
	if err != nil {
		return fmt.Errorf("error while executing copy to image command: %v", err)
	}
	defer execResponse.Close()

	DebugOutput(log, execResponse.Reader)

	return nil
}

// ContainerMount mounts an image file to an directory inside a container.
func ContainerMount(ctx context.Context, client *docker.Client, log logger.Logger, containerId *string, srcPath string, targetPath string) error {
	log.Debugf("mounting %s to %s", srcPath, targetPath)
	execConfig := types.ExecConfig{
		Cmd: []string{
			"/bin/sh", "-c",
			"mkdir -p " + targetPath + " && " +
				"mount " + srcPath + " " + targetPath,
		},
		AttachStdout: true,
		AttachStderr: true,
	}
	execResponse, err := ContainerExec(ctx, client, log, *containerId, execConfig)
	if err != nil {
		return fmt.Errorf("error while executing mount command: %v", err)
	}
	defer execResponse.Close()

	DebugOutput(log, execResponse.Reader)

	return nil
}

// ContainerUnmount unmounts a directory inside a container.
func ContainerUnmount(ctx context.Context, client *docker.Client, log logger.Logger, containerId *string, targetPath string) error {
	log.Debugf("unmounting %s", targetPath)
	execConfig := types.ExecConfig{
		Cmd:          []string{"umount", targetPath},
		AttachStdout: true,
		AttachStderr: true,
	}
	execResponse, err := ContainerExec(ctx, client, log, *containerId, execConfig)
	if err != nil {
		return fmt.Errorf("error while executing umount command: %v", err)
	}
	defer execResponse.Close()

	DebugOutput(log, execResponse.Reader)

	return nil
}

// createImage creates the empty rootfs image file.
func createImage(ctx context.Context, client *docker.Client, log logger.Logger, destPath string, imgFilePath string) (*string, error) {
	containerConfig := container.Config{
		OpenStdin: true,
		Tty:       true,
		Cmd:       strslice.StrSlice(config.ContainerCommand),
		Image:     config.BuilderContainerImageTag,
	}
	hostConfig := container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: destPath,
				Target: config.ContainerDestMountTarget,
			},
		},
		Privileged: true,
	}
	containerId, err := ContainerStart(ctx, client, log, nil, containerConfig, hostConfig, "", false)
	if err != nil {
		return nil, fmt.Errorf("error while starting builder docker container: %v", err)
	}

	log.Debug("creating empty rootfs image")
	imgExecConfig := types.ExecConfig{
		Cmd: []string{
			"/bin/sh", "-c",
			"dd if=/dev/zero " + strings.Join([]string{"of", imgFilePath}, "=") + " bs=1M count=1000 && " +
				"mkfs.ext4 " + imgFilePath,
		},
		AttachStdout: true,
		AttachStderr: true,
	}
	imgExecResponse, err := ContainerExec(ctx, client, log, *containerId, imgExecConfig)
	if err != nil {
		return containerId, fmt.Errorf("error while creating empty rootfs image: %v", err)
	}
	defer imgExecResponse.Close()

	DebugOutput(log, imgExecResponse.Reader)

	return containerId, nil
}

// copyRootFsToImage copys the rootfs from inside the container to the rootfs image file.
func copyRootFsToImage(ctx context.Context, client *docker.Client, log logger.Logger, destPath string, imgFilePath string, imageTag string) error {
	cleanup := defers.NewDefers()
	defer cleanup.CallAll()

	rootfsContainerConfig := container.Config{
		OpenStdin: true,
		Tty:       true,
		Cmd:       strslice.StrSlice(config.ContainerCommand),
		Image:     imageTag,
	}
	rootfsHostConfig := container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: destPath,
				Target: config.ContainerDestMountTarget,
			},
		},
		Privileged: true,
	}
	containerId, err := ContainerStart(ctx, client, log, cleanup, rootfsContainerConfig, rootfsHostConfig, "", true)
	if err != nil {
		return fmt.Errorf("error while starting export docker container: %v", err)
	}

	log.Debug("mounting rootfs image")
	if err := ContainerMount(ctx, client, log, containerId, imgFilePath, config.ContainerImageMountTarget); err != nil {
		return err
	}
	cleanup.Add(func() {
		log.Debug("unmounting rootfs image")
		if err := ContainerUnmount(ctx, client, log, containerId, config.ContainerImageMountTarget); err != nil {
			log.Errorf("error while unmounting rootfs image: %v", err)
			return
		}
	})

	log.Debug("discovering directories to copy")
	// The following command finds all directories and symlinks on the root level of the filesystem.
	findExecConfig := types.ExecConfig{
		Cmd:          []string{"find", "/", "-maxdepth", "1", "(", "-type", "d", "-o", "-type", "l", ")"},
		AttachStdout: true,
		AttachStderr: true,
	}
	findExecResponse, err := ContainerExec(ctx, client, log, *containerId, findExecConfig)
	if err != nil {
		return fmt.Errorf("error while executing find command: %v", err)
	}
	defer findExecResponse.Close()

	findExecLines, err := ParseExecOutput(findExecResponse.Reader)
	if err != nil {
		return fmt.Errorf("error while parsing find command exec output: %v", err)
	}

	// Iterate over the discovered filesystem directories and copy them to the rootfs image.
	log.Debug("copying directories and symlinks")
	for _, dir := range findExecLines {
		log.Debugf("handling directory or symlink: %s", dir)
		if slices.Contains([]string{config.ContainerDestMountTarget, config.ContainerImageMountTarget}, dir) {
			log.Debugf("directory is working directory: %s", dir)
			continue
		}
		if !strings.HasPrefix(dir, "/") || !utils.IsValidDirName(dir) {
			log.Debugf("directory is not a valid directory: %s", dir)
			continue
		}
		// Only create empty directory on the rootfs image.
		if slices.Contains(config.RootFsExcludeDirs, dir) {
			log.Debug("creating empty directory at destination")
			mkdirExecConfig := types.ExecConfig{
				Cmd:          []string{"mkdir", config.ContainerImageMountTarget + dir},
				AttachStdout: true,
				AttachStderr: true,
			}
			mkdirExecResponse, err := ContainerExec(ctx, client, log, *containerId, mkdirExecConfig)
			if err != nil {
				return fmt.Errorf("error while executing mkdir command: %v", err)
			}
			defer mkdirExecResponse.Close()
		} else {
			// Copy the whole directory to the rootfs image.
			if err := ContainerCopy(ctx, client, log, *containerId, dir, config.ContainerImageMountTarget+dir); err != nil {
				return fmt.Errorf("error while copying to image file: %v", err)
			}
		}
	}

	return nil
}

// finalizeImage finalizes the rootfs image by removing unnecessary files and applying the network config.
func finalizeImage(ctx context.Context, client *docker.Client, log logger.Logger, containerId *string, imgFilePath string) error {
	cleanup := defers.NewDefers()
	defer cleanup.CallAll()

	log.Debug("mounting rootfs image")
	if err := ContainerMount(ctx, client, log, containerId, imgFilePath, config.ContainerImageMountTarget); err != nil {
		return fmt.Errorf("error while mounting rootfs image: %v", err)
	}
	cleanup.Add(func() {
		log.Debug("unmounting rootfs image")
		if err := ContainerUnmount(ctx, client, log, containerId, config.ContainerImageMountTarget); err != nil {
			log.Errorf("error while unmounting rootfs image: %v", err)
			return
		}
	})

	log.Debug("removing unnecessary files")
	rmExecConfig := types.ExecConfig{
		Cmd: []string{
			"/bin/sh", "-c",
			"rm -rf " +
				config.ContainerImageMountTarget + "/var/lib/apt/lists/* " +
				config.ContainerImageMountTarget + "/var/cache/apt/archives/* " +
				config.ContainerImageMountTarget + "/var/tmp/* " +
				config.ContainerImageMountTarget + "/var/log/* " +
				config.ContainerImageMountTarget + "/usr/share/doc/* " +
				config.ContainerImageMountTarget + "/usr/share/man/* " +
				config.ContainerImageMountTarget + "/usr/share/info/* " +
				config.ContainerImageMountTarget + "/etc/apt/sources.list.d/* " +
				config.ContainerImageMountTarget + "/etc/apt/keyrings/*",
		},
		AttachStdout: true,
		AttachStderr: true,
	}
	rmExecResponse, err := ContainerExec(ctx, client, log, *containerId, rmExecConfig)
	if err != nil {
		return fmt.Errorf("error while removing unnecessary files: %v", err)
	}
	defer rmExecResponse.Close()

	DebugOutput(log, rmExecResponse.Reader)

	log.Debug("creating workspace directory")
	wrkExecConfig := types.ExecConfig{
		Cmd:          []string{"mkdir", config.ContainerImageMountTarget + "/workspace"},
		AttachStdout: true,
		AttachStderr: true,
	}
	wrkExecResponse, err := ContainerExec(ctx, client, log, *containerId, wrkExecConfig)
	if err != nil {
		return fmt.Errorf("error while creating workspace directory: %v", err)
	}
	defer wrkExecResponse.Close()

	DebugOutput(log, wrkExecResponse.Reader)

	log.Debug("applying network configuration")
	resExecConfig := types.ExecConfig{
		Cmd: []string{
			"/bin/sh", "-c",
			"rm " + config.ContainerImageMountTarget + "/etc/resolv.conf" + " && " +
				"echo \"nameserver 1.1.1.1\nnameserver 1.0.0.1\n\" > " + config.ContainerImageMountTarget + "/etc/resolv.conf",
		},
		AttachStdout: true,
		AttachStderr: true,
	}
	resExecResponse, err := ContainerExec(ctx, client, log, *containerId, resExecConfig)
	if err != nil {
		return fmt.Errorf("error while applying network configuration: %v", err)
	}
	defer resExecResponse.Close()

	DebugOutput(log, resExecResponse.Reader)

	return nil
}

// resizeImage resizes the rootfs image to its minimum size.
func resizeImage(ctx context.Context, client *docker.Client, log logger.Logger, containerId *string, imgFilePath string) error {
	log.Debug("resizing image file")
	execConfig := types.ExecConfig{
		Cmd: []string{
			"/bin/sh", "-c",
			"e2fsck -y -f " + imgFilePath + " && " +
				"resize2fs -M " + imgFilePath,
		},
		AttachStderr: true,
		AttachStdout: true,
	}
	execResponse, err := ContainerExec(ctx, client, log, *containerId, execConfig)
	if err != nil {
		return fmt.Errorf("error while executing image resize command: %v", err)
	}
	defer execResponse.Close()

	return nil
}
