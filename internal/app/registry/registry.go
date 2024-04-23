package registry

import (
	"context"
	"fmt"
	"time"

	"github.com/dennishilgert/apollo/internal/app/registry/api"
	"github.com/dennishilgert/apollo/internal/app/registry/cache"
	"github.com/dennishilgert/apollo/internal/app/registry/messaging"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
	"github.com/dennishilgert/apollo/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/pkg/health"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/messaging/consumer"
	"github.com/google/uuid"
)

var log = logger.NewLogger("apollo.registry")

type Options struct {
	ApiPort                   int
	MessagingBootstrapServers string
	MessagingWorkerCount      int
	CacheAddress              string
	CacheUsername             string
	CachePassword             string
	CacheDatabase             int
}

type ServiceRegistry interface {
	Run(ctx context.Context) error
}

type serviceRegistry struct {
	instanceUuid      string
	apiServer         api.ApiServer
	cacheClient       cache.CacheClient
	messagingConsumer consumer.MessagingConsumer
}

func NewServiceRegistry(opts Options) (ServiceRegistry, error) {
	instanceUuid := uuid.NewString()

	cacheClient := cache.NewCacheClient(
		instanceUuid,
		cache.Options{
			Address:           opts.CacheAddress,
			Username:          opts.CacheUsername,
			Password:          opts.CachePassword,
			Database:          opts.CacheDatabase,
			ExpirationTimeout: 10 * time.Second,
		},
	)

	messagingHandler := messaging.NewMessagingHandler(
		cacheClient,
		messaging.Options{},
	)
	messagingHandler.RegisterAll()

	messagingConsumer, err := consumer.NewMessagingConsumer(
		messagingHandler,
		consumer.Options{
			GroupId: "apollo_service_registry",
			Topics: []string{
				naming.MessagingInstanceHeartbeatTopic,
			},
			BootstrapServers: opts.MessagingBootstrapServers,
			WorkerCount:      opts.MessagingWorkerCount,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error while creating messaging consumer: %v", err)
	}

	apiServer := api.NewApiServer(
		cacheClient,
		api.Options{
			Port: opts.ApiPort,
		},
	)

	return &serviceRegistry{
		instanceUuid:      instanceUuid,
		apiServer:         apiServer,
		cacheClient:       cacheClient,
		messagingConsumer: messagingConsumer,
	}, nil
}

// Run starts the service registry and all its components.
func (s *serviceRegistry) Run(ctx context.Context) error {
	log.Info("apollo service registry is starting")

	healthStatusProvider := health.NewHealthStatusProvider(health.ProviderOptions{
		Targets: 1,
	})

	runner := runner.NewRunnerManager(
		func(ctx context.Context) error {
			log.Info("starting cache client listener")
			if err := s.cacheClient.Listen(ctx); err != nil {
				log.Error("failed to start cache client listener")
				return err
			}

			s.cacheClient.Close()
			return nil
		},
		func(ctx context.Context) error {
			log.Info("starting api server")
			if err := s.apiServer.Run(ctx, healthStatusProvider); err != nil {
				log.Error("failed to start api server")
				return err
			}
			return nil
		},
		func(ctx context.Context) error {
			if err := s.apiServer.Ready(ctx); err != nil {
				log.Error("api server did not become ready in time")
				return err
			}
			healthStatusProvider.Ready()
			log.Info("api server started")

			// Wait for the main context to be done.
			<-ctx.Done()
			return nil
		},
	)
	return runner.Run(ctx)
}
