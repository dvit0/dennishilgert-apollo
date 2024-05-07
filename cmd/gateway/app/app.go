package app

import (
	"github.com/dennishilgert/apollo/cmd/gateway/config"
	"github.com/dennishilgert/apollo/internal/app/gateway"
	"github.com/dennishilgert/apollo/internal/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/internal/pkg/logger"
	"github.com/dennishilgert/apollo/internal/pkg/signals"
	"github.com/joho/godotenv"
)

var log = logger.NewLogger("apollo.gateway")

// Run starts the frontend.
func Run() {
	// Load environment variables from .env file for local development.
	godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("starting apollo api gateway -- version %s", "TO_BE_IMPLEMENTED")
	log.Infof("log level set to: %s", log.LogLevel())

	ctx := signals.Context()
	gateway := gateway.NewApiGateway(
		gateway.Options{
			ApiPort:                cfg.ApiPort,
			ServiceRegistryAddress: cfg.ServiceRegistryAddress,
		},
	)
	if err != nil {
		log.Fatalf("error while creating api gateway: %v", err)
	}

	err = runner.NewRunnerManager(
		gateway.Run,
	).Run(ctx)
	if err != nil {
		log.Fatalf("error while running api gateway: %v", err)
	}

	log.Info("api gateway shut down gracefully")
}
