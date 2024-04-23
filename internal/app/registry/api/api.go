package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/dennishilgert/apollo/internal/app/registry/cache"
	"github.com/dennishilgert/apollo/pkg/health"
	"github.com/dennishilgert/apollo/pkg/logger"
	registrypb "github.com/dennishilgert/apollo/pkg/proto/registry/v1"
	sharedpb "github.com/dennishilgert/apollo/pkg/proto/shared/v1"
	"google.golang.org/grpc"
)

var log = logger.NewLogger("apollo.registry.api")

type Options struct {
	Port int
}

type ApiServer interface {
	Run(ctx context.Context, healthStatusProvider health.Provider) error
	Ready(ctx context.Context) error
}

type apiServer struct {
	registrypb.UnimplementedServiceRegistryServer

	port        int
	cacheClient cache.CacheClient
	readyCh     chan struct{}
	running     atomic.Bool
}

// NewApiServer creates a new ApiServer instance.
func NewApiServer(cacheClient cache.CacheClient, opts Options) ApiServer {
	return &apiServer{
		port:        opts.Port,
		cacheClient: cacheClient,
		readyCh:     make(chan struct{}),
	}
}

// Run starts the api server and listens for incoming requests.
func (a *apiServer) Run(ctx context.Context, healthStatusProvider health.Provider) error {
	if !a.running.CompareAndSwap(false, true) {
		return errors.New("api server is already running")
	}

	log.Infof("starting api server on port %d", a.port)
	server := grpc.NewServer()
	registrypb.RegisterServiceRegistryServer(server, a)

	healthServer := health.NewHealthServer(healthStatusProvider, log)
	healthServer.Register(server)

	lis, lErr := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if lErr != nil {
		log.Error("error while starting tcp listener")
		return lErr
	}
	// Close the ready channel to signalize that the api server is ready.
	close(a.readyCh)

	errCh := make(chan error, 1)
	go func() {
		defer close(errCh) // Ensure channel is closed to avoid goroutine leak.

		if err := server.Serve(lis); err != nil {
			log.Error("error while serving api server")
			errCh <- err
			return
		}
		errCh <- nil
	}()

	// Block until the context is done or an error occurs.
	var serveErr error
	select {
	case <-ctx.Done():
		log.Info("shutting down api server")
	case err := <-errCh: // Handle errors that might have occurred during Serve.
		if err != nil {
			serveErr = err
			log.Errorf("error while listening for requests")
		}
	}

	// Perform graceful shutdown and close the listener regardless of the select outcome.
	server.GracefulStop()
	if cErr := lis.Close(); cErr != nil && !errors.Is(cErr, net.ErrClosed) && serveErr == nil {
		log.Error("error while closing api server listener")
		return cErr
	}

	return serveErr
}

// Ready waits until the api server is ready or the context is cancelled due to timeout.
func (a *apiServer) Ready(ctx context.Context) error {
	select {
	case <-a.readyCh:
		return nil
	case <-ctx.Done():
		return errors.New("timeout while waiting for the api server to be ready")
	}
}

// Acquire acquires a lease for a service or worker instance.
func (a *apiServer) Acquire(ctx context.Context, req *registrypb.AcquireLeaseRequest) (*sharedpb.EmptyResponse, error) {
	switch instance := req.Instance.(type) {
	case *registrypb.AcquireLeaseRequest_ServiceInstance:
		if err := a.cacheClient.AddInstance(ctx, instance.ServiceInstance); err != nil {
			return nil, fmt.Errorf("failed to add service instance to cache: %w", err)
		}
		return &sharedpb.EmptyResponse{}, nil
	case *registrypb.AcquireLeaseRequest_WorkerInstance:
		if err := a.cacheClient.AddWorker(ctx, instance.WorkerInstance); err != nil {
			return nil, fmt.Errorf("failed to add worker instance to cache: %w", err)
		}
		return &sharedpb.EmptyResponse{}, nil
	default:
		return nil, errors.New("unknown instance type")
	}
}

// Release releases a lease for a service or worker instance.
func (a *apiServer) Release(ctx context.Context, req *registrypb.ReleaseLeaseRequest) (*sharedpb.EmptyResponse, error) {
	switch req.ServiceType {
	case registrypb.ServiceType_FLEET_MANAGER:
		if err := a.cacheClient.RemoveWorker(ctx, req.InstanceUuid); err != nil {
			return nil, fmt.Errorf("failed to remove worker instance from cache: %w", err)
		}
	default:
		if err := a.cacheClient.RemoveInstance(ctx, req.ServiceType, req.InstanceUuid); err != nil {
			return nil, fmt.Errorf("failed to remove service instance from cache: %w", err)
		}
	}
	return nil, errors.New("unknown service type")
}

// Instance returns an available instance for a given service type.
func (a *apiServer) Instance(ctx context.Context, req *registrypb.AvailableInstanceRequest) (*registrypb.AvailableInstanceResponse, error) {
	instance, err := a.cacheClient.AvailableInstance(ctx, req.ServiceType)
	if err != nil {
		return nil, fmt.Errorf("failed to get available instance from cache: %w", err)
	}
	return &registrypb.AvailableInstanceResponse{
		Instance: instance,
	}, nil
}
