package api

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"sync/atomic"
	"time"

	"github.com/dennishilgert/apollo/internal/app/worker/placement"
	"github.com/dennishilgert/apollo/pkg/health"
	"github.com/dennishilgert/apollo/pkg/logger"
	sharedpb "github.com/dennishilgert/apollo/pkg/proto/shared/v1"
	workerpb "github.com/dennishilgert/apollo/pkg/proto/worker/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	workerInstance, err := a.placementService.FunctionInitializationWorker(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get matching worker: %w", err)
	}
	clientConn, err := establishConnection(ctx, fmt.Sprintf("%s:%d", workerInstance.Host, workerInstance.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection to worker instance: %w", err)
	}
	apiClient := workerpb.NewWorkerManagerClient(clientConn)
	_, err = apiClient.InitializeFunction(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize function on worker instance: %w", err)
	}
	return &sharedpb.EmptyResponse{}, nil
}

func (a *apiServer) AllocateRunner(ctx context.Context, req *workerpb.AllocateRunnerRequest) (*workerpb.AllocateRunnerResponse, error) {
	workerInstance, err := a.placementService.RunnerAllocationWorker(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get matching worker: %w", err)
	}
	clientConn, err := establishConnection(ctx, fmt.Sprintf("%s:%d", workerInstance.Host, workerInstance.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection to worker instance: %w", err)
	}
	apiClient := workerpb.NewWorkerManagerClient(clientConn)
	res, err := apiClient.AllocateRunner(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate runner on worker instance: %w", err)
	}
	return res, nil
}

func establishConnection(ctx context.Context, address string) (*grpc.ClientConn, error) {
	const retrySeconds = 3     // trying to connect for a period of 3 seconds
	const retriesPerSecond = 2 // trying to connect 2 times per second
	for i := 0; i < (retrySeconds * retriesPerSecond); i++ {
		conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
		if err == nil {
			return conn, nil
		} else {
			if conn != nil {
				conn.Close()
			}
			log.Errorf("failed to establish connection to service registry - reason: %v", err)
		}
		// Wait before retrying, but stop if context is done.
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context done before connection to servie registry could be established: %w", ctx.Err())
		case <-time.After(time.Duration(math.Round(1000/retriesPerSecond)) * time.Millisecond): // retry delay
			continue
		}
	}
	return nil, fmt.Errorf("failed to establish connection to service registry after %d seconds", retrySeconds)
}
