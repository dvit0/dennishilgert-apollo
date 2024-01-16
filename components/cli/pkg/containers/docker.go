package containers

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
	ContainerDestMountTarget = "/dest"

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

// FetchImageIdByTag fetches a Docker image id by its tag name.
func FetchImageIdByTag(ctx context.Context, client *docker.Client, logger hclog.Logger, imageTag string) (string, error) {
	images, err := client.ImageList(ctx, types.ImageListOptions{All: true})
	if err != nil {
		return "", err
	}
	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == imageTag {
				return image.ID, nil
			}
		}
	}
	logger.Error("cannot find image", "tag", imageTag, "reason", err)
	return "", err
}

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
		PullParent:  true,
	})
	if err != nil {
		opLogger.Error("failed to build Docker image", "reason", err)
		return err
	}

	return processDockerOutput(opLogger, buildResponse.Body, dockerReaderStream())
}

func ImageExport(ctx context.Context, client *docker.Client, opLogger hclog.Logger, destPath string, imageTag string) error {
	cleanup := utils.NewDefers()
	defer cleanup.CallAll()

	// check if destination path is a directory and is writable.
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

func createImage(ctx context.Context, client *docker.Client, opLogger hclog.Logger, destPath string, imgFilePath string) (*string, error) {
	// create and start the builder container.
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

	// execute all commands on the builder container to create the empty rootfs ext4 image, create the mount dir and mount the image to that dir.
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
		opLogger.Error("error while executing the image creation commands")
		return containerId, err
	}
	defer imgExecResponse.Close()

	DebugOutput(opLogger, imgExecResponse.Reader)

	return containerId, nil
}

func copyRootFsToImage(ctx context.Context, client *docker.Client, opLogger hclog.Logger, destPath string, imgFilePath string, imageTag string) error {
	cleanup := utils.NewDefers()
	defer cleanup.CallAll()

	// create and start the export container.
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
	mntExecConfig := types.ExecConfig{
		Cmd: []string{
			"/bin/sh", "-c",
			"mkdir -p " + ContainerImageMountTarget + " && " +
				"mount " + imgFilePath + " " + ContainerImageMountTarget,
		},
		AttachStdout: true,
		AttachStderr: true,
	}
	mntExecResponse, err := ContainerExec(ctx, client, opLogger, *containerId, mntExecConfig)
	if err != nil {
		opLogger.Error("error while executing the image mount command")
		return err
	}
	defer mntExecResponse.Close()

	DebugOutput(opLogger, mntExecResponse.Reader)

	cleanup.Add(func() {
		// execute umount command on the container to unmount the rootfs ext4 image.
		opLogger.Debug("unmounting rootfs image")
		unmntExecConfig := types.ExecConfig{
			Cmd:          []string{"umount", ContainerDestMountTarget},
			AttachStdout: true,
			AttachStderr: true,
		}
		unmntExecResponse, err := ContainerExec(ctx, client, opLogger, *containerId, unmntExecConfig)
		if err != nil {
			opLogger.Error("error while executing unmount command")
			return
		}
		defer unmntExecResponse.Close()
	})

	// execute the find command on the container to get the filesystems root directories.
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
			opLogger.Debug("directory is working directory")
			continue
		}
		if !strings.HasPrefix(dir, "/") || !utils.IsValidDirName(dir) {
			opLogger.Debug("directory is not a valid directory")
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
		opLogger.Error("error while setting up resolv.conf")
		return err
	}
	defer resExecResponse.Close()

	DebugOutput(opLogger, resExecResponse.Reader)

	return nil
}

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

func DebugOutput(opLogger hclog.Logger, reader *bufio.Reader) {
	lines, err := ParseExecOutput(reader)
	if err != nil {
		opLogger.Error("error while parsing reader")
	}
	for _, line := range lines {
		opLogger.Debug(line)
	}
}

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
