package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/dennishilgert/apollo/internal/app/agent/runtime"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/proto/agent/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

var log = logger.NewLogger("apollo.agent.api")

type Options struct {
	Port int
}

type Server interface {
	Run(ctx context.Context) error
	Ready(ctx context.Context) error
}

type apiServer struct {
	agent.UnimplementedAgentServer

	port    int
	readyCh chan struct{}
	running atomic.Bool
}

func NewAPIServer(opts Options) Server {
	return &apiServer{
		port:    opts.Port,
		readyCh: make(chan struct{}),
	}
}

func (a *apiServer) Run(ctx context.Context) error {
	if !a.running.CompareAndSwap(false, true) {
		return errors.New("api server is already running")
	}

	log.Infof("starting api server on port %d", a.port)

	s := grpc.NewServer()
	agent.RegisterAgentServer(s, a)

	lis, err := net.Listen("tcp", ":"+fmt.Sprint(a.port))
	if err != nil {
		return fmt.Errorf("error while starting tcp listener: %w", err)
	}
	close(a.readyCh)

	errCh := make(chan error)
	go func() {
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

func (a *apiServer) Ready(ctx context.Context) error {
	select {
	case <-a.readyCh:
		return nil
	case <-ctx.Done():
		return errors.New("timeout while waiting for the api server to be ready")
	}
}

func (a *apiServer) Invoke(ctx context.Context, in *agent.InvokeRequest) (*agent.InvokeResponse, error) {
	fnCfg := runtime.Config{
		RuntimeBinaryPath: in.Runtime.BinaryPath,
		RuntimeBinaryArgs: in.Runtime.BinaryArgs,
	}
	fnCtx := runtime.Context{
		Runtime:        in.Runtime.Name,
		RuntimeVersion: in.Runtime.Version,
		RuntimeHandler: in.Runtime.Handler,
		MemoryLimit:    in.Function.MemoryLimit,
		VCpuCores:      in.Function.VcpuCores,
	}
	fnEvt := runtime.Event{
		RequestID:   in.Id,
		RequestType: in.Type,
		Data:        in.Data,
	}

	resultCh := make(chan *runtime.Result)
	errCh := make(chan error)
	go func() {
		result, err := runtime.Invoke(ctx, log.WithLogType("info"), fnCfg, fnCtx, fnEvt)
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
		RequestId:     result.RequestID,
		Status:        int32(result.Status),
		StatusMessage: result.StatusMessage,
		Duration:      result.Duration,
		Logs:          logs,
		Data:          data,
	}

	return response, nil
}
