package agent

import (
	"context"
	"fmt"

	"github.com/dennishilgert/apollo/internal/app/agent/api"
	"github.com/dennishilgert/apollo/pkg/concurrency"
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

func NewAgent(ctx context.Context, opts Options) (Agent, error) {
	return &agent{
		apiServer: api.NewAPIServer(api.Options{
			Port: opts.ApiPort,
		}),
	}, nil
}

func (a *agent) Run(ctx context.Context) error {
	log.Info("apollo agent is starting")

	runner := concurrency.NewRunnerManager(
		func(ctx context.Context) error {
			log.Info("starting api server")
			err := a.apiServer.Run(ctx)
			if err != nil {
				return fmt.Errorf("failed to start api server: %v", err)
			}
			return nil
		},
	)

	return runner.Run(ctx)
}
