package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/dennishilgert/apollo/internal/app/frontend/operator"
	"github.com/dennishilgert/apollo/internal/pkg/health"
	"github.com/dennishilgert/apollo/internal/pkg/logger"
	fleetpb "github.com/dennishilgert/apollo/internal/pkg/proto/fleet/v1"
	frontendpb "github.com/dennishilgert/apollo/internal/pkg/proto/frontend/v1"
	sharedpb "github.com/dennishilgert/apollo/internal/pkg/proto/shared/v1"
	"google.golang.org/grpc"
)

var log = logger.NewLogger("apollo.api")

type Options struct {
	Port       int
	WorkerUuid string
}

type Server interface {
	Run(ctx context.Context, healthStatusProvider health.Provider) error
	Ready(ctx context.Context) error
}

type apiServer struct {
	frontendpb.UnimplementedFrontendServer

	port             int
	readyCh          chan struct{}
	running          atomic.Bool
	frontendOperator operator.FrontendOperator
}

// NewApiServer creates a new Server.
func NewApiServer(frontendOperator operator.FrontendOperator, opts Options) Server {
	return &apiServer{
		port:             opts.Port,
		readyCh:          make(chan struct{}),
		frontendOperator: frontendOperator,
	}
}

// Run runs the api server.
func (a *apiServer) Run(ctx context.Context, healthStatusProvider health.Provider) error {
	if !a.running.CompareAndSwap(false, true) {
		return errors.New("api server is already running")
	}

	log.Infof("starting api server on port %d", a.port)
	server := grpc.NewServer()
	frontendpb.RegisterFrontendServer(server, a)

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

func (a *apiServer) InvokeFunction(ctx context.Context, req *frontendpb.InvokeFunctionRequest) (*fleetpb.InvokeFunctionResponse, error) {
	res, err := a.frontendOperator.InvokeFunction(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke function: %w", err)
	}
	return res, nil
}

func (a *apiServer) AddKernel(ctx context.Context, req *frontendpb.AddKernelRequest) (*sharedpb.EmptyResponse, error) {
	if err := a.frontendOperator.AddKernel(ctx, req); err != nil {
		return nil, fmt.Errorf("failed to add kernel: %w", err)
	}
	return &sharedpb.EmptyResponse{}, nil
}

func (a *apiServer) ListKernels(ctx context.Context, req *frontendpb.ListKernelsRequest) (*frontendpb.ListKernelsResponse, error) {
	res, err := a.frontendOperator.ListKernels(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list kernels: %w", err)
	}
	return res, nil
}

func (a *apiServer) RemoveKernel(ctx context.Context, req *frontendpb.RemoveKernelRequest) (*sharedpb.EmptyResponse, error) {
	if err := a.frontendOperator.RemoveKernel(ctx, req); err != nil {
		return nil, fmt.Errorf("failed to remove kernel: %w", err)
	}
	return &sharedpb.EmptyResponse{}, nil
}

func (a *apiServer) AddRuntime(ctx context.Context, req *frontendpb.AddRuntimeRequest) (*sharedpb.EmptyResponse, error) {
	if err := a.frontendOperator.AddRuntime(ctx, req); err != nil {
		return nil, fmt.Errorf("failed to add runtime: %w", err)
	}
	return &sharedpb.EmptyResponse{}, nil
}

func (a *apiServer) ListRuntimes(ctx context.Context, req *frontendpb.ListRuntimesRequest) (*frontendpb.ListRuntimesResponse, error) {
	res, err := a.frontendOperator.ListRuntimes(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list runtimes: %w", err)
	}
	return res, nil
}

func (a *apiServer) RemoveRuntime(ctx context.Context, req *frontendpb.RemoveRuntimeRequest) (*sharedpb.EmptyResponse, error) {
	if err := a.frontendOperator.RemoveRuntime(ctx, req); err != nil {
		return nil, fmt.Errorf("failed to remove runtime: %w", err)
	}
	return &sharedpb.EmptyResponse{}, nil
}

func (a *apiServer) CreateFunction(ctx context.Context, req *frontendpb.CreateFunctionRequest) (*frontendpb.CreateFunctionResponse, error) {
	res, err := a.frontendOperator.CreateFunction(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create function: %w", err)
	}
	return res, nil
}

func (a *apiServer) GetFunction(ctx context.Context, req *frontendpb.GetFunctionRequest) (*frontendpb.GetFunctionResponse, error) {
	res, err := a.frontendOperator.GetFunction(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get function: %w", err)
	}
	return res, nil
}

func (a *apiServer) ListFunctions(ctx context.Context, req *frontendpb.ListFunctionsRequest) (*frontendpb.ListFunctionsResponse, error) {
	res, err := a.frontendOperator.ListFunctions(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list functions: %w", err)
	}
	return res, nil
}

func (a *apiServer) GetFunctionCodeUploadUrl(ctx context.Context, req *frontendpb.FunctionCodeUploadUrlRequest) (*frontendpb.FunctionCodeUploadUrlResponse, error) {
	res, err := a.frontendOperator.UpdateFunctionCode(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update function code: %w", err)
	}
	return res, nil
}

func (a *apiServer) UpdateFunctionRuntime(ctx context.Context, req *frontendpb.UpdateFunctionRuntimeRequest) (*sharedpb.EmptyResponse, error) {
	if err := a.frontendOperator.UpdateFunctionRuntime(ctx, req); err != nil {
		return nil, fmt.Errorf("failed to update function runtime: %w", err)
	}
	return &sharedpb.EmptyResponse{}, nil
}

func (a *apiServer) UpdateFuntionResources(ctx context.Context, req *frontendpb.UpdateFunctionResourcesRequest) (*sharedpb.EmptyResponse, error) {
	if err := a.frontendOperator.UpdateFunctionResources(ctx, req); err != nil {
		return nil, fmt.Errorf("failed to update function resources: %w", err)
	}
	return &sharedpb.EmptyResponse{}, nil
}

func (a *apiServer) DeleteFunction(ctx context.Context, req *frontendpb.DeleteFunctionRequest) (*sharedpb.EmptyResponse, error) {
	if err := a.frontendOperator.DeleteFunction(ctx, req); err != nil {
		return nil, fmt.Errorf("failed to delete function: %w", err)
	}
	return &sharedpb.EmptyResponse{}, nil
}
