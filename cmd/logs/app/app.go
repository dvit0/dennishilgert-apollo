package app

import (
	"github.com/dennishilgert/apollo/cmd/logs/config"
	"github.com/dennishilgert/apollo/internal/app/logs"
	"github.com/dennishilgert/apollo/internal/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/internal/pkg/logger"
	"github.com/dennishilgert/apollo/internal/pkg/signals"
	"github.com/joho/godotenv"
)

var log = logger.NewLogger("apollo.logs")

// Run starts the logs service.
func Run() {
	// Load environment variables from .env file for local development.
	godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("starting apollo logs service -- version %s", "TO_BE_IMPLEMENTED")
	log.Infof("log level set to: %s", log.LogLevel())

	ctx := signals.Context()
	logsService, err := logs.NewLogsService(
		ctx,
		logs.Options{
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
			DatabaseAuthDb:            cfg.DatabaseAuthDb,
		},
	)
	if err != nil {
		log.Fatalf("error while creating logs service: %v", err)
	}

	err = runner.NewRunnerManager(
		logsService.Run,
	).Run(ctx)
	if err != nil {
		log.Fatalf("error while running logs service: %v", err)
	}

	log.Info("logs service shut down gracefully")
}
