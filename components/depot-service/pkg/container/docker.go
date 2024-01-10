package containers

import (
	"apollo/depot-service/pkg/utils"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	docker "github.com/docker/docker/client"
	"github.com/hashicorp/go-hclog"
)

var (
	// ContainerStopTimeout specifies the time the container is given to shutdown gracefully.
	ContainerStopTimeout = 30

	// ContainerRootFsExportCommand specifies the comand that is run when image export container starts.
	ContainerRootFsExportCommand = []string{"/bin/sh"}

	// ContainerRootFsExecShell specifies the binary that is used to execute the export commands.
	ContainerRootFsExecShell = []string{"/bin/sh", "-c"}

	// ContainerRootFsCopyTimeout specifies the time the image export command is given to copy the rootfs.
	ContainerRootFsCopyTimeout = time.Duration(time.Second * 15)

	// ContainerRootFsExportMountTarget specifies the mount target to export the rootfs to.
	ContainerRootFsExportMountTarget = "/export-rootfs"

	// ContainerRootFsExportExcludeDirs specifies the directories whose content is excluded from export.
	ContainerRootFsExportExcludeDirs = []string{"/boot", "/opt", "/proc", "/run", "/srv", "/sys", "/tmp"}
)

// GetDefaultClient returns a default instance of the Docker client.
func GetDefaultClient() (*docker.Client, error) {
	return docker.NewClientWithOpts(docker.WithAPIVersionNegotiation())
}

// FetchImageIdByTag fetches a Docker image id by its tag name.
func FetchImageIdByTag(ctx context.Context, client *docker.Client, targetTag string) (string, error) {
	images, err := client.ImageList(ctx, types.ImageListOptions{All: true})
	if err != nil {
		return "", err
	}
	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == targetTag {
				return image.ID, nil
			}
		}
	}
	return "", fmt.Errorf("failed to find image with tag: %q", targetTag)
}

