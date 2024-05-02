package fleet

import (
	"context"
	"fmt"
	"time"

	"github.com/dennishilgert/apollo/internal/app/fleet/api"
	"github.com/dennishilgert/apollo/internal/app/fleet/initializer"
	"github.com/dennishilgert/apollo/internal/app/fleet/messaging"
	"github.com/dennishilgert/apollo/internal/app/fleet/messaging/handler"
	"github.com/dennishilgert/apollo/internal/app/fleet/operator"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
	"github.com/dennishilgert/apollo/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/pkg/health"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/messaging/consumer"
	"github.com/dennishilgert/apollo/pkg/messaging/producer"
	"github.com/dennishilgert/apollo/pkg/metrics"
	registrypb "github.com/dennishilgert/apollo/pkg/proto/registry/v1"
	"github.com/dennishilgert/apollo/pkg/registry"
	"github.com/dennishilgert/apollo/pkg/storage"
	"github.com/dennishilgert/apollo/pkg/utils"
	"github.com/google/uuid"
)

var log = logger.NewLogger("apollo.manager")

type Options struct {
	ApiPort                   int
	AgentApiPort              int
	DataPath                  string
	FirecrackerBinaryPath     string
	ImageRegistryAddress      string
	MessagingBootstrapServers string
	MessagingWorkerCount      int
	StorageEndpoint           string
	StorageAccessKeyId        string
	StorageSecretAccessKey    string
	WatchdogCheckInterval     time.Duration
	WatchdogWorkerCount       int
	ServiceRegistryAddress    string
	HeartbeatInterval         int
}

type FleetManager interface {
	Run(ctx context.Context) error
}

type fleetManager struct {
	workerUuid                string
	instanceType              registrypb.InstanceType
	architecture              utils.OsArch
	ipAddress                 string
	apiPort                   int
	messagingBootstrapServers string
	runnerOperator            operator.RunnerOperator
	runnerInitializer         initializer.RunnerInitializer
	apiServer                 api.Server
	messagingProducer         producer.MessagingProducer
	messagingConsumer         consumer.MessagingConsumer
	metricsService            metrics.MetricsService
	serviceRegistryClient     registry.ServiceRegistryClient
}

