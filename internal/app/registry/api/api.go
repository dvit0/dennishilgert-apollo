package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/dennishilgert/apollo/internal/app/registry/lease"
	"github.com/dennishilgert/apollo/internal/pkg/health"
	"github.com/dennishilgert/apollo/internal/pkg/logger"
	registrypb "github.com/dennishilgert/apollo/internal/pkg/proto/registry/v1"
	sharedpb "github.com/dennishilgert/apollo/internal/pkg/proto/shared/v1"
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

	port         int
	leaseService lease.LeaseService
	readyCh      chan struct{}
	running      atomic.Bool
}

// NewApiServer creates a new ApiServer instance.
func NewApiServer(leaseService lease.LeaseService, opts Options) ApiServer {
	return &apiServer{
		port:         opts.Port,
		leaseService: leaseService,
		readyCh:      make(chan struct{}),
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
	stopped := make(chan bool, 1)
	go func() {
		server.GracefulStop()
		close(stopped)
	}()

	timer := time.NewTimer(5 * time.Second)
	select {
	case <-timer.C:
		log.Warn("api server did not stop gracefully in time - forcing shutdown")
		server.Stop()
	case <-stopped:
		timer.Stop()
	}

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

// AcquireLease acquires a lease for a service or worker instance.
func (a *apiServer) AcquireLease(ctx context.Context, req *registrypb.AcquireLeaseRequest) (*sharedpb.EmptyResponse, error) {
	if err := a.leaseService.AquireLease(ctx, req); err != nil {
		return nil, fmt.Errorf("failed to acquire lease: %w", err)
	}
	return &sharedpb.EmptyResponse{}, nil
}

// ReleaseLease releases a lease for a service or worker instance.
func (a *apiServer) ReleaseLease(ctx context.Context, req *registrypb.ReleaseLeaseRequest) (*sharedpb.EmptyResponse, error) {
	if err := a.leaseService.ReleaseLease(ctx, req.InstanceUuid, req.InstanceType.String()); err != nil {
		return nil, fmt.Errorf("failed to release lease: %w", err)
	}
	return &sharedpb.EmptyResponse{}, nil
}

// AvailableServiceInstance returns an available instance for a given service type.
func (a *apiServer) AvailableServiceInstance(ctx context.Context, req *registrypb.AvailableInstanceRequest) (*registrypb.AvailableInstanceResponse, error) {
	instance, err := a.leaseService.AvailableServiceInstance(ctx, req.InstanceType)
	if err != nil {
		return nil, fmt.Errorf("failed to get available service instance from cache: %w", err)
	}
	return &registrypb.AvailableInstanceResponse{
		Instance: instance,
	}, nil
}
