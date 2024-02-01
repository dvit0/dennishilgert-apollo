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

	err := logger.ApplyOptionsToLoggers(&opts.Logger)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("starting apollo manager -- version %s", "TO_BE_IMPLEMENTED")
	log.Infof("log level set to: %s", opts.Logger.OutputLevel)

	ctx := signals.Context()
	manager, err := manager.NewManager(ctx, manager.Options{
		FirecrackerBinaryPath: opts.FirecrackerBinaryPath,
		VmHealthCheckInterval: time.Duration(opts.VmHealthCheckInterval) * time.Second,
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
