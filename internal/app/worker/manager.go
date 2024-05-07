package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/dennishilgert/apollo/internal/app/worker/api"
	"github.com/dennishilgert/apollo/internal/app/worker/messaging"
	"github.com/dennishilgert/apollo/internal/app/worker/placement"
	"github.com/dennishilgert/apollo/internal/pkg/cache"
	"github.com/dennishilgert/apollo/internal/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/internal/pkg/health"
	"github.com/dennishilgert/apollo/internal/pkg/logger"
	"github.com/dennishilgert/apollo/internal/pkg/messaging/consumer"
	"github.com/dennishilgert/apollo/internal/pkg/messaging/producer"
	"github.com/dennishilgert/apollo/internal/pkg/metrics"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
	registrypb "github.com/dennishilgert/apollo/internal/pkg/proto/registry/v1"
	"github.com/dennishilgert/apollo/internal/pkg/registry"
	"github.com/dennishilgert/apollo/internal/pkg/utils"
	"github.com/google/uuid"
)

var log = logger.NewLogger("apollo.manager")

type Options struct {
	ApiPort                      int
	MessagingBootstrapServers    string
	MessagingWorkerCount         int
	CacheAddress                 string
	CacheUsername                string
	CachePassword                string
	CacheDatabase                int
	ServiceRegistryAddress       string
	HeartbeatInterval            int
	FunctionInitializationFactor int
}

type WorkerManager interface {
	Run(ctx context.Context) error
}

type workerManager struct {
	instanceUuid          string
	instanceType          registrypb.InstanceType
	ipAddress             string
	apiPort               int
	apiServer             api.ApiServer
	cacheClient           cache.CacheClient
	placementService      placement.PlacementService
	metricsService        metrics.MetricsService
	messagingProducer     producer.MessagingProducer
	messagingConsumer     consumer.MessagingConsumer
	serviceRegistryClient registry.ServiceRegistryClient
}

func NewManager(ctx context.Context, opts Options) (WorkerManager, error) {
	instanceUuid := uuid.NewString()
	instanceType := registrypb.InstanceType_WORKER_MANAGER

	ipAddress, err := utils.PrimaryIp()
	if err != nil {
		return nil, fmt.Errorf("getting primary ip failed: %w", err)
	}

	// Initialize metrics service with "nil" as the runner pool metrics function is not used in a service instance.
	metricsService := metrics.NewMetricsService(nil)

	messagingProducer, err := producer.NewMessagingProducer(
		ctx,
		producer.Options{
			BootstrapServers: opts.MessagingBootstrapServers,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("creating messaging producer failed: %w", err)
	}

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

	cacheClient := cache.NewCacheClient(
		instanceUuid,
		cache.Options{
			Address:  opts.CacheAddress,
			Username: opts.CacheUsername,
			Password: opts.CachePassword,
			Database: opts.CacheDatabase,
		},
	)

	placementService := placement.NewPlacementService(
		cacheClient,
		placement.Options{
			FunctionInitializationFactor: opts.FunctionInitializationFactor,
		},
	)

	messagingHandler := messaging.NewMessagingHandler(
		placementService,
		messaging.Options{},
	)
	messagingHandler.RegisterAll()

	messagingConsumer, err := consumer.NewMessagingConsumer(
		messagingHandler,
		consumer.Options{
			GroupId: "apollo_worker_manager",
			Topics: []string{
				naming.MessagingFunctionInitializationTopic,
				naming.MessagingFunctionDeinitializationTopic,
			},
			BootstrapServers: opts.MessagingBootstrapServers,
			WorkerCount:      opts.MessagingWorkerCount,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error while creating messaging consumer: %w", err)
	}

	apiServer := api.NewApiServer(
		placementService,
		api.Options{
			Port: opts.ApiPort,
		},
	)

	return &workerManager{
		instanceUuid:          instanceUuid,
		instanceType:          instanceType,
		ipAddress:             ipAddress,
		apiPort:               opts.ApiPort,
		apiServer:             apiServer,
		cacheClient:           cacheClient,
		metricsService:        metricsService,
		placementService:      placementService,
		messagingProducer:     messagingProducer,
		messagingConsumer:     messagingConsumer,
		serviceRegistryClient: serviceRegistryClient,
	}, nil
}

func (w *workerManager) Run(ctx context.Context) error {
	log.Info("apollo worker manager is starting")

	healthStatusProvider := health.NewHealthStatusProvider(health.ProviderOptions{
		Targets: 2,
	})

	runner := runner.NewRunnerManager(
		func(ctx context.Context) error {
			log.Info("establishing connection to service registry")
			if err := w.serviceRegistryClient.EstablishConnection(ctx); err != nil {
				return fmt.Errorf("failed to establish connection to service registry: %w", err)
			}

			log.Info("acquiring lease")
			metrics := w.metricsService.ServiceInstanceMetrics()
			acquireLeaseRequest := &registrypb.AcquireLeaseRequest{
				Instance: &registrypb.AcquireLeaseRequest_ServiceInstance{
					ServiceInstance: &registrypb.ServiceInstance{
						InstanceUuid: w.instanceUuid,
						InstanceType: w.instanceType,
						Host:         w.ipAddress,
						Port:         int32(w.apiPort),
						Metadata:     nil,
					},
				},
				Metrics: &registrypb.AcquireLeaseRequest_ServiceInstanceMetrics{
					ServiceInstanceMetrics: metrics,
				},
			}
			if err := w.serviceRegistryClient.AcquireLease(ctx, acquireLeaseRequest); err != nil {
				return fmt.Errorf("failed to acquire lease: %w", err)
			}

			healthStatusProvider.Ready()
			log.Info("lease acquired successfully")

			if err := w.serviceRegistryClient.SendHeartbeat(ctx); err != nil {
				return fmt.Errorf("failed to send heartbeat: %w", err)
			}

			log.Info("releasing lease")
			releaseCtx, releaseCtxCancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer releaseCtxCancel()
			if err := w.serviceRegistryClient.ReleaseLease(releaseCtx); err != nil {
				return fmt.Errorf("failed to release lease: %w", err)
			}

			log.Info("shutting down service registry connection")
			w.serviceRegistryClient.CloseConnection()

			return nil
		},
		func(ctx context.Context) error {
			// Wait for the main context to be done.
			<-ctx.Done()
			log.Info("closing cache client")
			w.cacheClient.Close()
			return nil
		},
		func(ctx context.Context) error {
			log.Info("starting api server")
			if err := w.apiServer.Run(ctx, healthStatusProvider); err != nil {
				log.Error("failed to start api server")
				return err
			}
			return nil
		},
		func(ctx context.Context) error {
			if err := w.apiServer.Ready(ctx); err != nil {
				log.Error("api server did not become ready in time")
				return err
			}
			healthStatusProvider.Ready()
			log.Info("api server started")

			// Wait for the main context to be done.
			<-ctx.Done()
			return nil
		},
		func(ctx context.Context) error {
			// Signalize that the messaging setup is done.
			w.messagingConsumer.SetupDone()

			log.Info("starting messaging consumer")
			if err := w.messagingConsumer.Start(ctx); err != nil {
				log.Error("failed to start messaging consumer")
				return err
			}
			return nil
		},
	)
	return runner.Run(ctx)
}
