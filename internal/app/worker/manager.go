package worker

import (
	"context"

	"github.com/dennishilgert/apollo/internal/app/worker/api"
	"github.com/dennishilgert/apollo/internal/app/worker/placement"
	"github.com/dennishilgert/apollo/pkg/cache"
	"github.com/dennishilgert/apollo/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/pkg/health"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/google/uuid"
)

var log = logger.NewLogger("apollo.manager")

type Options struct {
	ApiPort                   int
	MessagingBootstrapServers string
	MessagingWorkerCount      int
	CacheAddress              string
	CacheUsername             string
	CachePassword             string
	CacheDatabase             int
}

type WorkerManager interface {
	Run(ctx context.Context) error
}

type workerManager struct {
	instanceUuid     string
	apiServer        api.ApiServer
	cacheClient      cache.CacheClient
	placementService placement.PlacementService
}

func NewManager(ctx context.Context, opts Options) (WorkerManager, error) {
	instanceUuid := uuid.NewString()

	cacheClient := cache.NewCacheClient(
		instanceUuid,
		cache.Options{
			Address:  opts.CacheAddress,
			Username: opts.CacheUsername,
			Password: opts.CachePassword,
			Database: opts.CacheDatabase,
		},
	)

	placementService := placement.NewPlacementService(
		cacheClient,
		placement.Options{},
	)

	apiServer := api.NewApiServer(api.Options{
		Port: opts.ApiPort,
	})

	return &workerManager{
		instanceUuid:     instanceUuid,
		apiServer:        apiServer,
		cacheClient:      cacheClient,
		placementService: placementService,
	}, nil
}

func (w *workerManager) Run(ctx context.Context) error {
	log.Info("apollo worker manager is starting")

	healthStatusProvider := health.NewHealthStatusProvider(health.ProviderOptions{
		Targets: 1,
	})

	runner := runner.NewRunnerManager(
		func(ctx context.Context) error {
			// Wait for the main context to be done.
			<-ctx.Done()
			log.Info("closing cache client")
			w.cacheClient.Close()
			return nil
		},
		func(ctx context.Context) error {
			log.Info("starting api server")
			if err := w.apiServer.Run(ctx, healthStatusProvider); err != nil {
				log.Error("failed to start api server")
				return err
			}
			return nil
		},
		func(ctx context.Context) error {
			if err := w.apiServer.Ready(ctx); err != nil {
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
