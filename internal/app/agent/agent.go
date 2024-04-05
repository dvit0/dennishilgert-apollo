package agent

import (
	"context"
	"fmt"

	"github.com/dennishilgert/apollo/internal/app/agent/api"
	"github.com/dennishilgert/apollo/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/pkg/health"
	"github.com/dennishilgert/apollo/pkg/logger"
)

var log = logger.NewLogger("apollo.agent")

// Agent is an Apollo application that runs inside a Firecracker Micro VM.
type Agent interface {
	Run(ctx context.Context) error
}

// Options contains the options for `NewAgent`.
type Options struct {
	ApiPort int
}

type agent struct {
	apiServer api.Server
}

func NewAgent(opts Options) (Agent, error) {
	return &agent{
		apiServer: api.NewApiServer(api.Options{
			Port: opts.ApiPort,
		}),
	}, nil
}

func (a *agent) Run(ctx context.Context) error {
	log.Info("apollo agent is starting")

	healthStatusProvider := health.NewHealthStatusProvider(health.ProviderOptions{
		Targets: 1,
	})

	runner := runner.NewRunnerManager(
		func(ctx context.Context) error {
			log.Info("starting api server")
			if err := a.apiServer.Run(ctx, healthStatusProvider); err != nil {
				return fmt.Errorf("failed to start api server: %v", err)
			}
			return nil
		},
		func(ctx context.Context) error {
			if err := a.apiServer.Ready(ctx); err != nil {
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
