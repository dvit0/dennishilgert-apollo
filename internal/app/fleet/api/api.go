package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/dennishilgert/apollo/internal/app/fleet/operator"
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

	vmOperator operator.Operator
	port       int
	readyCh    chan struct{}
	running    atomic.Bool
}

// NewApiServer creates a new Server.
func NewApiServer(vmOperator operator.Operator, opts Options) Server {
	return &apiServer{
		vmOperator: vmOperator,
		port:       opts.Port,
		readyCh:    make(chan struct{}),
	}
}

// Run runs the api server.
func (a *apiServer) Run(ctx context.Context, healthStatusProvider health.Provider) error {
	if !a.running.CompareAndSwap(false, true) {
		return errors.New("api server is already running")
	}

	log.Infof("starting api server on port %d", a.port)

	s := grpc.NewServer()
	fleet.RegisterFleetManagerServer(s, a)

	healthServer := health.NewHealthServer(healthStatusProvider, log)
	healthServer.Register(s)

	lis, err := net.Listen("tcp", ":"+fmt.Sprint(a.port))
	if err != nil {
		return fmt.Errorf("error while starting tcp listener: %w", err)
	}
	// close the ready channel to signalize that the api server is ready
	close(a.readyCh)

	errCh := make(chan error, 1)
	go func() {
		defer close(errCh) // ensure channel is closed to avoid goroutine leak

		if err := s.Serve(lis); err != nil {
			errCh <- fmt.Errorf("error while serving api server: %w", err)
			return
		}
		errCh <- nil
	}()

	// block until the context is done
	<-ctx.Done()

	log.Info("shutting down api server")
	s.GracefulStop()
	err = <-errCh
	if err != nil {
		return err
	}
	err = lis.Close()
	if err != nil && !errors.Is(err, net.ErrClosed) {
		return fmt.Errorf("error while closing api server listener: %w", err)
	}

	return nil
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
	return nil, fmt.Errorf("to be implemented")
}

func (a *apiServer) Execute(ctx context.Context, in *fleet.ExecuteFunctionRequest) (*fleet.ExecuteFunctionResponse, error) {
	result, err := a.vmOperator.ExecuteFunction(ctx, in)
	if err != nil {
		return nil, fmt.Errorf("failed to execute function: %v", err)
	}
	return result, nil
}
