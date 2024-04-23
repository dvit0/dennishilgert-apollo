package app

import (
	"github.com/dennishilgert/apollo/cmd/registry/config"
	"github.com/dennishilgert/apollo/internal/app/registry"
	"github.com/dennishilgert/apollo/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/signals"
	"github.com/joho/godotenv"
)

var log = logger.NewLogger("apollo.registry")

// Run starts the service registry.
func Run() {
	// Load environment variables from .env file for local development.
	godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("starting apollo service registry -- version %s", "TO_BE_IMPLEMENTED")
	log.Infof("log level set to: %s", log.LogLevel())

	ctx := signals.Context()
	registry, err := registry.NewServiceRegistry(registry.Options{
		ApiPort:                   cfg.ApiPort,
		MessagingBootstrapServers: cfg.MessagingBootstrapServers,
		MessagingWorkerCount:      cfg.MessagingWorkerCount,
		CacheAddress:              cfg.CacheAddress,
		CacheUsername:             cfg.CacheUsername,
		CachePassword:             cfg.CachePassword,
		CacheDatabase:             cfg.CacheDatabase,
	})
	if err != nil {
		log.Fatalf("error while creating service registry: %v", err)
	}

	err = runner.NewRunnerManager(
		registry.Run,
	).Run(ctx)
	if err != nil {
		log.Fatalf("error while running service registry: %v", err)
	}

	log.Info("service registry shut down gracefully")
}
