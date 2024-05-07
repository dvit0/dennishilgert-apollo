package app

import (
	"github.com/dennishilgert/apollo/cmd/worker/config"
	"github.com/dennishilgert/apollo/internal/app/worker"
	"github.com/dennishilgert/apollo/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/signals"
	"github.com/joho/godotenv"
)

var log = logger.NewLogger("apollo.manager")

func Run() {
	// Load environment variables from .env file for local development.
	godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	log.Info("starting apollo worker manager -- version TO_BE_IMPLEMENTED")
	log.Info("log level set to: ", log.LogLevel())

	ctx := signals.Context()
	manager, err := worker.NewManager(
		ctx,
		worker.Options{
			ApiPort:                      cfg.ApiPort,
			MessagingBootstrapServers:    cfg.MessagingBootstrapServers,
			MessagingWorkerCount:         cfg.MessagingWorkerCount,
			CacheAddress:                 cfg.CacheAddress,
			CacheUsername:                cfg.CacheUsername,
			CachePassword:                cfg.CachePassword,
			CacheDatabase:                cfg.CacheDatabase,
			ServiceRegistryAddress:       cfg.ServiceRegistryAddress,
			HeartbeatInterval:            cfg.HeartbeatInterval,
			FunctionInitializationFactor: cfg.FunctionInitializationFactor,
		},
	)
	if err != nil {
		log.Fatalf("error while creating worker manager: %v", err)
	}

	err = runner.NewRunnerManager(
		manager.Run,
	).Run(ctx)
	if err != nil {
		log.Fatalf("error while running worker manager: %v", err)
	}

	log.Info("worker manager shut down gracefully")
}