// NewManager creates a new FleetManager instance.
func NewManager(ctx context.Context, opts Options) (FleetManager, error) {
	workerUuid := uuid.NewString()
	instanceType := registrypb.InstanceType_FLEET_MANAGER
	architecture := utils.DetectArchitecture()

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

	runnerInitializer := initializer.NewRunnerInitializer(
		storageService,
		initializer.Options{
			DataPath:             opts.DataPath,
			ImageRegistryAddress: opts.ImageRegistryAddress,
		},
	)

	runnerOperator, err := operator.NewRunnerOperator(
		ctx,
		runnerInitializer,
		operator.Options{
			WorkerUuid:                workerUuid,
			IpAddress:                 ipAddress,
			ApiPort:                   opts.ApiPort,
			AgentApiPort:              opts.AgentApiPort,
			MessagingBootstrapServers: opts.MessagingBootstrapServers,
			OsArch:                    architecture,
			FirecrackerBinaryPath:     opts.FirecrackerBinaryPath,
			WatchdogCheckInterval:     opts.WatchdogCheckInterval,
			WatchdogWorkerCount:       opts.WatchdogWorkerCount,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error while creating runner operator: %v", err)
	}

	metricsService := metrics.NewMetricsService(
		runnerOperator.RunnerPoolMetrics,
	)

	messagingProducer, err := producer.NewMessagingProducer(
		ctx,
		producer.Options{
			BootstrapServers: opts.MessagingBootstrapServers,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error while creating messaging producer: %v", err)
	}

	messagingHandler := handler.NewMessagingHandler(
		runnerOperator,
		handler.Options{
			WorkerUuid: workerUuid,
		},
	)
	messagingHandler.RegisterAll()

	messagingConsumer, err := consumer.NewMessagingConsumer(
		messagingHandler,
		consumer.Options{
			GroupId: "apollo_fleet_manager",
			Topics: []string{
				naming.MessagingWorkerRelatedAgentReadyTopic(workerUuid),
			},
			BootstrapServers: opts.MessagingBootstrapServers,
			WorkerCount:      opts.MessagingWorkerCount,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error while creating messaging consumer: %v", err)
	}

	serviceRegistryClient := registry.NewServiceRegistryClient(
		metricsService,
		messagingProducer,
		registry.Options{
			Address:           opts.ServiceRegistryAddress,
			HeartbeatInterval: time.Duration(opts.HeartbeatInterval) * time.Second,
			InstanceUuid:      workerUuid,
			InstanceType:      instanceType,
		},
	)

	apiServer := api.NewApiServer(
		runnerOperator,
		runnerInitializer,
		messagingProducer,
		api.Options{
			Port:       opts.ApiPort,
			WorkerUuid: workerUuid,
		},
	)

	return &fleetManager{
		workerUuid:                workerUuid,
		instanceType:              instanceType,
		architecture:              architecture,
		ipAddress:                 ipAddress,
		apiPort:                   opts.ApiPort,
		messagingBootstrapServers: opts.MessagingBootstrapServers,
		runnerOperator:            runnerOperator,
		runnerInitializer:         runnerInitializer,
		apiServer:                 apiServer,
		messagingProducer:         messagingProducer,
		messagingConsumer:         messagingConsumer,
		metricsService:            metricsService,
		serviceRegistryClient:     serviceRegistryClient,
	}, nil
}

// Run starts the fleet manager.
func (f *fleetManager) Run(ctx context.Context) error {
	log.Info("apollo runner manager is starting")

	// Cleanup after context done
	defer func() {
		if err := f.messagingProducer.Close(); err != nil {
			log.Errorf("failed to close messaging producer: %v", err)
		}
	}()

	healthStatusProvider := health.NewHealthStatusProvider(health.ProviderOptions{
		Targets: 3,
	})

	runner := runner.NewRunnerManager(
		func(ctx context.Context) error {
			log.Info("establishing connection to service registry")
			if err := f.serviceRegistryClient.EstablishConnection(ctx); err != nil {
				log.Errorf("failed to establish connection to service registry: %v", err)
				return err
			}

			log.Info("acquiring lease")
			metrics := f.metricsService.WorkerInstanceMetrics()
			acquireLeaseRequest := &registrypb.AcquireLeaseRequest{
				Instance: &registrypb.AcquireLeaseRequest_WorkerInstance{
					WorkerInstance: &registrypb.WorkerInstance{
						WorkerUuid:           f.workerUuid,
						Architecture:         f.architecture.String(),
						Host:                 f.ipAddress,
						Port:                 int32(f.apiPort),
						InitializedFunctions: f.runnerInitializer.InitializedFunctions(),
					},
				},
				Metrics: &registrypb.AcquireLeaseRequest_WorkerInstanceMetrics{
					WorkerInstanceMetrics: metrics,
				},
			}
			if err := f.serviceRegistryClient.AcquireLease(ctx, acquireLeaseRequest); err != nil {
				log.Errorf("failed to acquire lease: %v", err)
				return err
			}

			healthStatusProvider.Ready()
			log.Info("lease acquired successfully")

			if err := f.serviceRegistryClient.SendHeartbeat(ctx); err != nil {
				log.Errorf("failed to send heartbeat: %v", err)
				return err
			}

			log.Info("releasing lease")
			releaseCtx, releaseCtxCancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer releaseCtxCancel()
			if err := f.serviceRegistryClient.ReleaseLease(releaseCtx); err != nil {
				log.Errorf("failed to release lease: %v", err)
				return err
			}

			log.Info("shutting down service registry connection")
			f.serviceRegistryClient.CloseConnection()

			return nil
		},
		func(ctx context.Context) error {
			log.Info("initializing runner operator")
			if err := f.runnerOperator.Init(ctx); err != nil {
				log.Errorf("failed to initialize runner operator: %v", err)
				return err
			}
			return nil
		},
		func(ctx context.Context) error {
			log.Info("preparing filesystem")
			if err := f.runnerInitializer.InitializeDataDir(); err != nil {
				log.Errorf("failed to initialize data directory: %v", err)
				return err
			}
			healthStatusProvider.Ready()
			log.Info("filesystem initialized")

			// Wait for the main context to be done.
			<-ctx.Done()
			return nil
		},
		func(ctx context.Context) error {
			log.Info("starting api server")
			if err := f.apiServer.Run(ctx, healthStatusProvider); err != nil {
				log.Errorf("failed to start api server: %v", err)
				return err
			}
			return nil
		},
		func(ctx context.Context) error {
			if err := f.apiServer.Ready(ctx); err != nil {
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
			log.Info("setting up messaging")
			if err := messaging.CreateRelatedTopic(ctx, f.messagingBootstrapServers, f.workerUuid); err != nil {
				log.Errorf("failed to create related topic: %v", err)
				return err
			}
			// Signalize that the messaging setup is done.
			f.messagingConsumer.SetupDone()

			healthStatusProvider.Ready()
			log.Info("messaging has been set up")

			// Wait for the main context to be done.
			<-ctx.Done()
			log.Info("cleaning up messaging")
			cleanupCtx, cleanupCtxCancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cleanupCtxCancel()
			if err := messaging.DeleteRelatedTopic(cleanupCtx, f.messagingBootstrapServers, f.workerUuid); err != nil {
				log.Errorf("failed to cleanup messaging: %v", err)
				return err
			}
			return nil
		},
		func(ctx context.Context) error {
			log.Info("starting messaging consumer")
			if err := f.messagingConsumer.Start(ctx); err != nil {
				log.Errorf("failed to start messaging consumer: %v", err)
				return err
			}
			return nil
		},
	)
	return runner.Run(ctx)
}
