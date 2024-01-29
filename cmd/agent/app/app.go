package app

import (
	"os"

	"github.com/dennishilgert/apollo/cmd/agent/options"
	"github.com/dennishilgert/apollo/internal/app/agent"
	"github.com/dennishilgert/apollo/pkg/concurrency"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/signals"
)

var log = logger.NewLogger("apollo.agent")

func Run() {
	opts := options.New(os.Args[1:])

	err := logger.ApplyOptionsToLoggers(&opts.Logger)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("starting apollo agent -- version %s", "TO_BE_IMPLEMENTED")
	log.Infof("log level set to: %s", opts.Logger.OutputLevel)

	ctx := signals.Context()
	agent, err := agent.NewAgent(ctx, agent.Options{
		ApiPort: opts.ApiPort,
	})
	if err != nil {
		log.Fatalf("error while creating agent: %v", err)
	}

	err = concurrency.NewRunnerManager(
		agent.Run,
	).Run(ctx)
	if err != nil {
		log.Fatalf("error while running agent: %v", err)
	}

	log.Info("agent shut down gracefully")
}
