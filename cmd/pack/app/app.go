package app

import (
	"github.com/dennishilgert/apollo/cmd/pack/config"
	"github.com/dennishilgert/apollo/internal/app/pack"
	"github.com/dennishilgert/apollo/internal/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/internal/pkg/logger"
	"github.com/dennishilgert/apollo/internal/pkg/signals"
	"github.com/joho/godotenv"
)

var log = logger.NewLogger("apollo.package")

// Run starts the frontend.
func Run() {
	// Load environment variables from .env file for local development.
	godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("starting apollo package service -- version %s", "TO_BE_IMPLEMENTED")
	log.Infof("log level set to: %s", log.LogLevel())

	ctx := signals.Context()
	packageService, err := pack.NewPackageService(
		ctx,
		pack.Options{
			ApiPort:                   cfg.ApiPort,
			WorkingDir:                cfg.WorkingDir,
			PackageWorkerCount:        cfg.PackageWorkerCount,
			ImageRegistryAddress:      cfg.ImageRegistryAddress,
			MessagingBootstrapServers: cfg.MessagingBootstrapServers,
			MessagingWorkerCount:      cfg.MessagingWorkerCount,
			ServiceRegistryAddress:    cfg.ServiceRegistryAddress,
			HeartbeatInterval:         cfg.HeartbeatInterval,
			StorageEndpoint:           cfg.StorageEndpoint,
			StorageAccessKeyId:        cfg.StorageAccessKeyId,
			StorageSecretAccessKey:    cfg.StorageSecretAccessKey,
		},
	)
	if err != nil {
		log.Fatalf("error while creating package service: %v", err)
	}

	err = runner.NewRunnerManager(
		packageService.Run,
	).Run(ctx)
	if err != nil {
		log.Fatalf("error while running package service: %v", err)
	}

	log.Info("package service shut down gracefully")
}
