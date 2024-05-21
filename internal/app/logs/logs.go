package logs

import (
	"context"
	"fmt"
	"time"

	"github.com/dennishilgert/apollo/internal/app/logs/api"
	"github.com/dennishilgert/apollo/internal/app/logs/db"
	"github.com/dennishilgert/apollo/internal/app/logs/messaging"
	"github.com/dennishilgert/apollo/internal/app/logs/operator"
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

var log = logger.NewLogger("apollo.logs")

type Options struct {
	ApiPort                   int
	MessagingBootstrapServers string
	MessagingWorkerCount      int
	ServiceRegistryAddress    string
	HeartbeatInterval         int
	DatabaseHost              string
	DatabasePort              int
	DatabaseUsername          string
	DatabasePassword          string
	DatabaseDb                string
	DatabaseAuthDb            string
}

type LogsService interface {
	Run(ctx context.Context) error
}

type logsService struct {
	instanceUuid          string
	instanceType          registrypb.InstanceType
	ipAddress             string
	apiPort               int
	databaseClient        db.DatabaseClient
	messagingProducer     producer.MessagingProducer
	messagingConsumer     consumer.MessagingConsumer
	metricsService        metrics.MetricsService
	serviceRegistryClient registry.ServiceRegistryClient
	logsOperator          operator.LogsOperator
	apiServer             api.Server
}

func NewLogsService(ctx context.Context, opts Options) (LogsService, error) {
	instanceUuid := uuid.NewString()
	instanceType := registrypb.InstanceType_LOGS_SERVICE

	ipAddress, err := utils.PrimaryIp()
	if err != nil {
		return nil, fmt.Errorf("getting primary ip failed: %w", err)
	}

	databaseClient, err := db.NewDatabaseClient(
		ctx,
		db.Options{
			Host:         opts.DatabaseHost,
			Port:         opts.DatabasePort,
			Username:     opts.DatabaseUsername,
			Password:     opts.DatabasePassword,
			Database:     opts.DatabaseDb,
			AuthDatabase: opts.DatabaseAuthDb,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create database client: %w", err)
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

	logsOperator := operator.NewLogsOperator(
		databaseClient,
	)

	messagingHandler := messaging.NewMessagingHandler(
		logsOperator,
		messaging.Options{},
	)
	messagingHandler.RegisterAll()

	messagingConsumer, err := consumer.NewMessagingConsumer(
		messagingHandler,
		consumer.Options{
			GroupId: "apollo_logs_service",
			Topics: []string{
				naming.MessagingFunctionInvocationLogsTopic,
			},
			BootstrapServers: opts.MessagingBootstrapServers,
			WorkerCount:      opts.MessagingWorkerCount,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error while creating messaging consumer: %v", err)
	}

	apiServer := api.NewApiServer(
		logsOperator,
		api.Options{
			Port: opts.ApiPort,
		},
	)

	return &logsService{
		instanceUuid:          instanceUuid,
		instanceType:          instanceType,
		ipAddress:             ipAddress,
		apiPort:               opts.ApiPort,
		databaseClient:        databaseClient,
		messagingProducer:     messagingProducer,
		messagingConsumer:     messagingConsumer,
		metricsService:        metricsService,
		serviceRegistryClient: serviceRegistryClient,
		logsOperator:          logsOperator,
		apiServer:             apiServer,
	}, nil
}

func (f *logsService) Run(ctx context.Context) error {
	log.Info("apollo logs service is starting")

	healthStatusProvider := health.NewHealthStatusProvider(health.ProviderOptions{
		Targets: 2,
	})

	runner := runner.NewRunnerManager(
		func(ctx context.Context) error {
			log.Info("establishing connection to service registry")
			if err := f.serviceRegistryClient.EstablishConnection(ctx); err != nil {
				return fmt.Errorf("failed to establish connection to service registry: %w", err)
			}

			log.Info("acquiring lease")
			metrics := f.metricsService.ServiceInstanceMetrics()
			acquireLeaseRequest := &registrypb.AcquireLeaseRequest{
				Instance: &registrypb.AcquireLeaseRequest_ServiceInstance{
					ServiceInstance: &registrypb.ServiceInstance{
						InstanceUuid: f.instanceUuid,
						InstanceType: f.instanceType,
						Host:         f.ipAddress,
						Port:         int32(f.apiPort),
						Metadata:     nil,
					},
				},
				Metrics: &registrypb.AcquireLeaseRequest_ServiceInstanceMetrics{
					ServiceInstanceMetrics: metrics,
				},
			}
			if err := f.serviceRegistryClient.AcquireLease(ctx, acquireLeaseRequest); err != nil {
				return fmt.Errorf("failed to acquire lease: %w", err)
			}

			healthStatusProvider.Ready()
			log.Info("lease acquired successfully")

			if err := f.serviceRegistryClient.SendHeartbeat(ctx); err != nil {
				return fmt.Errorf("failed to send heartbeat: %w", err)
			}

			log.Info("releasing lease")
			releaseCtx, releaseCtxCancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer releaseCtxCancel()
			if err := f.serviceRegistryClient.ReleaseLease(releaseCtx); err != nil {
				return fmt.Errorf("failed to release lease: %w", err)
			}

			log.Info("shutting down service registry connection")
			f.serviceRegistryClient.CloseConnection()

			return nil
		},
		func(ctx context.Context) error {
			// Wait for the main context to be done.
			<-ctx.Done()

			if err := f.databaseClient.Close(); err != nil {
				log.Errorf("failed to close database client: %v", err)
				return err
			}
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
			// Signalize that the messaging setup is done.
			f.messagingConsumer.SetupDone()

			log.Info("starting messaging consumer")
			if err := f.messagingConsumer.Start(ctx); err != nil {
				log.Error("failed to start messaging consumer")
				return err
			}
			return nil
		},
	)
	return runner.Run(ctx)
}
