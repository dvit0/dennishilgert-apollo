package clients

import (
	"context"
	"fmt"

	frontendpb "github.com/dennishilgert/apollo/internal/pkg/proto/frontend/v1"
)

type FrontendClient interface {
	Close()
	CreateFunction(ctx context.Context, req *frontendpb.CreateFunctionRequest) (*frontendpb.CreateFunctionResponse, error)
}

type frontendClient struct {
	grpcClient GrpcClient
	client     frontendpb.FrontendClient
}

func NewFrontendClient(ctx context.Context, address string) (FrontendClient, error) {
	grpcClient := NewGrpcClient(address)
	if err := grpcClient.EstablishConnection(ctx); err != nil {
		return nil, fmt.Errorf("failed to establish connection to the service: %w", err)
	}

	client := frontendpb.NewFrontendClient(grpcClient.ClientConn())

	return &frontendClient{
		grpcClient: grpcClient,
		client:     client,
	}, nil
}

func (f *frontendClient) Close() {
	f.grpcClient.CloseConnection()
}

func (f *frontendClient) CreateFunction(ctx context.Context, req *frontendpb.CreateFunctionRequest) (*frontendpb.CreateFunctionResponse, error) {
	return f.client.CreateFunction(ctx, req)
}
