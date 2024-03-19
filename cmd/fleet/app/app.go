package app

import (
	"time"

	"github.com/dennishilgert/apollo/cmd/fleet/config"

	"github.com/dennishilgert/apollo/internal/app/fleet"
	"github.com/dennishilgert/apollo/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/signals"
	"github.com/joho/godotenv"
)

var log = logger.NewLogger("apollo.manager")

func Run() {
	// load environment variables from .env file for local development
	godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("starting apollo manager -- version %s", "TO_BE_IMPLEMENTED")
	log.Infof("log level set to: %s", log.LogLevel())

	ctx := signals.Context()
	manager, err := fleet.NewManager(fleet.Options{
		ApiPort:                cfg.ApiPort,
		DataPath:               cfg.DataPath,
		FirecrackerBinaryPath:  cfg.FirecrackerBinaryPath,
		DockerImageRegistryUrl: cfg.DockerImageRegistryUrl,
		WatchdogCheckInterval:  time.Duration(cfg.WatchdogCheckInterval) * time.Second,
		WatchdogWorkerCount:    cfg.WatchdogWorkerCount,
		AgentApiPort:           cfg.AgentApiPort,
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
