package app

import (
	"os"

	"github.com/dennishilgert/apollo/cmd/manager/options"
	"github.com/dennishilgert/apollo/pkg/concurrency"
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

	err = concurrency.NewRunnerManager().Run(ctx)
	if err != nil {
		log.Fatalf("error while running manager: %v", err)
	}

	log.Info("manager shut down gracefully")
}
