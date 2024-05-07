package app

import (
	"github.com/dennishilgert/apollo/cmd/frontend/config"
	"github.com/dennishilgert/apollo/internal/app/frontend"
	"github.com/dennishilgert/apollo/internal/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/internal/pkg/logger"
	"github.com/dennishilgert/apollo/internal/pkg/signals"
	"github.com/joho/godotenv"
)

var log = logger.NewLogger("apollo.frontend")

// Run starts the frontend.
func Run() {
	// Load environment variables from .env file for local development.
	godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("starting apollo frontend -- version %s", "TO_BE_IMPLEMENTED")
	log.Infof("log level set to: %s", log.LogLevel())

	ctx := signals.Context()
	frontend, err := frontend.NewFrontend(
		ctx,
		frontend.Options{
			ApiPort:                   cfg.ApiPort,
			MessagingBootstrapServers: cfg.MessagingBootstrapServers,
			MessagingWorkerCount:      cfg.MessagingWorkerCount,
			ServiceRegistryAddress:    cfg.ServiceRegistryAddress,
			HeartbeatInterval:         cfg.HeartbeatInterval,
			DatabaseHost:              cfg.DatabaseHost,
			DatabasePort:              cfg.DatabasePort,
			DatabaseUsername:          cfg.DatabaseUsername,
			DatabasePassword:          cfg.DatabasePassword,
			DatabaseDb:                cfg.DatabaseDb,
		},
	)
	if err != nil {
		log.Fatalf("error while creating frontend: %v", err)
	}

	err = runner.NewRunnerManager(
		frontend.Run,
	).Run(ctx)
	if err != nil {
		log.Fatalf("error while running frontend: %v", err)
	}

	log.Info("frontend shut down gracefully")
}
