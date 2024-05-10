package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/dennishilgert/apollo/internal/app/fleet/initializer"
	"github.com/dennishilgert/apollo/internal/app/fleet/operator"
	"github.com/dennishilgert/apollo/internal/pkg/health"
	"github.com/dennishilgert/apollo/internal/pkg/logger"
	"github.com/dennishilgert/apollo/internal/pkg/messaging/producer"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
	fleetpb "github.com/dennishilgert/apollo/internal/pkg/proto/fleet/v1"
	messagespb "github.com/dennishilgert/apollo/internal/pkg/proto/messages/v1"
	sharedpb "github.com/dennishilgert/apollo/internal/pkg/proto/shared/v1"
	"google.golang.org/grpc"
)

var log = logger.NewLogger("apollo.manager.api")

type Options struct {
	Port       int
	WorkerUuid string
}

type Server interface {
	Run(ctx context.Context, healthStatusProvider health.Provider) error
	Ready(ctx context.Context) error
}

type apiServer struct {
	fleetpb.UnimplementedFleetManagerServer

	runnerOperator    operator.RunnerOperator
	runnerInitializer initializer.RunnerInitializer
	messagingProducer producer.MessagingProducer
	port              int
	workerUuid        string
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
		workerUuid:        opts.WorkerUuid,
		readyCh:           make(chan struct{}),
	}
}

// Run runs the api server.
func (a *apiServer) Run(ctx context.Context, healthStatusProvider health.Provider) error {
	if !a.running.CompareAndSwap(false, true) {
		return errors.New("api server is already running")
	}

	// Assign the application context.
	a.appCtx = ctx

	log.Infof("starting api server on port %d", a.port)
	server := grpc.NewServer()
	fleetpb.RegisterFleetManagerServer(server, a)

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

// InitializeFunction initializes a function.
func (a *apiServer) InitializeFunction(ctx context.Context, req *fleetpb.InitializeFunctionRequest) (*sharedpb.EmptyResponse, error) {
	// Handle preparation request asynchronous and respond immediately.
	go func() {
		bgCtx, cancel := context.WithTimeout(a.appCtx, time.Minute*10)
		defer cancel()

		var message messagespb.FunctionInitializationResponseMessage
		if err := a.runnerInitializer.InitializeFunction(bgCtx, req); err != nil {
			log.Errorf("failed to initialize function: %v", err)
			message = messagespb.FunctionInitializationResponseMessage{
				Function:   req.Function,
				WorkerUuid: a.workerUuid,
				Reason:     err.Error(),
				Success:    false,
			}
		} else {
			log.Infof("function has been initialized successfully: %s", req.Function.Uuid)
			message = messagespb.FunctionInitializationResponseMessage{
				Function:   req.Function,
				WorkerUuid: a.workerUuid,
				Reason:     "ok",
				Success:    true,
			}
		}
		a.messagingProducer.Publish(bgCtx, naming.MessagingFunctionInitializationResponsesTopic, &message)
	}()
	return &sharedpb.EmptyResponse{}, nil
}

// DeinitializeFunction deinitializes a function.
func (a *apiServer) DeinitializeFunction(ctx context.Context, req *fleetpb.DeinitializeFunctionRequest) (*sharedpb.EmptyResponse, error) {
	// Handle deinitialization request asynchronous and respond immediately.
	go func() {
		bgCtx, cancel := context.WithTimeout(a.appCtx, time.Minute*10)
		defer cancel()

		functionIdentifier := naming.FunctionIdentifier(req.Function.Uuid, req.Function.Version)

		a.runnerOperator.TeardownRunnersByFunction(bgCtx, functionIdentifier)

		var message messagespb.FunctionDeinitializationResponseMessage
		if err := a.runnerInitializer.DeinitializeFunction(bgCtx, req); err != nil {
			log.Errorf("failed to deinitialize function: %v", err)
			message = messagespb.FunctionDeinitializationResponseMessage{
				Function:   req.Function,
				WorkerUuid: a.workerUuid,
				Reason:     err.Error(),
				Success:    false,
			}
		} else {
			log.Infof("function has been deinitialized successfully: %s", req.Function.Uuid)
			message = messagespb.FunctionDeinitializationResponseMessage{
				Function:   req.Function,
				WorkerUuid: a.workerUuid,
				Reason:     "ok",
				Success:    true,
			}
		}
		a.messagingProducer.Publish(bgCtx, naming.MessagingFunctionDeinitializationResponsesTopic, &message)
	}()
	return &sharedpb.EmptyResponse{}, nil
}

// ProvisionRunner provisions a runner.
func (a *apiServer) ProvisionRunner(ctx context.Context, req *fleetpb.ProvisionRunnerRequest) (*fleetpb.ProvisionRunnerResponse, error) {
	result, err := a.runnerOperator.ProvisionRunner(req)
	if err != nil {
		return nil, fmt.Errorf("failed to provision runner: %w", err)
	}
	return result, nil
}

// AvailableRunner checks if a runner for a given function is available.
func (a *apiServer) AvailableRunner(ctx context.Context, req *fleetpb.AvailableRunnerRequest) (*fleetpb.AvailableRunnerResponse, error) {
	result, err := a.runnerOperator.AvailableRunner(req)
	if err != nil {
		return nil, fmt.Errorf("failed to check runner availability: %w", err)
	}
	return result, nil
}

// InvokeFunction invokes a function.
func (a *apiServer) InvokeFunction(ctx context.Context, req *fleetpb.InvokeFunctionRequest) (*fleetpb.InvokeFunctionResponse, error) {
	result, err := a.runnerOperator.InvokeFunction(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke function: %w", err)
	}
	return result, nil
}
