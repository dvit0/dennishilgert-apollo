package grpc

import (
	agent "apollo/proto/go/agent/v1"
	"context"

	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AgentServiceClient struct{}

var (
	agentGrpcClient agent.AgentServiceClient
)

func NewGrpcClient(ctx context.Context, logger hclog.Logger) error {
	serverAddress := "0.0.0.0:8080"
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

func Execute(ctx context.Context, logger hclog.Logger) (*agent.ExecutionResponse, error) {
	if err := NewGrpcClient(ctx, logger); err != nil {
		logger.Error("failed to create grpc client")
		return nil, err
	}
	body, _ := structpb.NewStruct(map[string]interface{}{
		"message": "okay",
	})
	req := &agent.ExecutionRequest{
		RawPath:               "",
		RawQueryString:        "",
		Cookies:               []string{},
		Headers:               map[string]string{},
		QueryStringParameters: map[string]string{},
		PathParameters:        map[string]string{},
		RequestContext: &agent.RequestContext{
			RequestId:    "abcde-fghijkl-mnopqr-stuvw-xyz",
			DomainName:   "",
			DomainPrefix: "",
			Http: &agent.RequestContextHttp{
				Method:    "GET",
				Path:      "/path",
				Protocol:  "tcp",
				SourceIp:  "127.0.0.1",
				UserAgent: "chrome",
			},
			Timestamp: &timestamppb.Timestamp{
				Nanos: timestamppb.Now().GetNanos(),
			},
		},
		Body: body,
	}
	res, err := agentGrpcClient.Execute(ctx, req)
	if err != nil {
		logger.Error("error while execute request", "reason", status.Convert(err).Message())
		return nil, err
	}

	return res, nil
}
