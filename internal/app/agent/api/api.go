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
	"github.com/dennishilgert/apollo/pkg/proto/agent/v1"
	"github.com/dennishilgert/apollo/pkg/proto/shared/v1"
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
	agent.UnimplementedAgentServer

	port              int
	readyCh           chan struct{}
	running           atomic.Bool
	persistentRuntime runtime.PersistentRuntime
}

func NewApiServer(opts Options) Server {
	return &apiServer{
		port:    opts.Port,
		readyCh: make(chan struct{}),
	}
}

func (a *apiServer) Run(ctx context.Context, healthStatusProvider health.Provider) error {
	if !a.running.CompareAndSwap(false, true) {
		return errors.New("api server is already running")
	}

	log.Infof("starting api server on port %d", a.port)

	s := grpc.NewServer()
	agent.RegisterAgentServer(s, a)

	healthServer := health.NewHealthServer(healthStatusProvider, log)
	healthServer.Register(s)

	lis, err := net.Listen("tcp", ":"+fmt.Sprint(a.port))
	if err != nil {
		return fmt.Errorf("error while starting tcp listener: %w", err)
	}
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

	s.GracefulStop()
	err = <-errCh
	if err != nil {
		return err
	}
	err = lis.Close()
	if err != nil && !errors.Is(err, net.ErrClosed) {
		return fmt.Errorf("error while closing listener: %w", err)
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

func (a *apiServer) Initialize(ctx context.Context, req *agent.InitializeRuntimeRequest) (*shared.EmptyResponse, error) {
	persistentRuntime, err := runtime.NewPersistentRuntime(ctx, runtime.Config{
		BinaryPath: req.BinaryPath,
		BinaryArgs: req.BinaryArgs,
	})
	if err != nil {
		return nil, err
	}
	if err := persistentRuntime.Initialize(req.Handler); err != nil {
		return nil, err
	}

	a.persistentRuntime = persistentRuntime

	return &shared.EmptyResponse{}, nil
}

func (a *apiServer) Invoke(ctx context.Context, in *agent.InvokeRequest) (*agent.InvokeResponse, error) {
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
		Data:      in.Event.Data,
	}

	resultCh := make(chan *runtime.Result, 1)
	errCh := make(chan error, 1)
	go func() {
		// ensure the channels are closed to avoid goroutine leak
		defer func() {
			close(errCh)
			close(resultCh)
		}()

		result, err := a.persistentRuntime.Invoke(ctx, fnCtx, fnEvt)
		if err != nil {
			errCh <- err
			return
		}
		errCh <- nil
		resultCh <- result
	}()

	// block until the error channel receives a message
	err := <-errCh
	if err != nil {
		return nil, err
	}

	// because the error channel received a nil message, read result from result channel
	result := <-resultCh

	logs, err := runtime.LogsToStructList(result.Logs)
	if err != nil {
		log.Fatalf("failed to convert result log lines to struct list: %v", err)
		return nil, err
	}
	data, err := structpb.NewStruct(result.Data)
	if err != nil {
		log.Fatalf("failed to convert result data to struct: %v", err)
		return nil, err
	}

	response := &agent.InvokeResponse{
		EventUuid:     result.EventUuid,
		Status:        int32(result.Status),
		StatusMessage: result.StatusMessage,
		Duration:      result.Duration,
		Logs:          logs,
		Data:          data,
	}

	return response, nil
}
