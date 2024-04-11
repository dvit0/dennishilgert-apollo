package initializer

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/dennishilgert/apollo/internal/app/fleet/operator/runner"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
	"github.com/dennishilgert/apollo/pkg/container"
	"github.com/dennishilgert/apollo/pkg/logger"
	fleetpb "github.com/dennishilgert/apollo/pkg/proto/fleet/v1"
	"github.com/dennishilgert/apollo/pkg/storage"
	"github.com/dennishilgert/apollo/pkg/utils"
)

var log = logger.NewLogger("apollo.manager.initializer")

type Options struct {
	DataPath             string
	ImageRegistryAddress string
}

type RunnerInitializer interface {
	DataPath() string
	InitializeDataDir() error
	InitializeRunner(ctx context.Context, cfg *runner.Config) error
	RemoveRunner(ctx context.Context, runnerUuid string) error
	InitializeFunction(ctx context.Context, request *fleetpb.InitializeFunctionRequest) error
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

func (r *runnerInitializer) DataPath() string {
	return r.dataPath
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

func (r *runnerInitializer) InitializeRunner(ctx context.Context, cfg *runner.Config) error {
	path := naming.RunnerStoragePath(r.dataPath, cfg.RunnerUuid)

	log.Debugf("initializing runner: %s", cfg.RunnerUuid)
	if err := prepareTargetDirectory(path); err != nil {
		return err
	}

	_, err := os.Create(cfg.LogFilePath)
	if err != nil {
		return err
	}
	_, err = os.Create(cfg.StdOutFilePath)
	if err != nil {
		return err
	}
	_, err = os.Create(cfg.StdErrFilePath)
	if err != nil {
		return err
	}

	return nil
}

func (r *runnerInitializer) RemoveRunner(ctx context.Context, runnerUuid string) error {
	path := naming.RunnerStoragePath(r.dataPath, runnerUuid)

	if err := os.RemoveAll(path); err != nil {
		return err
	}

	return nil
}

func (r *runnerInitializer) InitializeFunction(ctx context.Context, request *fleetpb.InitializeFunctionRequest) error {
	path := naming.FunctionStoragePath(r.dataPath, request.FunctionUuid)
	filename := naming.FunctionImageFileName(request.FunctionUuid)

	log.Debugf("check if function dependencies are already initialized")
	if err := r.initializeKernel(ctx, request.Kernel.Name, request.Kernel.Version); err != nil {
		return err
	}
	if err := r.initializeRuntime(ctx, request.Runtime.Name, request.Runtime.Version); err != nil {
		return err
	}
	log.Debugf("check if function is already initialized")
	exists, _ := utils.FileExists(strings.Join([]string{path, filename}, string(os.PathSeparator)))
	if exists {
		log.Infof("function is already initialized: %s", request.FunctionUuid)
		return nil
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
	path := naming.KernelStoragePath(r.dataPath, kernelName, kernelVersion)
	filename := naming.KernelFileName(kernelName, kernelVersion)

	log.Debugf("check if kernel is already initialized")
	exists, _ := utils.FileExists(strings.Join([]string{path, filename}, string(os.PathSeparator)))
	if exists {
		log.Infof("kernel is already initialized: %s %s", kernelName, kernelVersion)
		return nil
	}

	log.Infof("initializing kernel: %s %s", kernelName, kernelVersion)
	if err := prepareTargetDirectory(path); err != nil {
		return err
	}
	log.Debugf("downloading kernel image: %s %s", kernelName, kernelVersion)
	if err := r.storageService.DownloadObject(ctx, naming.StorageKernelBucketName, filename, strings.Join([]string{path, filename}, string(os.PathSeparator))); err != nil {
		return err
	}

	return nil
}

func (r *runnerInitializer) initializeRuntime(ctx context.Context, runtimeName string, runtimeVersion string) error {
	path := naming.RuntimeStoragePath(r.dataPath, runtimeName, runtimeVersion)
	filename := naming.RuntimeImageFileName(runtimeName, runtimeVersion)

	log.Debugf("check if runtime is already initialized")
	exists, _ := utils.FileExists(strings.Join([]string{path, filename}, string(os.PathSeparator)))
	if exists {
		log.Infof("runtime is already initialized: %s %s", runtimeName, runtimeVersion)
		return nil
	}

	log.Infof("initializing runtime: %s %s", runtimeName, runtimeVersion)
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
	if err := container.ImageExport(ctx, dockerClient, log, path, refString, filename); err != nil {
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
