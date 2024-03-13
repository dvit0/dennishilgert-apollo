package fleet

import (
	"context"
	"fmt"
	"time"

	"github.com/dennishilgert/apollo/internal/app/fleet/api"
	"github.com/dennishilgert/apollo/internal/app/fleet/operator"
	"github.com/dennishilgert/apollo/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/pkg/health"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/utils"
)

var log = logger.NewLogger("apollo.manager")

type Options struct {
	ApiPort               int
	FirecrackerBinaryPath string
	WatchdogCheckInterval time.Duration
	WatchdogWorkerCount   int
	AgentApiPort          int
}

type Manager interface {
	Run(ctx context.Context) error
}

type manager struct {
	runnerOperator operator.Operator
	apiServer      api.Server
}

func NewManager(opts Options) (Manager, error) {
	runnerOperator, err := operator.NewRunnerOperator(operator.Options{
		OsArch:                utils.DetectArchitecture(),
		FirecrackerBinaryPath: opts.FirecrackerBinaryPath,
		WatchdogCheckInterval: opts.WatchdogCheckInterval,
		WatchdogWorkerCount:   opts.WatchdogWorkerCount,
		AgentApiPort:          opts.AgentApiPort,
	})
	if err != nil {
		return nil, fmt.Errorf("error while creating vm operator: %v", err)
	}

	apiServer := api.NewApiServer(runnerOperator, api.Options{
		Port: opts.ApiPort,
	})

	return &manager{
		runnerOperator: runnerOperator,
		apiServer:      apiServer,
	}, nil
}

func (m *manager) Run(ctx context.Context) error {
	log.Info("apollo manager is starting")

	healthStatusProvider := health.NewHealthStatusProvider(health.ProviderOptions{
		Targets: 1,
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
			log.Info("starting api server")
			if err := m.apiServer.Run(ctx, healthStatusProvider); err != nil {
				return fmt.Errorf("failed to start api server: %v", err)
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