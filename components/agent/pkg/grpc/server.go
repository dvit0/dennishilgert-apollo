package grpc

import (
	agent "apollo/proto/go/agent/v1"
	"context"
	"net"

	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type server struct {
	agent.UnimplementedAgentServiceServer
}

func NewGrpcServer(logger hclog.Logger) (*grpc.Server, error) {
	protocol := "tcp"
	address := "0.0.0.0:8080"
	listen, err := net.Listen(protocol, address)
	if err != nil {
		logger.Error("failed to create listener", "protocol", protocol, "address", address, "reason", err)
		return nil, err
	}
	grpcServer := grpc.NewServer()
	agent.RegisterAgentServiceServer(grpcServer, &server{})

	errorChan := make(chan error)

	go func() {
		if err := grpcServer.Serve(listen); err != nil {
			logger.Error("failed to serve grpc server", "reason", err)
			errorChan <- err
		}
		close(errorChan)
	}()

	select {
	case err := <-errorChan:
		return nil, err

	default:
		return grpcServer, nil
	}
}

func (s *server) Execute(ctx context.Context, req *agent.FunctionRequest) (*agent.FunctionResponse, error) {
	data, _ := structpb.NewStruct(map[string]interface{}{
		"content": "Hello World!",
	})
	return &agent.FunctionResponse{
		Id:             req.Id,
		Status:         200,
		StatusMessage:  "ok",
		Data:           data,
		InvocationTime: timestamppb.Now(),
		CompletionTime: timestamppb.Now(),
	}, nil
}
