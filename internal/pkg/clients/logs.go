package clients

import (
	"context"
	"fmt"

	logspb "github.com/dennishilgert/apollo/internal/pkg/proto/logs/v1"
)

type LogsServiceClient interface {
	Close()
	FunctionInvocationLogs(ctx context.Context, req *logspb.FunctionInvocationLogsRequest) (*logspb.FunctionInvocationLogsResponse, error)
}

type logsServiceClient struct {
	grpcClient GrpcClient
	client     logspb.LogsServiceClient
}

func NewLogsServiceClient(ctx context.Context, address string) (LogsServiceClient, error) {
	grpcClient := NewGrpcClient(address)
	if err := grpcClient.EstablishConnection(ctx); err != nil {
		return nil, fmt.Errorf("failed to establish connection to the service: %w", err)
	}

	client := logspb.NewLogsServiceClient(grpcClient.ClientConn())

	return &logsServiceClient{
		grpcClient: grpcClient,
		client:     client,
	}, nil
}

func (l *logsServiceClient) Close() {
	l.grpcClient.CloseConnection()
}

func (l *logsServiceClient) FunctionInvocationLogs(ctx context.Context, req *logspb.FunctionInvocationLogsRequest) (*logspb.FunctionInvocationLogsResponse, error) {
	return l.client.FunctionInvocationLogs(ctx, req)
}
