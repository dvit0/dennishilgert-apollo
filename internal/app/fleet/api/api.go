package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/dennishilgert/apollo/internal/app/fleet/operator"
	"github.com/dennishilgert/apollo/internal/app/fleet/preparer"
	"github.com/dennishilgert/apollo/pkg/health"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/proto/fleet/v1"
	"github.com/dennishilgert/apollo/pkg/proto/shared/v1"
	"google.golang.org/grpc"
)

var log = logger.NewLogger("apollo.manager.api")

type Options struct {
	Port int
}

type Server interface {
	Run(ctx context.Context, healthStatusProvider health.Provider) error
	Ready(ctx context.Context) error
}

type apiServer struct {
	fleet.UnimplementedFleetManagerServer

	runnerOperator operator.RunnerOperator
	runnerPreparer preparer.RunnerPreparer
	port           int
	readyCh        chan struct{}
	running        atomic.Bool
}

// NewApiServer creates a new Server.
func NewApiServer(runnerOperator operator.RunnerOperator, runnerPreparer preparer.RunnerPreparer, opts Options) Server {
	return &apiServer{
		runnerOperator: runnerOperator,
		runnerPreparer: runnerPreparer,
		port:           opts.Port,
		readyCh:        make(chan struct{}),
	}
}

// Run runs the api server.
func (a *apiServer) Run(ctx context.Context, healthStatusProvider health.Provider) error {
	if !a.running.CompareAndSwap(false, true) {
		return errors.New("api server is already running")
	}

	log.Infof("starting api server on port %d", a.port)
	server := grpc.NewServer()
	fleet.RegisterFleetManagerServer(server, a)

	healthServer := health.NewHealthServer(healthStatusProvider, log)
	healthServer.Register(server)

	lis, lErr := net.Listen("tcp", ":"+fmt.Sprint(a.port))
	if lErr != nil {
		return fmt.Errorf("error while starting tcp listener: %w", lErr)
	}
	// close the ready channel to signalize that the api server is ready
	close(a.readyCh)

	errCh := make(chan error, 1)
	go func() {
		defer close(errCh) // ensure channel is closed to avoid goroutine leak

		if err := server.Serve(lis); err != nil {
			errCh <- fmt.Errorf("error while serving api server: %w", err)
			return
		}
		errCh <- nil
	}()

	// block until the context is done or an error occurs
	var serveErr error
	select {
	case <-ctx.Done():
		log.Info("shutting down api server")
	case err := <-errCh: // Handle errors that might have occurred during Serve
		if err != nil {
			serveErr = err
			log.Errorf("error while listening for requests: %v", err)
		}
	}

	// perform graceful shutdown and close the listener regardless of the select outcome
	server.GracefulStop()
	if cErr := lis.Close(); cErr != nil && !errors.Is(cErr, net.ErrClosed) && serveErr == nil {
		return fmt.Errorf("error while closing api server listener: %w", cErr)
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

func (a *apiServer) Initialize(ctx context.Context, in *fleet.InitializeFunctionRequest) (*shared.EmptyResponse, error) {
	if err := a.runnerPreparer.InitializeFunction(ctx, in); err != nil {
		return nil, fmt.Errorf("failed to initialize function: %v", err)
	}
	return &shared.EmptyResponse{}, nil
}

func (a *apiServer) Execute(ctx context.Context, in *fleet.ExecuteFunctionRequest) (*fleet.ExecuteFunctionResponse, error) {
	result, err := a.runnerOperator.ExecuteFunction(ctx, in)
	if err != nil {
		return nil, fmt.Errorf("failed to execute function: %v", err)
	}
	return result, nil
}
