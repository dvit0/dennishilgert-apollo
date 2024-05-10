package operator

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dennishilgert/apollo/internal/pkg/concurrency/worker"
	"github.com/dennishilgert/apollo/internal/pkg/container"
	"github.com/dennishilgert/apollo/internal/pkg/logger"
	"github.com/dennishilgert/apollo/internal/pkg/messaging/producer"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
	fleetpb "github.com/dennishilgert/apollo/internal/pkg/proto/fleet/v1"
	frontendpb "github.com/dennishilgert/apollo/internal/pkg/proto/frontend/v1"
	messagespb "github.com/dennishilgert/apollo/internal/pkg/proto/messages/v1"
	"github.com/dennishilgert/apollo/internal/pkg/storage"
	"github.com/dennishilgert/apollo/internal/pkg/utils"
)

var log = logger.NewLogger("apollo.package.operator")

type Options struct {
	WorkingDir           string
	WorkerCount          int
	ImageRegistryAddress string
}

type PackageOperator interface {
	Run(ctx context.Context) error
	DownloadDependencies(ctx context.Context) error
	BuildPackage(function *fleetpb.FunctionSpecs)
}

type packageOperator struct {
	workingDir           string
	imageRegistryAddress string
	workerManager        worker.WorkerManager
	storageService       storage.StorageService
	messagingProducer    producer.MessagingProducer
}

func NewPackageOperator(storageService storage.StorageService, messagingProducer producer.MessagingProducer, opts Options) PackageOperator {
	workerManager := worker.NewWorkerManager(opts.WorkerCount)

	return &packageOperator{
		workingDir:           opts.WorkingDir,
		imageRegistryAddress: opts.ImageRegistryAddress,
		workerManager:        workerManager,
		storageService:       storageService,
		messagingProducer:    messagingProducer,
	}
}

func (p *packageOperator) Run(ctx context.Context) error {
	if err := p.workerManager.Run(ctx); err != nil {
		return err
	}
	return nil
}

func (p *packageOperator) DownloadDependencies(ctx context.Context) error {
	log.Infof("downloading dependencies")
	dockerfiles, err := p.storageService.ListContents(ctx, naming.StorageDependenciesBucketName, "dockerfiles/")
	if err != nil {
		return fmt.Errorf("failed to list dockerfiles: %w", err)
	}
	for _, dockerfile := range dockerfiles {
		dockerfilePath := strings.Join([]string{p.workingDir, dockerfile}, string(os.PathSeparator))
		exists, _ := utils.FileExists(dockerfilePath)
		if exists {
			log.Infof("dockerfile already present: %s", dockerfile)
			continue
		}
		if !exists {
			log.Infof("downloading dockerfile: %s", dockerfile)
			if err := p.storageService.DownloadObject(ctx, naming.StorageDependenciesBucketName, dockerfile, dockerfilePath); err != nil {
				return fmt.Errorf("failed to download object: %w", err)
			}
		}
	}
	return nil
}

func (p *packageOperator) BuildPackage(function *fleetpb.FunctionSpecs) {
	task := worker.NewTask(func(ctx context.Context) (*fleetpb.FunctionSpecs, error) {
		buildDir, err := p.createBuildDir()
		if err != nil {
			return function, fmt.Errorf("failed to build package: %w", err)
		}
		defer os.RemoveAll(buildDir)

		log.Infof("building package in: %s", buildDir)

		destFileName := "archive.zip"
		tempDir, err := os.MkdirTemp("", "build-*")
		if err != nil {
			return function, fmt.Errorf("failed to create temporary directory: %w", err)
		}
		defer os.RemoveAll(tempDir)

		fileName := naming.FunctionCodeStorageName(function.Uuid, function.Version)
		destFilePath := strings.Join([]string{tempDir, destFileName}, string(os.PathSeparator))
		if err := p.storageService.DownloadObject(ctx, naming.StorageFunctionBucketName, fileName, destFilePath); err != nil {
			return function, fmt.Errorf("failed to download object: %w", err)
		}

		extractedPath := strings.Join([]string{buildDir, "extracted"}, string(os.PathSeparator))
		if err := os.MkdirAll(extractedPath, 0777); err != nil {
			return function, fmt.Errorf("failed to create extracted directory: %w", err)
		}
		if err := utils.Unzip(destFilePath, extractedPath); err != nil {
			return function, fmt.Errorf("failed to unzip file: %w", err)
		}

		dockerfileName := "Dockerfile.package"
		if err := utils.CopyFile(
			strings.Join([]string{p.workingDir, "dockerfiles", dockerfileName}, string(os.PathSeparator)),
			strings.Join([]string{buildDir, dockerfileName}, string(os.PathSeparator)),
		); err != nil {
			return function, fmt.Errorf("failed to copy Dockerfile: %w", err)
		}

		dockerClient, err := container.GetDefaultClient()
		if err != nil {
			return function, fmt.Errorf("failed to get default docker client: %w", err)
		}
		refString := naming.ImageRefStr(p.imageRegistryAddress, function.Uuid, function.Version)
		if err := container.ImageBuild(ctx, dockerClient, log, buildDir, dockerfileName, refString); err != nil {
			return function, fmt.Errorf("failed to build image: %w", err)
		}
		if err := container.ImagePush(ctx, dockerClient, log, refString); err != nil {
			return function, fmt.Errorf("failed to push image: %w", err)
		}

		return function, nil
	}, 60*time.Second)

	task.Callback(func(function *fleetpb.FunctionSpecs, err error) {
		if err != nil {
			log.Errorf("failed to build package: %v", err)
			p.messagingProducer.Publish(context.Background(), naming.MessagingFunctionStatusUpdateTopic, messagespb.FunctionStatusUpdateMessage{
				Function: function,
				Status:   frontendpb.FunctionStatus_PACKING_FAILED,
				Reason:   err.Error(),
			})
			return
		}
		p.messagingProducer.Publish(context.Background(), naming.MessagingFunctionStatusUpdateTopic, messagespb.FunctionStatusUpdateMessage{
			Function: function,
			Status:   frontendpb.FunctionStatus_PACKED,
			Reason:   "ok",
		})
		log.Infof("package built successfully")
	})

	p.workerManager.Add(task)
}

func (p *packageOperator) createBuildDir() (string, error) {
	currentMillis := time.Now().UnixMilli()
	dirName := fmt.Sprintf("build-%d", currentMillis)
	dirPath := strings.Join([]string{p.workingDir, dirName}, string(os.PathSeparator))

	if err := os.MkdirAll(dirPath, 0777); err != nil {
		return "", fmt.Errorf("failed to create build directory: %w", err)
	}
	return dirPath, nil
}
