package preparer

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/dennishilgert/apollo/internal/pkg/naming"
	"github.com/dennishilgert/apollo/pkg/container"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/proto/fleet/v1"
	"github.com/dennishilgert/apollo/pkg/utils"
)

var log = logger.NewLogger("apollo.manager.preparer")

type Options struct {
	DataPath             string
	ImageRegistryAddress string
}

type RunnerPreparer interface {
	PrepareDataDir() error
	InitializeFunction(ctx context.Context, request *fleet.InitializeFunctionRequest) error
}

type runnerPreparer struct {
	dataPath             string
	imageRegistryAddress string
}

func NewRunnerPreparer(opts Options) RunnerPreparer {
	return &runnerPreparer{
		dataPath:             opts.DataPath,
		imageRegistryAddress: opts.ImageRegistryAddress,
	}
}

func (r *runnerPreparer) PrepareDataDir() error {
	exists, fileInfo := utils.FileExists(r.dataPath)
	if !exists {
		if err := os.MkdirAll(r.dataPath, 0777); err != nil {
			return fmt.Errorf("failed to create data directory: %v", err)
		}
		return nil
	}
	ok, err := utils.IsDirAndWritable(r.dataPath, fileInfo)
	if !ok {
		return fmt.Errorf("data path is not a directory or not writable: %v", err)
	}
	return nil
}

func (r *runnerPreparer) InitializeFunction(ctx context.Context, request *fleet.InitializeFunctionRequest) error {
	path := strings.Join([]string{r.dataPath, request.FunctionUuid}, string(os.PathSeparator))
	log.Infof("initializing function at: %s", path)

	exists, fileInfo := utils.FileExists(path)
	if exists {
		ok, err := utils.IsDirAndWritable(path, fileInfo)
		if !ok {
			return fmt.Errorf("target path is not a directory or not writable: %v", err)
		}
		empty, err := utils.IsDirEmpty(path)
		if !empty {
			return fmt.Errorf("target directory is not empty")
		}
		if err != nil {
			return fmt.Errorf("failed to check if target directory is empty: %v", err)
		}
	} else {
		if err := os.MkdirAll(path, 0777); err != nil {
			return fmt.Errorf("failed to create target directory: %v", err)
		}
	}

	dockerClient, err := container.GetDefaultClient()
	if err != nil {
		return err
	}

	refStrRootFs := naming.ImageRefStr(r.imageRegistryAddress, naming.ImageNameRootFs(request.FunctionUuid), request.RootfsImageTag)
	refStrCode := naming.ImageRefStr(r.imageRegistryAddress, naming.ImageNameCode(request.FunctionUuid), request.CodeImageTag)

	log.Infof("pulling rootfs image: %s", refStrRootFs)
	if err := container.ImagePull(ctx, dockerClient, log, refStrRootFs); err != nil {
		log.Errorf("failed to pull rootfs image: %v", err)
		return err
	}

	log.Infof("exporting rootfs image: %s", refStrRootFs)
	if err := container.ImageExport(ctx, dockerClient, log, path, refStrRootFs, strings.Join([]string{request.FunctionUuid, "rootfs.ext4"}, "-")); err != nil {
		return err
	}

	log.Infof("pullung code image: %s", refStrCode)
	if err := container.ImagePull(ctx, dockerClient, log, refStrCode); err != nil {
		return err
	}

	log.Infof("exporting code image: %s", refStrCode)
	if err := container.ImageExport(ctx, dockerClient, log, path, refStrCode, strings.Join([]string{request.FunctionUuid, "code.ext4"}, "-")); err != nil {
		return err
	}

	return nil
}
