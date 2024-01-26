package container

import (
	"apollo/cli/pkg/utils"
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	docker "github.com/docker/docker/client"
	dockerArchive "github.com/docker/docker/pkg/archive"
	"github.com/hashicorp/go-hclog"
)

var (
	// Name of the rootfs image file.
	RootFsImageFileName = "rootfs.ext4"

	// Specifies the directories whose content is excluded from export.
	RootFsExcludeDirs = []string{"/boot", "/opt", "/proc", "/run", "/srv", "/sys", "/tmp"}

	// Specifies the Docker container tag to use for the worker container.
	BuilderContainerImageTag = "debian:stable"

	// Specifies the mount target for the destination directory on the host.
	ContainerDestMountTarget = "/dist"

	// Specifies the mount target for the rootfs image.
	ContainerImageMountTarget = "/rootfs-image"

	// Specifies the time the container is given to shutdown gracefully.
	ContainerStopTimeout = 5

	// Specifies the comand that is run when image export container starts.
	ContainerCommand = []string{"/bin/sh"}

	// Specifies the time the image export command is given to copy the rootfs.
	RootFsCopyTimeout = time.Duration(time.Second * 15)
)

// GetDefaultClient returns a default instance of the Docker client.
func GetDefaultClient() (*docker.Client, error) {
	return docker.NewClientWithOpts(docker.WithAPIVersionNegotiation())
}

// FetchImageIdByTag fetchs a Docker image id by its tag name.
func FetchImageIdByTag(ctx context.Context, client *docker.Client, logger hclog.Logger, imageTag string) (*string, error) {
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
	logger.Error("cannot find image", "tag", imageTag, "reason", err)
	return nil, err
}

// ImagePull pulls an image from the docker image registry.
func ImagePull(ctx context.Context, client *docker.Client, opLogger hclog.Logger, refStr string) error {
	response, err := client.ImagePull(ctx, refStr, types.ImagePullOptions{All: false})
	if err != nil {
		return err
	}
	if err := processDockerOutput(opLogger, response, dockerReaderStatus()); err != nil {
		return err
	}
	return nil
}

// ImagePush pushes an image to the docker image registry.
func ImagePush(ctx context.Context, client *docker.Client, opLogger hclog.Logger, imageTag string) error {
	response, err := client.ImagePush(ctx, imageTag, types.ImagePushOptions{All: false})
	if err != nil {
		return err
	}
	if err := processDockerOutput(opLogger, response, dockerReaderStream()); err != nil {
		return err
	}
	return nil
}

// ImageRemove removes an image from the Docker host.
func ImageRemove(ctx context.Context, client *docker.Client, opLogger hclog.Logger, imageTag string) error {
	opLogger = opLogger.With("image-tag", imageTag)
	imageId, err := FetchImageIdByTag(ctx, client, opLogger, imageTag)
	if err != nil {
		opLogger.Error("failed to fetch image id by tag", "reason", err)
		return err
	}
	responses, err := client.ImageRemove(ctx, *imageId, types.ImageRemoveOptions{Force: true})
	if err != nil {
		opLogger.Error("failed to remove image", "reason", err)
		return err
	}
	for _, response := range responses {
		opLogger.Debug("docker image removal status", "image-id", imageId, "deleted", response.Deleted, "untagged", response.Untagged)
	}
	return nil
}

// ImageBuild builds the rootfs image from a Dockerfile.
func ImageBuild(ctx context.Context, client *docker.Client, logger hclog.Logger, sourcePath string, dockerfilePath string, imageTag string) error {
	if !strings.HasSuffix(sourcePath, string(os.PathSeparator)) {
		sourcePath = fmt.Sprintf("%s%s", sourcePath, string(os.PathSeparator))
	}

	opLogger := logger.With("dir-context", sourcePath, "dockerfile", dockerfilePath, "image-tag", imageTag)

	buildContext, err := dockerArchive.TarWithOptions(sourcePath, &dockerArchive.TarOptions{})
	if err != nil {
		opLogger.Error("failed to create tar archive as Docker build context", "reason", err)
		return err
	}
	defer buildContext.Close()

	buildResponse, err := client.ImageBuild(ctx, buildContext, types.ImageBuildOptions{
		Dockerfile:  dockerfilePath,
		Tags:        []string{imageTag},
		ForceRemove: true,
		Remove:      true,
		PullParent:  false, // this should be enabled in production to always use the latest version of the parent
	})
	if err != nil {
		opLogger.Error("failed to build Docker image", "reason", err)
		return err
	}

	return processDockerOutput(opLogger, buildResponse.Body, dockerReaderStream())
}

