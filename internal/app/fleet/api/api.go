package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/dennishilgert/apollo/internal/app/fleet/initializer"
	"github.com/dennishilgert/apollo/internal/app/fleet/messaging/producer"
	"github.com/dennishilgert/apollo/internal/app/fleet/operator"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
	"github.com/dennishilgert/apollo/pkg/health"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/proto/fleet/v1"
	"github.com/dennishilgert/apollo/pkg/proto/messages/v1"
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

	runnerOperator    operator.RunnerOperator
	runnerInitializer initializer.RunnerInitializer
	messagingProducer producer.MessagingProducer
	port              int
	readyCh           chan struct{}
	appCtx            context.Context
	running           atomic.Bool
}

// NewApiServer creates a new Server.
func NewApiServer(
	runnerOperator operator.RunnerOperator,
	runnerInitializer initializer.RunnerInitializer,
	messagingProducer producer.MessagingProducer,
	opts Options,
) Server {
	return &apiServer{
		runnerOperator:    runnerOperator,
		runnerInitializer: runnerInitializer,
		messagingProducer: messagingProducer,
		port:              opts.Port,
		readyCh:           make(chan struct{}),
	}
}

// Run runs the api server.
func (a *apiServer) Run(ctx context.Context, healthStatusProvider health.Provider) error {
	if !a.running.CompareAndSwap(false, true) {
		return errors.New("api server is already running")
	}

	// Assign the application context
	a.appCtx = ctx

	log.Infof("starting api server on port %d", a.port)
	server := grpc.NewServer()
	fleet.RegisterFleetManagerServer(server, a)

	healthServer := health.NewHealthServer(healthStatusProvider, log)
	healthServer.Register(server)

	lis, lErr := net.Listen("tcp", ":"+fmt.Sprint(a.port))
	if lErr != nil {
		log.Error("error while starting tcp listener")
		return lErr
	}
	// close the ready channel to signalize that the api server is ready
	close(a.readyCh)

	errCh := make(chan error, 1)
	go func() {
		defer close(errCh) // ensure channel is closed to avoid goroutine leak

		if err := server.Serve(lis); err != nil {
			log.Error("error while serving api server")
			errCh <- err
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
			log.Errorf("error while listening for requests")
		}
	}

	// perform graceful shutdown and close the listener regardless of the select outcome
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

func (a *apiServer) Initialize(ctx context.Context, req *fleet.InitializeFunctionRequest) (*shared.EmptyResponse, error) {
	// handle preparation request asynchronous and respond immediately
	go func() {
		bgCtx, cancel := context.WithTimeout(a.appCtx, time.Minute*10)
		defer cancel()

		if err := a.runnerInitializer.InitializeFunction(bgCtx, req); err != nil {
			log.Errorf("failed to prepare function: %v", err)
			message := &messages.FunctionInitializationFailed{
				FunctionUuid: req.FunctionUuid,
				WorkerUuid:   "TO_BE_REPLACED_WITH_WORKER_UUID",
				Reason:       err.Error(),
			}
			a.messagingProducer.Publish(bgCtx, naming.MessagingFunctionStatusUpdateTopic, message)
		}
		log.Infof("function has been prepared successfully: %s", req.FunctionUuid)
		message := &messages.FunctionInitialized{
			FunctionUuid: req.FunctionUuid,
			WorkerUuid:   "TO_BE_REPLACED_WITH_WORKER_UUID",
		}
		a.messagingProducer.Publish(bgCtx, naming.MessagingFunctionStatusUpdateTopic, message)
	}()
	return &shared.EmptyResponse{}, nil
}

func (a *apiServer) Provision(ctx context.Context, req *fleet.ProvisionRunnerRequest) (*fleet.ProvisionRunnerResponse, error) {
	result, err := a.runnerOperator.ProvisionRunner(req)
	if err != nil {
		log.Error("failed to provision runner")
		return nil, err
	}
	return result, nil
}

func (a *apiServer) Invoke(ctx context.Context, req *fleet.InvokeFunctionRequest) (*fleet.InvokeFunctionResponse, error) {
	result, err := a.runnerOperator.InvokeFunction(ctx, req)
	if err != nil {
		log.Error("failed to invoke function")
		return nil, err
	}
	return result, nil
}
