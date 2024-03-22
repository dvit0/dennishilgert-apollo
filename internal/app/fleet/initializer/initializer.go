package initializer

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/dennishilgert/apollo/internal/pkg/naming"
	"github.com/dennishilgert/apollo/pkg/container"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/proto/fleet/v1"
	"github.com/dennishilgert/apollo/pkg/storage"
	"github.com/dennishilgert/apollo/pkg/utils"
)

var log = logger.NewLogger("apollo.manager.initializer")

type Options struct {
	DataPath             string
	ImageRegistryAddress string
}

type RunnerInitializer interface {
	InitializeDataDir() error
	InitializeFunction(ctx context.Context, request *fleet.PrepareRunnerRequest) error
}

type runnerInitializer struct {
	storageService       storage.StorageService
	dataPath             string
	imageRegistryAddress string
}

func NewRunnerInitializer(storageService storage.StorageService, opts Options) RunnerInitializer {
	return &runnerInitializer{
		storageService:       storageService,
		dataPath:             opts.DataPath,
		imageRegistryAddress: opts.ImageRegistryAddress,
	}
}

func (r *runnerInitializer) InitializeDataDir() error {
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

func (r *runnerInitializer) InitializeFunction(ctx context.Context, request *fleet.PrepareRunnerRequest) error {
	path := strings.Join([]string{r.dataPath, "functions", request.FunctionUuid}, string(os.PathSeparator))
	filename := strings.Join([]string{request.FunctionUuid, "ext4"}, ".")

	log.Debugf("check if function is already initialized")
	exists, _ := utils.FileExists(strings.Join([]string{path, filename}, string(os.PathSeparator)))
	if exists {
		log.Infof("function is already initialized: %s", request.FunctionUuid)
		return nil
	}
	if err := r.initializeKernel(ctx, request.KernelName, request.KernelVersion); err != nil {
		return err
	}
	if err := r.initializeRuntime(ctx, request.RuntimeName, request.RuntimeVersion); err != nil {
		return err
	}

	log.Infof("initializing function: %s", request.FunctionUuid)
	if err := prepareTargetDirectory(path); err != nil {
		return err
	}
	dockerClient, err := container.GetDefaultClient()
	if err != nil {
		return err
	}
	refString := naming.ImageRefStr(r.imageRegistryAddress, request.FunctionUuid, "latest")

	log.Infof("pulling function image: %s", refString)
	if err := container.ImagePull(ctx, dockerClient, log, refString); err != nil {
		log.Errorf("failed to pull function image: %v", err)
		return err
	}
	log.Infof("exporting function image: %s", refString)
	if err := container.ImageExport(ctx, dockerClient, log, path, refString, filename); err != nil {
		return err
	}

	return nil
}

func (r *runnerInitializer) initializeKernel(ctx context.Context, kernelName string, kernelVersion string) error {
	kernel := strings.Join([]string{kernelName, kernelVersion}, "-")
	path := strings.Join([]string{r.dataPath, "kernels", kernel}, string(os.PathSeparator))

	log.Debugf("check if kernel is already initialized")
	exists, _ := utils.FileExists(strings.Join([]string{path, kernel}, string(os.PathSeparator)))
	if exists {
		log.Infof("kernel is already initialized: %s", kernel)
		return nil
	}

	log.Infof("initializing kernel: %s", kernel)
	if err := prepareTargetDirectory(path); err != nil {
		return err
	}
	if err := r.storageService.DownloadObject(ctx, naming.StorageKernelBucketName, kernel, strings.Join([]string{path, kernel}, string(os.PathSeparator))); err != nil {
		return err
	}

	return nil
}

func (r *runnerInitializer) initializeRuntime(ctx context.Context, runtimeName string, runtimeVersion string) error {
	runtime := strings.Join([]string{runtimeName, runtimeVersion}, "-")
	path := strings.Join([]string{r.dataPath, "runtimes", runtime}, string(os.PathSeparator))
	filename := strings.Join([]string{runtime, "ext4"}, ".")

	log.Debugf("check if runtime is already initialized")
	exists, _ := utils.FileExists(strings.Join([]string{path, filename}, string(os.PathSeparator)))
	if exists {
		log.Infof("runtime is already initialized: %s", runtime)
		return nil
	}

	log.Infof("initializing runtime: %s", runtime)
	if err := prepareTargetDirectory(path); err != nil {
		return err
	}
	dockerClient, err := container.GetDefaultClient()
	if err != nil {
		return err
	}
	refString := naming.ImageRefStr(r.imageRegistryAddress, runtimeName, runtimeVersion)

	log.Infof("pulling runtime image: %s", refString)
	if err := container.ImagePull(ctx, dockerClient, log, refString); err != nil {
		log.Errorf("failed to pull runtime image: %v", err)
		return err
	}
	log.Infof("exporting runtime image: %s", refString)
	if err := container.ImageExport(ctx, dockerClient, log, path, refString, strings.Join([]string{runtime, "ext4"}, ".")); err != nil {
		return err
	}

	return nil
}

func prepareTargetDirectory(path string) error {
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
	return nil
}
