package app

import (
	"os"
	"time"

	"github.com/dennishilgert/apollo/cmd/manager/options"
	"github.com/dennishilgert/apollo/internal/app/manager"
	"github.com/dennishilgert/apollo/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/signals"
)

var log = logger.NewLogger("apollo.manager")

func Run() {
	opts := options.New(os.Args[1:])

	if err := logger.ApplyOptionsToLoggers(&opts.Logger); err != nil {
		log.Fatal(err)
	}

	log.Infof("starting apollo manager -- version %s", "TO_BE_IMPLEMENTED")
	log.Infof("log level set to: %s", opts.Logger.OutputLevel)

	ctx := signals.Context()
	manager, err := manager.NewManager(manager.Options{
		ApiPort:               opts.ApiPort,
		FirecrackerBinaryPath: opts.FirecrackerBinaryPath,
		WatchdogCheckInterval: time.Duration(opts.WatchdogCheckInterval) * time.Second,
		WatchdogWorkerCount:   opts.WatchdogWorkerCount,
		AgentApiPort:          opts.AgentApiPort,
	})
	if err != nil {
		log.Fatalf("error while creating manager: %v", err)
	}

	err = runner.NewRunnerManager(
		manager.Run,
	).Run(ctx)
	if err != nil {
		log.Fatalf("error while running manager: %v", err)
	}

	log.Info("manager shut down gracefully")
}
