package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/dennishilgert/apollo/internal/app/worker/placement"
	"github.com/dennishilgert/apollo/pkg/health"
	"github.com/dennishilgert/apollo/pkg/logger"
	sharedpb "github.com/dennishilgert/apollo/pkg/proto/shared/v1"
	workerpb "github.com/dennishilgert/apollo/pkg/proto/worker/v1"
	"google.golang.org/grpc"
)

var log = logger.NewLogger("apollo.manager.api")

type Options struct {
	Port int
}

type ApiServer interface {
	Run(ctx context.Context, healthStatusProvider health.Provider) error
	Ready(ctx context.Context) error
}

type apiServer struct {
	workerpb.UnimplementedWorkerManagerServer

	port             int
	readyCh          chan struct{}
	running          atomic.Bool
	placementService placement.PlacementService
}

// NewApiServer creates a new ApiServer instance.
func NewApiServer(placementService placement.PlacementService, opts Options) ApiServer {
	return &apiServer{
		port:             opts.Port,
		readyCh:          make(chan struct{}),
		placementService: placementService,
	}
}

// Run starts the api server and listens for incoming requests.
func (a *apiServer) Run(ctx context.Context, healthStatusProvider health.Provider) error {
	if !a.running.CompareAndSwap(false, true) {
		return errors.New("api server is already running")
	}

	log.Infof("starting api server on port %d", a.port)
	server := grpc.NewServer()
	workerpb.RegisterWorkerManagerServer(server, a)

	healthServer := health.NewHealthServer(healthStatusProvider, log)
	healthServer.Register(server)

	lis, lErr := net.Listen("tcp", ":"+fmt.Sprint(a.port))
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

func (a *apiServer) InitializeFunction(ctx context.Context, req *workerpb.InitializeFunctionRequest) (*sharedpb.EmptyResponse, error) {
	if err := a.placementService.AllocateFunctionInitialization(ctx, req); err != nil {
		return nil, fmt.Errorf("failed to allocate function initialization: %w", err)
	}
	return &sharedpb.EmptyResponse{}, nil
}

func (a *apiServer) AllocateInvocation(ctx context.Context, req *workerpb.AllocateInvocationRequest) (*workerpb.AllocateInvocationResponse, error) {
	runner, err := a.placementService.FindAvailableRunner(ctx, req)
	if err == nil {
		return &workerpb.AllocateInvocationResponse{
			RunnerUuid:        runner.RunnerUuid,
			WorkerNodeAddress: runner.WorkerNodeAddress,
		}, nil
	}
	log.Debugf("no runner available: %v", err)

	res, err := a.placementService.AllocateRunnerProvisioning(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate runner provisioning: %w", err)
	}
	return &workerpb.AllocateInvocationResponse{
		RunnerUuid:        res.RunnerUuid,
		WorkerNodeAddress: res.WorkerNodeAddress,
	}, nil
}
