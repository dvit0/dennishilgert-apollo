package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/dennishilgert/apollo/internal/app/agent/runtime"
	"github.com/dennishilgert/apollo/internal/pkg/health"
	"github.com/dennishilgert/apollo/internal/pkg/logger"
	"github.com/dennishilgert/apollo/internal/pkg/messaging/producer"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
	agentpb "github.com/dennishilgert/apollo/internal/pkg/proto/agent/v1"
	logspb "github.com/dennishilgert/apollo/internal/pkg/proto/logs/v1"
	messagespb "github.com/dennishilgert/apollo/internal/pkg/proto/messages/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var log = logger.NewLogger("apollo.agent.api")

type Options struct {
	Port               int
	FunctionIdentifier string
}

type Server interface {
	Run(ctx context.Context, healthStatusProvider health.Provider) error
	Ready(ctx context.Context) error
}

type apiServer struct {
	agentpb.UnimplementedAgentServer

	port               int
	readyCh            chan struct{}
	running            atomic.Bool
	functionIdentifier string
	persistentRuntime  runtime.PersistentRuntime
	messagingProducer  producer.MessagingProducer
}

// NewApiServer creates a new Server.
func NewApiServer(persistentRuntime runtime.PersistentRuntime, messagingProducer producer.MessagingProducer, opts Options) Server {
	return &apiServer{
		port:               opts.Port,
		readyCh:            make(chan struct{}),
		functionIdentifier: opts.FunctionIdentifier,
		persistentRuntime:  persistentRuntime,
		messagingProducer:  messagingProducer,
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
	event := runtime.Event{
		EventUuid: in.Event.Uuid,
		EventType: in.Event.Type,
		SourceIp:  in.Event.SourceIp,
		Headers:   in.Event.Headers,
		Params:    in.Event.Params,
		Payload:   in.Event.Payload,
	}

	log.Debugf("invoking function with event: %s", in.Event.Uuid)
	result, err := a.persistentRuntime.Invoke(ctx, event)
	if err != nil {
		return nil, fmt.Errorf("function invocation failed: %w", err)
	}

	logs := append(result.Logs, &logspb.LogEntry{
		Timestamp:  timestamppb.Now(),
		LogLevel:   "info",
		LogMessage: fmt.Sprintf("billed duration for exection: %dms", result.Duration),
	})

	log.Debugf("publishing function invocation logs")
	a.messagingProducer.Publish(ctx, naming.MessagingFunctionInvocationLogsTopic, messagespb.FunctionInvocationLogsMessage{
		FunctionIdentifier: a.functionIdentifier,
		EventUuid:          result.EventUuid,
		Logs:               logs,
	})

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
