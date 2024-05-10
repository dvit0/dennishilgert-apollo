package pack

import (
	"context"
	"fmt"
	"time"

	"github.com/dennishilgert/apollo/internal/app/pack/api"
	"github.com/dennishilgert/apollo/internal/app/pack/messaging"
	"github.com/dennishilgert/apollo/internal/app/pack/operator"
	"github.com/dennishilgert/apollo/internal/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/internal/pkg/health"
	"github.com/dennishilgert/apollo/internal/pkg/logger"
	"github.com/dennishilgert/apollo/internal/pkg/messaging/consumer"
	"github.com/dennishilgert/apollo/internal/pkg/messaging/producer"
	"github.com/dennishilgert/apollo/internal/pkg/metrics"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
	registrypb "github.com/dennishilgert/apollo/internal/pkg/proto/registry/v1"
	"github.com/dennishilgert/apollo/internal/pkg/registry"
	"github.com/dennishilgert/apollo/internal/pkg/storage"
	"github.com/dennishilgert/apollo/internal/pkg/utils"
	"github.com/google/uuid"
)

var log = logger.NewLogger("apollo.package")

type Options struct {
	ApiPort                   int
	WorkingDir                string
	PackageWorkerCount        int
	ImageRegistryAddress      string
	MessagingBootstrapServers string
	MessagingWorkerCount      int
	ServiceRegistryAddress    string
	HeartbeatInterval         int
	StorageEndpoint           string
	StorageAccessKeyId        string
	StorageSecretAccessKey    string
}

type PackageService interface {
	Run(ctx context.Context) error
}

type packageService struct {
	instanceUuid          string
	instanceType          registrypb.InstanceType
	ipAddress             string
	apiPort               int
	storageService        storage.StorageService
	packageOperator       operator.PackageOperator
	messagingProducer     producer.MessagingProducer
	messagingConsumer     consumer.MessagingConsumer
	metricsService        metrics.MetricsService
	serviceRegistryClient registry.ServiceRegistryClient
	apiServer             api.Server
}