// exports rootfs from the container to the rootfs image file.
func ImageExport(ctx context.Context, client *docker.Client, opLogger hclog.Logger, destPath string, imageTag string) error {
	cleanup := utils.NewDefers()
	defer cleanup.CallAll()

	// check if destination path is a directory and is writable.
	opLogger.Debug("checking destination path", "dest-path", destPath)
	err := utils.IsDirAndWritable(destPath)
	if err != nil {
		opLogger.Error("error while checking destination path", "reason", err)
		return err
	}

	imgFilePath := strings.Join([]string{ContainerDestMountTarget, RootFsImageFileName}, string(os.PathSeparator))

	opLogger.Info("creating empty image ...")
	workerContainerId, err := createImage(ctx, client, opLogger, destPath, imgFilePath)
	if err != nil {
		opLogger.Error("error while creating image file")
		return err
	}

	opLogger.Info("copying rootfs from container to image ...")
	if err := copyRootFsToImage(ctx, client, opLogger, destPath, imgFilePath, imageTag); err != nil {
		opLogger.Error("error while copying rootfs to image file")
		return err
	}

	opLogger.Info("finalizing image ...")
	if err := finalizeImage(ctx, client, opLogger, workerContainerId, imgFilePath); err != nil {
		opLogger.Error("error while finalizing image file")
		return err
	}

	opLogger.Info("resizing image to minimum size ...")
	if err := resizeImage(ctx, client, opLogger, workerContainerId, imgFilePath); err != nil {
		opLogger.Error("error while resizing image file")
		return err
	}

	cleanup.Add(func() {
		ContainerStop(ctx, client, opLogger, *workerContainerId)
	})

	return nil
}

// createImage creates the empty rootfs image file.
func createImage(ctx context.Context, client *docker.Client, opLogger hclog.Logger, destPath string, imgFilePath string) (*string, error) {
	containerConfig := container.Config{
		OpenStdin: true,
		Tty:       true,
		Cmd:       strslice.StrSlice(ContainerCommand),
		Image:     BuilderContainerImageTag,
	}
	hostConfig := container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: destPath,
				Target: ContainerDestMountTarget,
			},
		},
		Privileged: true,
	}
	containerId, err := ContainerStart(ctx, client, opLogger, nil, containerConfig, hostConfig, "", false)
	if err != nil {
		opLogger.Error("error while starting builder Docker container")
		return nil, err
	}

	opLogger.Debug("creating empty rootfs image")
	imgExecConfig := types.ExecConfig{
		Cmd: []string{
			"/bin/sh", "-c",
			"dd if=/dev/zero " + strings.Join([]string{"of", imgFilePath}, "=") + " bs=1M count=1000 && " +
				"mkfs.ext4 " + imgFilePath,
		},
		AttachStdout: true,
		AttachStderr: true,
	}
	imgExecResponse, err := ContainerExec(ctx, client, opLogger, *containerId, imgExecConfig)
	if err != nil {
		opLogger.Error("error while creating empty rootfs image")
		return containerId, err
	}
	defer imgExecResponse.Close()

	DebugOutput(opLogger, imgExecResponse.Reader)

	return containerId, nil
}

