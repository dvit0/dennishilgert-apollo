package gateway

import (
	"context"
	"fmt"

	"github.com/dennishilgert/apollo/internal/app/gateway/rest"
	"github.com/dennishilgert/apollo/internal/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/internal/pkg/health"
	"github.com/dennishilgert/apollo/internal/pkg/logger"
	"github.com/dennishilgert/apollo/internal/pkg/registry"
	"github.com/google/uuid"
)

var log = logger.NewLogger("apollo.gateway")

type Options struct {
	ApiPort                int
	ServiceRegistryAddress string
}

type ApiGateway interface {
	Run(ctx context.Context) error
}

type apiGateway struct {
	instanceUuid          string
	apiPort               int
	serviceRegistryClient registry.ServiceRegistryClient
	restServer            rest.RestServer
}

func NewApiGateway(opts Options) ApiGateway {
	instanceUuid := uuid.NewString()

	// Assign nil to the metricsService and messagingProducer as they are not used in this context.
	serviceRegistryClient := registry.NewServiceRegistryClient(
		nil,
		nil,
		registry.Options{
			Address:      opts.ServiceRegistryAddress,
			InstanceUuid: instanceUuid,
		},
	)

	restServer := rest.NewRestServer(
		serviceRegistryClient,
		rest.Options{
			ApiPort: opts.ApiPort,
		},
	)

	return &apiGateway{
		instanceUuid:          instanceUuid,
		apiPort:               opts.ApiPort,
		serviceRegistryClient: serviceRegistryClient,
		restServer:            restServer,
	}
}

func (g *apiGateway) Run(ctx context.Context) error {
	log.Info("apollo api gateway is starting")

	healthStatusProvider := health.NewHealthStatusProvider(health.ProviderOptions{
		Targets: 1,
	})

	runner := runner.NewRunnerManager(
		func(ctx context.Context) error {
			log.Info("establishing connection to service registry")
			if err := g.serviceRegistryClient.EstablishConnection(ctx); err != nil {
				return fmt.Errorf("failed to establish connection to service registry: %w", err)
			}

			healthStatusProvider.Ready()

			// Wait for the main context to be done.
			<-ctx.Done()

			log.Info("shutting down service registry connection")
			g.serviceRegistryClient.CloseConnection()

			return nil
		},
		func(ctx context.Context) error {
			log.Info("starting rest server")
			if err := g.restServer.Run(); err != nil {
				return fmt.Errorf("failed to start rest server: %w", err)
			}
			return nil
		},
		func(ctx context.Context) error {
			// Wait for the main context to be done.
			<-ctx.Done()

			return g.restServer.Shutdown()
		},
	)
	return runner.Run(ctx)
}