func NewPackageService(ctx context.Context, opts Options) (PackageService, error) {
	instanceUuid := uuid.NewString()
	instanceType := registrypb.InstanceType_PACKAGE_SERVICE

	ipAddress, err := utils.PrimaryIp()
	if err != nil {
		return nil, fmt.Errorf("getting primary ip failed: %w", err)
	}

	storageService, err := storage.NewStorageService(storage.Options{
		Endpoint:        opts.StorageEndpoint,
		AccessKeyId:     opts.StorageAccessKeyId,
		SecretAccessKey: opts.StorageSecretAccessKey,
	})
	if err != nil {
		return nil, fmt.Errorf("error while creating storage service: %v", err)
	}

	messagingProducer, err := producer.NewMessagingProducer(
		ctx,
		producer.Options{
			BootstrapServers: opts.MessagingBootstrapServers,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error while creating messaging producer: %v", err)
	}

	packageOperator := operator.NewPackageOperator(
		storageService,
		messagingProducer,
		operator.Options{
			WorkingDir:           opts.WorkingDir,
			WorkerCount:          opts.PackageWorkerCount,
			ImageRegistryAddress: opts.ImageRegistryAddress,
		},
	)
	if err := packageOperator.DownloadDependencies(ctx); err != nil {
		return nil, fmt.Errorf("error while downloading dependencies: %w", err)
	}

	// Initialize metrics service with "nil" as the runner pool metrics function is not used in a service instance.
	metricsService := metrics.NewMetricsService(nil)

	serviceRegistryClient := registry.NewServiceRegistryClient(
		metricsService,
		messagingProducer,
		registry.Options{
			Address:           opts.ServiceRegistryAddress,
			HeartbeatInterval: time.Duration(opts.HeartbeatInterval) * time.Second,
			InstanceUuid:      instanceUuid,
			InstanceType:      instanceType,
		},
	)

	messagingHandler := messaging.NewMessagingHandler(
		packageOperator,
		messagingProducer,
		messaging.Options{},
	)
	messagingHandler.RegisterAll()

	messagingConsumer, err := consumer.NewMessagingConsumer(
		messagingHandler,
		consumer.Options{
			GroupId: "apollo_frontend",
			Topics: []string{
				naming.MessagingFunctionCodeUploadedTopic,
			},
			BootstrapServers: opts.MessagingBootstrapServers,
			WorkerCount:      opts.MessagingWorkerCount,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error while creating messaging consumer: %v", err)
	}

	apiServer := api.NewApiServer(
		storageService,
		api.Options{
			Port: opts.ApiPort,
		},
	)

	return &packageService{
		instanceUuid:          instanceUuid,
		instanceType:          instanceType,
		ipAddress:             ipAddress,
		apiPort:               opts.ApiPort,
		storageService:        storageService,
		packageOperator:       packageOperator,
		messagingProducer:     messagingProducer,
		messagingConsumer:     messagingConsumer,
		metricsService:        metricsService,
		serviceRegistryClient: serviceRegistryClient,
		apiServer:             apiServer,
	}, nil
}

func (p *packageService) Run(ctx context.Context) error {
	log.Info("apollo package service is starting")

	healthStatusProvider := health.NewHealthStatusProvider(health.ProviderOptions{
		Targets: 2,
	})

	runner := runner.NewRunnerManager(
		func(ctx context.Context) error {
			log.Info("establishing connection to service registry")
			if err := p.serviceRegistryClient.EstablishConnection(ctx); err != nil {
				return fmt.Errorf("failed to establish connection to service registry: %w", err)
			}

			log.Info("acquiring lease")
			metrics := p.metricsService.ServiceInstanceMetrics()
			acquireLeaseRequest := &registrypb.AcquireLeaseRequest{
				Instance: &registrypb.AcquireLeaseRequest_ServiceInstance{
					ServiceInstance: &registrypb.ServiceInstance{
						InstanceUuid: p.instanceUuid,
						InstanceType: p.instanceType,
						Host:         p.ipAddress,
						Port:         int32(p.apiPort),
						Metadata:     nil,
					},
				},
				Metrics: &registrypb.AcquireLeaseRequest_ServiceInstanceMetrics{
					ServiceInstanceMetrics: metrics,
				},
			}
			if err := p.serviceRegistryClient.AcquireLease(ctx, acquireLeaseRequest); err != nil {
				return fmt.Errorf("failed to acquire lease: %w", err)
			}

			healthStatusProvider.Ready()
			log.Info("lease acquired successfully")

			if err := p.serviceRegistryClient.SendHeartbeat(ctx); err != nil {
				return fmt.Errorf("failed to send heartbeat: %w", err)
			}

			log.Info("releasing lease")
			releaseCtx, releaseCtxCancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer releaseCtxCancel()
			if err := p.serviceRegistryClient.ReleaseLease(releaseCtx); err != nil {
				return fmt.Errorf("failed to release lease: %w", err)
			}

			log.Info("shutting down service registry connection")
			p.serviceRegistryClient.CloseConnection()

			return nil
		},
		func(ctx context.Context) error {
			log.Info("starting api server")
			if err := p.apiServer.Run(ctx, healthStatusProvider); err != nil {
				log.Errorf("failed to start api server: %v", err)
				return err
			}
			return nil
		},
		func(ctx context.Context) error {
			if err := p.apiServer.Ready(ctx); err != nil {
				log.Errorf("api server did not become ready in time: %v", err)
				return err
			}
			healthStatusProvider.Ready()
			log.Info("api server started")

			// Wait for the main context to be done.
			<-ctx.Done()
			return nil
		},
		func(ctx context.Context) error {
			log.Info("starting package operator")
			if err := p.packageOperator.Run(ctx); err != nil {
				log.Error("failed to start package operator")
				return err
			}
			return nil
		},
		func(ctx context.Context) error {
			// Signalize that the messaging setup is done.
			p.messagingConsumer.SetupDone()

			log.Info("starting messaging consumer")
			if err := p.messagingConsumer.Start(ctx); err != nil {
				log.Error("failed to start messaging consumer")
				return err
			}
			return nil
		},
	)
	return runner.Run(ctx)
}
