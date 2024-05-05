package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/dennishilgert/apollo/internal/app/agent/runtime"
	"github.com/dennishilgert/apollo/pkg/health"
	"github.com/dennishilgert/apollo/pkg/logger"
	agentpb "github.com/dennishilgert/apollo/pkg/proto/agent/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

var log = logger.NewLogger("apollo.agent.api")

type Options struct {
	Port int
}

type Server interface {
	Run(ctx context.Context, healthStatusProvider health.Provider) error
	Ready(ctx context.Context) error
}

type apiServer struct {
	agentpb.UnimplementedAgentServer

	port              int
	readyCh           chan struct{}
	running           atomic.Bool
	persistentRuntime runtime.PersistentRuntime
}

// NewApiServer creates a new Server.
func NewApiServer(persistentRuntime runtime.PersistentRuntime, opts Options) Server {
	return &apiServer{
		port:              opts.Port,
		readyCh:           make(chan struct{}),
		persistentRuntime: persistentRuntime,
	}
}

// Run runs the api server.
func (a *apiServer) Run(ctx context.Context, healthStatusProvider health.Provider) error {
	if !a.running.CompareAndSwap(false, true) {
		return fmt.Errorf("api server already running")
	}

	s := grpc.NewServer()
	agentpb.RegisterAgentServer(s, a)

	healthServer := health.NewHealthServer(healthStatusProvider, log)
	healthServer.Register(s)

	lis, err := net.Listen("tcp", ":"+fmt.Sprint(a.port))
	if err != nil {
		return fmt.Errorf("starting tcp listener failed: %w", err)
	}
	close(a.readyCh)

	errCh := make(chan error, 1)
	go func() {
		defer close(errCh) // Ensure channel is closed to avoid goroutine leak.

		if err := s.Serve(lis); err != nil {
			errCh <- fmt.Errorf("serving api server failed: %w", err)
			return
		}
		errCh <- nil
	}()

	// Block until the context is done.
	<-ctx.Done()

	s.GracefulStop()
	err = <-errCh
	if err != nil {
		return err
	}
	err = lis.Close()
	if err != nil && !errors.Is(err, net.ErrClosed) {
		return fmt.Errorf("closing tcp listener failed: %w", err)
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

// Invoke invokes a function.
func (a *apiServer) Invoke(ctx context.Context, in *agentpb.InvokeRequest) (*agentpb.InvokeResponse, error) {
	fnCtx := runtime.Context{
		Runtime:        in.Context.Runtime,
		RuntimeVersion: in.Context.RuntimeVersion,
		RuntimeHandler: in.Context.RuntimeHandler,
		MemoryLimit:    in.Context.MemoryLimit,
		VCpuCores:      in.Context.VCpuCores,
	}
	fnEvt := runtime.Event{
		EventUuid: in.Event.Uuid,
		EventType: in.Event.Type,
		Payload:   in.Event.Payload,
	}

	log.Debugf("invoking function with event: %s", in.Event.Uuid)
	result, err := a.persistentRuntime.Invoke(ctx, fnCtx, fnEvt)
	if err != nil {
		return nil, fmt.Errorf("function invocation failed: %w", err)
	}

	logs, err := runtime.LogsToStructList(result.Logs)
	if err != nil {
		return nil, fmt.Errorf("log lines conversion failed: %w", err)
	}
	data, err := structpb.NewStruct(result.Data)
	if err != nil {
		return nil, fmt.Errorf("result data conversion failed: %w", err)
	}

	response := &agentpb.InvokeResponse{
		EventUuid:     result.EventUuid,
		Status:        int32(result.Status),
		StatusMessage: result.StatusMessage,
		Duration:      result.Duration,
		Logs:          logs,
		Data:          data,
	}

	return response, nil
}
