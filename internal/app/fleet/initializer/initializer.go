package initializer

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/dennishilgert/apollo/internal/app/fleet/operator/runner"
	"github.com/dennishilgert/apollo/internal/pkg/container"
	"github.com/dennishilgert/apollo/internal/pkg/logger"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
	fleetpb "github.com/dennishilgert/apollo/internal/pkg/proto/fleet/v1"
	"github.com/dennishilgert/apollo/internal/pkg/storage"
	"github.com/dennishilgert/apollo/internal/pkg/utils"
)

var log = logger.NewLogger("apollo.initializer")

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
	DeinitializeFunction(ctx context.Context, request *fleetpb.DeinitializeFunctionRequest) error
	InitializedFunctions() []string
}

type runnerInitializer struct {
	dataPath             string
	imageRegistryAddress string
	storageService       storage.StorageService
}

// NewRunnerInitializer creates a new RunnerInitializer instance.
func NewRunnerInitializer(storageService storage.StorageService, opts Options) RunnerInitializer {
	return &runnerInitializer{
		dataPath:             opts.DataPath,
		imageRegistryAddress: opts.ImageRegistryAddress,
		storageService:       storageService,
	}
}

// DataPath returns the data path.
func (r *runnerInitializer) DataPath() string {
	return r.dataPath
}

// InitializeDataDir initializes the data directory.
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

// InitializeRunner initializes a runner.
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

// RemoveRunner removes a runner.
func (r *runnerInitializer) RemoveRunner(ctx context.Context, runnerUuid string) error {
	path := naming.RunnerStoragePath(r.dataPath, runnerUuid)

	if err := os.RemoveAll(path); err != nil {
		return err
	}

	return nil
}

// InitializeFunction initializes a function.
func (r *runnerInitializer) InitializeFunction(ctx context.Context, request *fleetpb.InitializeFunctionRequest) error {
	path := naming.FunctionStoragePath(r.dataPath, request.Function.Uuid)
	filename := naming.FunctionImageFileName(request.Function.Version)
	functionIdentifier := naming.FunctionIdentifier(request.Function.Uuid, request.Function.Version)

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
		log.Infof("function is already initialized: %s", functionIdentifier)
		return nil
	}

	log.Infof("initializing function: %s", functionIdentifier)
	if err := prepareTargetDirectory(path); err != nil {
		return err
	}
	dockerClient, err := container.GetDefaultClient()
	if err != nil {
		return err
	}
	refString := naming.ImageRefStr(r.imageRegistryAddress, request.Function.Uuid, request.Function.Version)
	if request.Function.Version == naming.RuntimeInitialCodeDeclarator() {
		// If the function version is the initial code declarator, then use the runtime specific initial image.
		refString = naming.RuntimeInitialImageRefStr(r.imageRegistryAddress, request.Runtime.Name, request.Runtime.Version, request.Runtime.Architecture)
	}

	log.Infof("pulling function image: %s", refString)
	if err := container.ImagePull(ctx, dockerClient, log, refString); err != nil {
		log.Errorf("failed to pull function image: %v", err)
		return err
	}
	log.Infof("exporting function image: %s", refString)
	if err := container.ImageExport(ctx, dockerClient, log, path, refString, filename, []string{"/code"}); err != nil {
		return err
	}

	return nil
}

func (r *runnerInitializer) DeinitializeFunction(ctx context.Context, request *fleetpb.DeinitializeFunctionRequest) error {
	path := naming.FunctionStoragePath(r.dataPath, request.Function.Uuid)
	filename := naming.FunctionImageFileName(request.Function.Version)
	functionIdentifier := naming.FunctionIdentifier(request.Function.Uuid, request.Function.Version)

	log.Debugf("check if function is initialized")
	exists, _ := utils.FileExists(strings.Join([]string{path, filename}, string(os.PathSeparator)))
	if !exists {
		log.Infof("function is not initialized: %s", functionIdentifier)
		return nil
	}

	log.Infof("deinitializing function: %s", functionIdentifier)
	if err := os.RemoveAll(strings.Join([]string{path, filename}, string(os.PathSeparator))); err != nil {
		return err
	}

	return nil
}

// InitializedFunctions returns a list of initialized functions.
func (r *runnerInitializer) InitializedFunctions() []string {
	log.Debug("checking for initialized functions")
	initializedFunctions := make([]string, 0)
	path := naming.FunctionStoragePathBase(r.dataPath)
	if exists, _ := utils.FileExists(path); !exists {
		log.Debugf("no functions are initialized")
		return initializedFunctions
	}
	dirs, err := os.ReadDir(path)
	if err != nil {
		log.Errorf("failed to read directory: %v", err)
		return initializedFunctions
	}
	for _, dir := range dirs {
		log.Debugf("checking function %s for versions", dir.Name())
		if !dir.IsDir() {
			continue
		}
		if (dir.Name() == ".") || (dir.Name() == "..") {
			continue
		}
		functionUuid := dir.Name()
		versionsPath := naming.FunctionStoragePath(r.dataPath, functionUuid)
		versions, err := os.ReadDir(versionsPath)
		if err != nil {
			log.Errorf("failed to read function directory: %v", err)
			continue
		}
		for _, version := range versions {
			if version.IsDir() {
				continue
			}
			if (version.Name() == ".") || (version.Name() == "..") {
				continue
			}
			functionVersion := naming.FunctionExtractVersionFromImageFileName(version.Name())

			functionIdentifier := naming.FunctionIdentifier(functionUuid, functionVersion)
			initializedFunctions = append(initializedFunctions, functionIdentifier)
			log.Debugf("found initialized function: %s", functionIdentifier)
		}
	}
	return initializedFunctions
}

// initializeKernel initializes a kernel.
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

// initializeRuntime initializes a runtime.
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
	if err := container.ImageExport(ctx, dockerClient, log, path, refString, filename, []string{}); err != nil {
		return err
	}

	return nil
}

// prepareTargetDirectory prepares the target directory.
func prepareTargetDirectory(path string) error {
	exists, fileInfo := utils.FileExists(path)
	if exists {
		ok, err := utils.IsDirAndWritable(path, fileInfo)
		if !ok {
			return fmt.Errorf("target path is not a directory or not writable: %v", err)
		}
	} else {
		if err := os.MkdirAll(path, 0777); err != nil {
			return fmt.Errorf("failed to create target directory: %v", err)
		}
	}
	return nil
}