// copyRootFsToImage copys the rootfs from inside the container to the rootfs image file.
func copyRootFsToImage(ctx context.Context, client *docker.Client, opLogger hclog.Logger, destPath string, imgFilePath string, imageTag string) error {
	cleanup := utils.NewDefers()
	defer cleanup.CallAll()

	rootfsContainerConfig := container.Config{
		OpenStdin: true,
		Tty:       true,
		Cmd:       strslice.StrSlice(ContainerCommand),
		Image:     imageTag,
	}
	rootfsHostConfig := container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: destPath,
				Target: ContainerDestMountTarget,
			},
		},
		Privileged: true,
	}
	containerId, err := ContainerStart(ctx, client, opLogger, cleanup, rootfsContainerConfig, rootfsHostConfig, "", true)
	if err != nil {
		opLogger.Error("error while starting export Docker container")
		return err
	}

	opLogger.Debug("mounting rootfs image")
	if err := ContainerMount(ctx, client, opLogger, containerId, imgFilePath, ContainerImageMountTarget); err != nil {
		opLogger.Error("error while mounting rootfs image")
		return err
	}
	cleanup.Add(func() {
		opLogger.Debug("unmounting rootfs image")
		if err := ContainerUnmount(ctx, client, opLogger, containerId, ContainerImageMountTarget); err != nil {
			opLogger.Error("error while unmounting rootfs image")
			return
		}
	})

	opLogger.Debug("discovering directories to copy")
	findExecConfig := types.ExecConfig{
		Cmd:          []string{"find", "/", "-maxdepth", "1", "-type", "d"},
		AttachStdout: true,
		AttachStderr: true,
	}
	findExecResponse, err := ContainerExec(ctx, client, opLogger, *containerId, findExecConfig)
	if err != nil {
		opLogger.Error("error while executing find command")
		return err
	}
	defer findExecResponse.Close()

	findExecLines, err := ParseExecOutput(findExecResponse.Reader)
	if err != nil {
		opLogger.Error("error while parsing find command exec output", "reason", err)
		return err
	}

	// iterate over the discovered filesystem directories and copy them to the rootfs image.
	opLogger.Debug("copying directories")
	for _, dir := range findExecLines {
		opLogger.Debug("handling directory", "dir", dir)
		if slices.Contains([]string{ContainerDestMountTarget, ContainerImageMountTarget}, dir) {
			opLogger.Debug("directory is working directory", "dir", dir)
			continue
		}
		if !strings.HasPrefix(dir, "/") || !utils.IsValidDirName(dir) {
			opLogger.Debug("directory is not a valid directory", "dir", dir)
			continue
		}
		// only create empty directory on the rootfs image.
		if slices.Contains(RootFsExcludeDirs, dir) {
			// only create empty directory on the rootfs image.
			opLogger.Debug("creating empty directory at destination")
			mkdirExecConfig := types.ExecConfig{
				Cmd:          []string{"mkdir", ContainerImageMountTarget + dir},
				AttachStdout: true,
				AttachStderr: true,
			}
			mkdirExecResponse, err := ContainerExec(ctx, client, opLogger, *containerId, mkdirExecConfig)
			if err != nil {
				opLogger.Error("error while executing mkdir command")
				return err
			}
			defer mkdirExecResponse.Close()
		} else {
			// copy the whole directory to the rootfs image.
			if err := ContainerCopy(ctx, client, opLogger, *containerId, dir, ContainerImageMountTarget+dir); err != nil {
				opLogger.Error("error while copying to image file")
				return err
			}
		}
	}

	return nil
}

// finalizeImage finalizes the rootfs image by removing unnecessary files and applying the network config.
func finalizeImage(ctx context.Context, client *docker.Client, opLogger hclog.Logger, containerId *string, imgFilePath string) error {
	cleanup := utils.NewDefers()
	defer cleanup.CallAll()

	opLogger.Debug("mounting rootfs image")
	if err := ContainerMount(ctx, client, opLogger, containerId, imgFilePath, ContainerImageMountTarget); err != nil {
		opLogger.Error("error while mounting rootfs image")
		return err
	}
	cleanup.Add(func() {
		opLogger.Debug("unmounting rootfs image")
		if err := ContainerUnmount(ctx, client, opLogger, containerId, ContainerImageMountTarget); err != nil {
			opLogger.Error("error while unmounting rootfs image")
			return
		}
	})

	opLogger.Debug("removing unnecessary files")
	rmExecConfig := types.ExecConfig{
		Cmd: []string{
			"/bin/sh", "-c",
			"rm -rf " +
				ContainerImageMountTarget + "/var/lib/apt/lists/* " +
				ContainerImageMountTarget + "/var/cache/apt/archives/* " +
				ContainerImageMountTarget + "/var/tmp/* " +
				ContainerImageMountTarget + "/var/log/* " +
				ContainerImageMountTarget + "/usr/share/doc/* " +
				ContainerImageMountTarget + "/usr/share/man/* " +
				ContainerImageMountTarget + "/usr/share/info/* " +
				ContainerImageMountTarget + "/etc/apt/sources.list.d/* " +
				ContainerImageMountTarget + "/etc/apt/keyrings/*",
		},
		AttachStdout: true,
		AttachStderr: true,
	}
	rmExecResponse, err := ContainerExec(ctx, client, opLogger, *containerId, rmExecConfig)
	if err != nil {
		opLogger.Error("error while removing unnecessary files")
		return err
	}
	defer rmExecResponse.Close()

	DebugOutput(opLogger, rmExecResponse.Reader)

	opLogger.Debug("creating workspace directory")
	wrkExecConfig := types.ExecConfig{
		Cmd:          []string{"mkdir", ContainerImageMountTarget + "/workspace"},
		AttachStdout: true,
		AttachStderr: true,
	}
	wrkExecResponse, err := ContainerExec(ctx, client, opLogger, *containerId, wrkExecConfig)
	if err != nil {
		opLogger.Error("error while creating workspace directory")
		return err
	}
	defer wrkExecResponse.Close()

	DebugOutput(opLogger, wrkExecResponse.Reader)

	opLogger.Debug("applying network configuration")
	resExecConfig := types.ExecConfig{
		Cmd: []string{
			"/bin/sh", "-c",
			"rm " + ContainerImageMountTarget + "/etc/resolv.conf" + " && " +
				"echo \"nameserver 1.1.1.1\nnameserver 1.0.0.1\n\" > " + ContainerImageMountTarget + "/etc/resolv.conf",
		},
		AttachStdout: true,
		AttachStderr: true,
	}
	resExecResponse, err := ContainerExec(ctx, client, opLogger, *containerId, resExecConfig)
	if err != nil {
		opLogger.Error("error while applying network configuration")
		return err
	}
	defer resExecResponse.Close()

	DebugOutput(opLogger, resExecResponse.Reader)

	return nil
}

