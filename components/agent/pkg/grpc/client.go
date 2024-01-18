package grpc

import (
	agent "apollo/proto/go/agent/v1"
	"context"

	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type AgentServiceClient struct{}

var (
	agentGrpcClient agent.AgentServiceClient
)

func NewGrpcClient(ctx context.Context, logger hclog.Logger) error {
	serverAddress := ":8080"
	conn, err := grpc.DialContext(ctx, serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("failed to connect to server", "server-address", serverAddress, "reason", err)
		agentGrpcClient = nil
		return err
	}
	if agentGrpcClient != nil {
		conn.Close()
		return nil
	}
	agentGrpcClient = agent.NewAgentServiceClient(conn)
	return nil
}

func Execute(ctx context.Context, logger hclog.Logger) (*agent.FunctionResponse, error) {
	if err := NewGrpcClient(ctx, logger); err != nil {
		logger.Error("failed to create grpc client")
		return nil, err
	}
	body, _ := structpb.NewStruct(map[string]interface{}{
		"content": "important-data",
	})
	req := &agent.FunctionRequest{
		Id:   "86e1414c-ceb9-4b31-a33d-52525f2c1f59",
		Type: "http",
		Data: &agent.FunctionRequest_HttpData{
			HttpData: &agent.HTTPRequestData{
				Method:      "GET",
				Path:        "/data/migrate",
				SourceIp:    "127.0.0.1",
				Headers:     map[string]string{},
				QueryParams: map[string]string{},
				Body:        body,
			},
		},
	}
	res, err := agentGrpcClient.Invoke(ctx, req)
	if err != nil {
		logger.Error("error while execute request", "reason", status.Convert(err).Message())
		return nil, err
	}

	return res, nil
}