// ImageExport exports the rootfs of a Docker container.
func ImageExport(ctx context.Context, client *docker.Client, logger hclog.Logger, exportPath string, imageTag string) error {
	opLogger := logger.With("tag-name", imageTag)

	cleanup := utils.NewDefers()
	defer cleanup.CallAll()

	containerConfig := &container.Config{
		OpenStdin: true,
		Tty:       true,
		Cmd:       strslice.StrSlice(ContainerRootFsExportCommand),
		Image:     imageTag,
	}

	hostConfig := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: exportPath,
				Target: ContainerRootFsExportMountTarget,
			},
		},
	}

	opLogger.Debug("starting Docker container for rootfs export")

	containerCreateResponse, err := client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "temp-rootfs-export")
	if err != nil {
		opLogger.Error("failed to create Docker container", "reason", err)
		return err
	}

	cleanup.Add(func() {
		ContainerRemove(context.Background(), client, opLogger, containerCreateResponse.ID)
	})

	opLogger = opLogger.With("container-id", containerCreateResponse.ID)
	opLogger.Debug("container started")

	if err := client.ContainerStart(ctx, containerCreateResponse.ID, types.ContainerStartOptions{}); err != nil {
		opLogger.Error("failed to start Docker container", "reason", err)
		return err
	}

	cleanup.Add(func() {
		ContainerStop(context.Background(), client, opLogger, containerCreateResponse.ID)
	})

	bareList := strings.Join(ContainerRootFsExportExcludeDirs, " ")
	dirsNoCopy := "LIST=\"" + strings.Join([]string{"/", ContainerRootFsExportMountTarget}, " ") + " " + bareList + "\"; "
	dirsMkdirOnly := "LIST=\"" + bareList + "\"; "

	commands := []string{
		// Iterate over all directories and if the current directory is in the list, create an empty one in the mount target
		dirsMkdirOnly + "for d in $(find / -maxdepth 1 -type d); do if echo $LIST | grep -w $d > /dev/null; then mkdir " + ContainerRootFsExportMountTarget + "${d}; fi; done; exit 0",
		// Iterate over all directories and if the current directory is not in the list, copy it to the mount target
		dirsNoCopy + "for d in $(find / -maxdepth 1 -type d); do if echo $LIST | grep -v -w $d > /dev/null; then tar c \"$d\" | tar x -C " + ContainerRootFsExportMountTarget + "; fi; done",
		// Remove the mount target directory from the mount target
		fmt.Sprintf("rm -r %s%s", ContainerRootFsExportMountTarget, ContainerRootFsExportMountTarget),
	}

	for idx, command := range commands {
		opLogger.Debug(fmt.Sprintf("running command %d of %d", idx+1, len(commands)))

		execCreateResponse, err := client.ContainerExecCreate(ctx, containerCreateResponse.ID, types.ExecConfig{
			AttachStdout: true,
			AttachStderr: true,
			Cmd: func() []string {
				cmd := ContainerRootFsExecShell
				return append(cmd, command)
			}(),
		})
		if err != nil {
			opLogger.Error("failed to create command execution", "container-id", containerCreateResponse.ID, "reason", err)
			return err
		}

		execAttachResponse, err := client.ContainerExecAttach(ctx, execCreateResponse.ID, types.ExecStartCheck{
			Tty: true,
		})
		if err != nil {
			opLogger.Error("failed to attach to command execution", "container-id", containerCreateResponse.ID, "reason", err)
			return err
		}

		chanDone := make(chan struct{}, 1)
		chanError := make(chan error, 1)
		execReadCtx, execReadCtxCancelFunc := context.WithTimeout(ctx, ContainerRootFsCopyTimeout)
		defer execReadCtxCancelFunc()

		go func() {
			defer execAttachResponse.Close()

			for {
				bs, err := execAttachResponse.Reader.ReadBytes('\n')
				if execReadCtx.Err() != nil {
					return
				}
				if err != nil {
					if err == io.EOF {
						close(chanDone) // finished reading successfully
						return
					}
					chanError <- err
					return
				}
				opLogger.Debug("command exec attach output", strings.TrimSpace(string(bs)))
			}
		}()

		select {
		case <-chanDone:
			opLogger.Debug(fmt.Sprintf("command exec %d of %d finished successfully", idx+1, len(commands)))
			close(chanError)
		case err := <-chanError:
			opLogger.Error(fmt.Sprintf("command exec %d of %d finished with an error", idx+1, len(commands)), "reason", err)
			close(chanDone)
			return err
		case <-execReadCtx.Done():
			close(chanDone)
			close(chanError)
			if execReadCtx.Err() != nil {
				opLogger.Error(fmt.Sprintf("command execution %d of %d finished with an context error", idx+1, len(commands)), "reason", execReadCtx.Err())
				return execReadCtx.Err()
			}
		}
	}

	return nil
}

// ImagePull pulls a specified Docker image from the image registry.
func ImagePull(ctx context.Context, client *docker.Client, logger hclog.Logger, imageName string) error {
	response, err := client.ImagePull(ctx, imageName, types.ImagePullOptions{All: false})
	if err != nil {
		return err
	}
	if err := ProcessDockerOutput(logger.Named("image-pull"), response, DockerReaderStatus()); err != nil {
		return err
	}

	return nil
}

// ImageRemove removes the local copy of a Docker image.
func ImageRemove(ctx context.Context, client *docker.Client, logger hclog.Logger, imageTag string) error {
	opLogger := logger.With("image-tag", imageTag)
	opLogger.Debug("removing Docker image")

	imageId, err := FetchImageIdByTag(ctx, client, imageTag)
	if err != nil {
		opLogger.Error("failed to fetch image id by tag", "reason", err)
		return err
	}
	responses, err := client.ImageRemove(ctx, imageId, types.ImageRemoveOptions{Force: true})
	if err != nil {
		opLogger.Error("failed to remove Docker image", "image-id", imageId, "reason", err)
		return err
	}
	for _, response := range responses {
		opLogger.Debug("Docker image removal status", "image-id", imageId, "deleted", response.Deleted, "untagged", response.Untagged)
	}

	return nil
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