// resizeImage resizes the rootfs image to its minimum size.
func resizeImage(ctx context.Context, client *docker.Client, opLogger hclog.Logger, containerId *string, imgFilePath string) error {
	opLogger.Debug("resizing image file")
	execConfig := types.ExecConfig{
		Cmd: []string{
			"/bin/sh", "-c",
			"e2fsck -y -f " + imgFilePath + " && " +
				"resize2fs -M " + imgFilePath,
		},
		AttachStderr: true,
		AttachStdout: true,
	}
	execResponse, err := ContainerExec(ctx, client, opLogger, *containerId, execConfig)
	if err != nil {
		opLogger.Error("error while executing image resize command")
		return err
	}
	defer execResponse.Close()

	return nil
}

// ContainerExec executes a command in a given container and returns the output reader.
func ContainerExec(ctx context.Context, client *docker.Client, opLogger hclog.Logger, containerId string, execConfig types.ExecConfig) (*types.HijackedResponse, error) {
	createResponse, err := client.ContainerExecCreate(ctx, containerId, execConfig)
	if err != nil {
		opLogger.Error("failed to create command exec", "reason", err)
		return nil, err
	}

	attachResponse, err := client.ContainerExecAttach(ctx, createResponse.ID, types.ExecStartCheck{})
	if err != nil {
		opLogger.Error("failed to attach to command exec", "reason", err)
		return nil, err
	}

	if err := client.ContainerExecStart(ctx, createResponse.ID, types.ExecStartCheck{}); err != nil {
		opLogger.Error("failed to start command exec", "reason", err)
		return nil, err
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
func DebugOutput(opLogger hclog.Logger, reader *bufio.Reader) {
	lines, err := ParseExecOutput(reader)
	if err != nil {
		opLogger.Error("error while parsing reader")
	}
	for _, line := range lines {
		opLogger.Debug(line)
	}
}

// ContainerCopy copies a file inside a container.
func ContainerCopy(ctx context.Context, client *docker.Client, opLogger hclog.Logger, containerId string, srcPath string, dstPath string) error {
	opLogger.Debug("copying", "src", srcPath, "dst", dstPath)
	execConfig := types.ExecConfig{
		Cmd:          []string{"cp", "-r", srcPath, dstPath},
		AttachStdout: true,
		AttachStderr: true,
	}
	execResponse, err := ContainerExec(ctx, client, opLogger, containerId, execConfig)
	if err != nil {
		opLogger.Error("error while executing copy to image command")
		return err
	}
	defer execResponse.Close()

	DebugOutput(opLogger, execResponse.Reader)

	return nil
}

// ContainerMount mounts an image file to an directory inside a container.
func ContainerMount(ctx context.Context, client *docker.Client, opLogger hclog.Logger, containerId *string, srcPath string, targetPath string) error {
	opLogger.Debug("mounting", "src-path", srcPath, "target-path", targetPath)
	execConfig := types.ExecConfig{
		Cmd: []string{
			"/bin/sh", "-c",
			"mkdir -p " + targetPath + " && " +
				"mount " + srcPath + " " + targetPath,
		},
		AttachStdout: true,
		AttachStderr: true,
	}
	execResponse, err := ContainerExec(ctx, client, opLogger, *containerId, execConfig)
	if err != nil {
		opLogger.Error("error while executing mount command")
		return err
	}
	defer execResponse.Close()

	DebugOutput(opLogger, execResponse.Reader)

	return nil
}

// ContainerUnmount unmounts a directory inside a container.
func ContainerUnmount(ctx context.Context, client *docker.Client, opLogger hclog.Logger, containerId *string, targetPath string) error {
	opLogger.Debug("unmounting", "target-path", targetPath)
	execConfig := types.ExecConfig{
		Cmd:          []string{"umount", targetPath},
		AttachStdout: true,
		AttachStderr: true,
	}
	execResponse, err := ContainerExec(ctx, client, opLogger, *containerId, execConfig)
	if err != nil {
		opLogger.Error("error while executing umount command")
		return err
	}
	defer execResponse.Close()

	DebugOutput(opLogger, execResponse.Reader)

	return nil
}

// ContainerStart creates and starts a container and adds its stop and removal to the cleanup defers.
func ContainerStart(ctx context.Context, client *docker.Client, opLogger hclog.Logger, cleanup utils.Defers, containerConfig container.Config, hostConfig container.HostConfig, name string, destroyAfter bool) (*string, error) {
	opLogger.Debug("creating container", "image-tag", containerConfig.Image)
	containerCreateResponse, err := client.ContainerCreate(ctx, &containerConfig, &hostConfig, nil, nil, name)
	if err != nil {
		opLogger.Error("failed to create Docker container", "reason", err)
		return nil, err
	}

	opLogger = opLogger.With("container-id", containerCreateResponse.ID[:12])

	opLogger.Debug("starting container")
	if err := client.ContainerStart(ctx, containerCreateResponse.ID, types.ContainerStartOptions{}); err != nil {
		opLogger.Error("failed to start Docker container", "reason", err)
		return nil, err
	}

	if destroyAfter {
		cleanup.Add(func() {
			ContainerRemove(context.Background(), client, opLogger, containerCreateResponse.ID)
		})
		cleanup.Add(func() {
			ContainerStop(context.Background(), client, opLogger, containerCreateResponse.ID)
		})
	}

	return &containerCreateResponse.ID, nil
}

// ContainerStop stops a container gracefully or kills it after a timeout.
func ContainerStop(ctx context.Context, client *docker.Client, opLogger hclog.Logger, containerId string) {
	opLogger.Debug("stopping container")
	go func() {
		if err := client.ContainerStop(ctx, containerId, container.StopOptions{Timeout: &ContainerStopTimeout}); err != nil {
			opLogger.Warn("failed to stop container gracefully, killing", "reason", err)
			if err := client.ContainerKill(ctx, containerId, "SIGKILL"); err != nil {
				opLogger.Warn("failed to kill container", "reason", err)
			}
		}
	}()

	opLogger.Debug("waiting for container to stop")
	chanStopOk, chanStopErr := client.ContainerWait(ctx, containerId, container.WaitConditionNotRunning)
	select {
	case ok := <-chanStopOk:
		opLogger.Debug("container stopped", "exit-code", ok.StatusCode, "error-message", ok.Error)
	case err := <-chanStopErr:
		opLogger.Warn("error while waiting for container to be stopped", "reason", err)
	}
}

// ContainerRemove removes a Docker container instance.
func ContainerRemove(ctx context.Context, client *docker.Client, opLogger hclog.Logger, containerId string) {
	opLogger.Debug("removing container")
	containerRemoveOptions := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}
	go func() {
		if err := client.ContainerRemove(ctx, containerId, containerRemoveOptions); err != nil {
			opLogger.Warn("failed to remove container", "reason", err)
		}
	}()

	opLogger.Debug("waiting for container to be removed")
	chanRemoveOk, chanRemoveErr := client.ContainerWait(ctx, containerId, container.WaitConditionRemoved)
	select {
	case ok := <-chanRemoveOk:
		opLogger.Debug("container removed", "exit-code", ok.StatusCode, "error-message", ok.Error)
	case err := <-chanRemoveErr:
		opLogger.Warn("error while waiting for container to be removed", "reason", err)
	}
}
