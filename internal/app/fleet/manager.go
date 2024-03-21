package fleet

import (
	"context"
	"fmt"
	"time"

	"github.com/dennishilgert/apollo/internal/app/fleet/api"
	"github.com/dennishilgert/apollo/internal/app/fleet/messaging"
	"github.com/dennishilgert/apollo/internal/app/fleet/operator"
	"github.com/dennishilgert/apollo/internal/app/fleet/preparer"
	"github.com/dennishilgert/apollo/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/pkg/health"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/utils"
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
	WatchdogCheckInterval     time.Duration
	WatchdogWorkerCount       int
}

type FleetManager interface {
	Run(ctx context.Context) error
}

type fleetManager struct {
	runnerOperator   operator.RunnerOperator
	runnerPreparer   preparer.RunnerPreparer
	apiServer        api.Server
	messagingService messaging.MessagingService
}

func NewManager(opts Options) (FleetManager, error) {
	runnerOperator, err := operator.NewRunnerOperator(operator.Options{
		AgentApiPort:          opts.AgentApiPort,
		OsArch:                utils.DetectArchitecture(),
		FirecrackerBinaryPath: opts.FirecrackerBinaryPath,
		WatchdogCheckInterval: opts.WatchdogCheckInterval,
		WatchdogWorkerCount:   opts.WatchdogWorkerCount,
	})
	if err != nil {
		return nil, fmt.Errorf("error while creating runner operator: %v", err)
	}

	runnerPreparer := preparer.NewRunnerPreparer(preparer.Options{
		DataPath:             opts.DataPath,
		ImageRegistryAddress: opts.ImageRegistryAddress,
	})

	apiServer := api.NewApiServer(
		runnerOperator,
		runnerPreparer,
		api.Options{
			Port: opts.ApiPort,
		},
	)

	messagingService, err := messaging.NewMessagingService(
		runnerPreparer,
		messaging.Options{
			BootstrapServers: opts.MessagingBootstrapServers,
			WorkerCount:      opts.MessagingWorkerCount,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error while creating messaging service: %v", err)
	}

	return &fleetManager{
		runnerOperator:   runnerOperator,
		runnerPreparer:   runnerPreparer,
		apiServer:        apiServer,
		messagingService: messagingService,
	}, nil
}

func (m *fleetManager) Run(ctx context.Context) error {
	log.Info("apollo manager is starting")

	healthStatusProvider := health.NewHealthStatusProvider(health.ProviderOptions{
		Targets: 2,
	})

	runner := runner.NewRunnerManager(
		func(ctx context.Context) error {
			log.Info("initializing runner operator")
			if err := m.runnerOperator.Init(ctx); err != nil {
				return fmt.Errorf("failed to initialize runner operator: %v", err)
			}
			return nil
		},
		func(ctx context.Context) error {
			log.Info("preparing filesystem")
			if err := m.runnerPreparer.PrepareDataDir(); err != nil {
				return fmt.Errorf("failed to prepare data directory: %v", err)
			}
			healthStatusProvider.Ready()
			log.Info("filesystem prepared")
			<-ctx.Done()
			return nil
		},
		func(ctx context.Context) error {
			log.Info("starting api server")
			if err := m.apiServer.Run(ctx, healthStatusProvider); err != nil {
				return fmt.Errorf("failed to start api server: %v", err)
			}
			return nil
		},
		func(ctx context.Context) error {
			log.Info("starting messaging service")
			if err := m.messagingService.Start(ctx); err != nil {
				return fmt.Errorf("failed to start messaging service: %v", err)
			}
			return nil
		},
		func(ctx context.Context) error {
			if err := m.apiServer.Ready(ctx); err != nil {
				return fmt.Errorf("api server did not become ready in time: %v", err)
			}
			healthStatusProvider.Ready()
			log.Info("api server started")
			<-ctx.Done()
			return nil
		},
	)
	return runner.Run(ctx)
}
