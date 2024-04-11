package fleet

import (
	"context"
	"fmt"
	"time"

	"github.com/dennishilgert/apollo/internal/app/fleet/api"
	"github.com/dennishilgert/apollo/internal/app/fleet/initializer"
	"github.com/dennishilgert/apollo/internal/app/fleet/messaging"
	"github.com/dennishilgert/apollo/internal/app/fleet/messaging/consumer"
	"github.com/dennishilgert/apollo/internal/app/fleet/operator"
	"github.com/dennishilgert/apollo/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/pkg/health"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/messaging/producer"
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
}

type FleetManager interface {
	Run(ctx context.Context) error
}

type fleetManager struct {
	workerUuid                string
	messagingBootstrapServers string
	runnerOperator            operator.RunnerOperator
	runnerInitializer         initializer.RunnerInitializer
	apiServer                 api.Server
	messagingProducer         producer.MessagingProducer
	messagingConsumer         consumer.MessagingConsumer
}

func NewManager(ctx context.Context, opts Options) (FleetManager, error) {
	workerUuid := uuid.NewString()

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
			AgentApiPort:              opts.AgentApiPort,
			MessagingBootstrapServers: opts.MessagingBootstrapServers,
			OsArch:                    utils.DetectArchitecture(),
			FirecrackerBinaryPath:     opts.FirecrackerBinaryPath,
			WatchdogCheckInterval:     opts.WatchdogCheckInterval,
			WatchdogWorkerCount:       opts.WatchdogWorkerCount,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error while creating runner operator: %v", err)
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

	messagingConsumer, err := consumer.NewMessagingConsumer(
		runnerOperator,
		consumer.Options{
			WorkerUuid:       workerUuid,
			BootstrapServers: opts.MessagingBootstrapServers,
			WorkerCount:      opts.MessagingWorkerCount,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error while creating messaging consumer: %v", err)
	}

	apiServer := api.NewApiServer(
		runnerOperator,
		runnerInitializer,
		messagingProducer,
		api.Options{
			Port: opts.ApiPort,
		},
	)

	return &fleetManager{
		workerUuid:                workerUuid,
		messagingBootstrapServers: opts.MessagingBootstrapServers,
		runnerOperator:            runnerOperator,
		runnerInitializer:         runnerInitializer,
		apiServer:                 apiServer,
		messagingProducer:         messagingProducer,
		messagingConsumer:         messagingConsumer,
	}, nil
}

func (m *fleetManager) Run(ctx context.Context) error {
	log.Info("apollo manager is starting")

	// Cleanup after context done
	defer func() {
		if err := m.messagingProducer.Close(); err != nil {
			log.Errorf("failed to close messaging producer: %v", err)
		}
	}()

	healthStatusProvider := health.NewHealthStatusProvider(health.ProviderOptions{
		Targets: 2,
	})

	runner := runner.NewRunnerManager(
		func(ctx context.Context) error {
			log.Info("initializing runner operator")
			if err := m.runnerOperator.Init(ctx); err != nil {
				log.Error("failed to initialize runner operator")
				return err
			}
			return nil
		},
		func(ctx context.Context) error {
			log.Info("preparing filesystem")
			if err := m.runnerInitializer.InitializeDataDir(); err != nil {
				log.Error("failed to initialize data directory")
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
			if err := m.apiServer.Run(ctx, healthStatusProvider); err != nil {
				log.Error("failed to start api server")
				return err
			}
			return nil
		},
		func(ctx context.Context) error {
			if err := m.apiServer.Ready(ctx); err != nil {
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
			log.Info("setting up messaging")
			if err := messaging.CreateRelatedTopic(ctx, m.messagingBootstrapServers, m.workerUuid); err != nil {
				log.Error("failed to setup messaging")
				return err
			}
			// Signalize that the messaging setup is done.
			m.messagingConsumer.SetupDone()

			// Wait for the main context to be done.
			<-ctx.Done()
			log.Info("cleaning up messaging")
			cleanupCtx, cleanupCtxCancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cleanupCtxCancel()
			if err := messaging.DeleteRelatedTopic(cleanupCtx, m.messagingBootstrapServers, m.workerUuid); err != nil {
				log.Error("failed to cleanup messaging")
				return err
			}
			return nil
		},
		func(ctx context.Context) error {
			log.Info("starting messaging consumer")
			if err := m.messagingConsumer.Start(ctx); err != nil {
				log.Error("failed to start messaging consumer")
				return err
			}
			return nil
		},
	)
	return runner.Run(ctx)
}
