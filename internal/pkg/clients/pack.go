package clients

import (
	"context"
	"fmt"

	packpb "github.com/dennishilgert/apollo/internal/pkg/proto/pack/v1"
)

type PackageServiceClient interface {
	Close()
	PresignedUploadUrl(ctx context.Context, req *packpb.PresignedUploadUrlRequest) (*packpb.PresignedUploadUrlResponse, error)
}

type packageServiceClient struct {
	grpcClient GrpcClient
	client     packpb.PackageServiceClient
}

func NewPackageServiceClient(ctx context.Context, address string) (PackageServiceClient, error) {
	grpcClient := NewGrpcClient(address)
	if err := grpcClient.EstablishConnection(ctx); err != nil {
		return nil, fmt.Errorf("failed to establish connection to the service: %w", err)
	}

	client := packpb.NewPackageServiceClient(grpcClient.ClientConn())

	return &packageServiceClient{
		grpcClient: grpcClient,
		client:     client,
	}, nil
}

func (p *packageServiceClient) Close() {
	p.grpcClient.CloseConnection()
}

func (p *packageServiceClient) PresignedUploadUrl(ctx context.Context, req *packpb.PresignedUploadUrlRequest) (*packpb.PresignedUploadUrlResponse, error) {
	return p.client.PresignedUploadUrl(ctx, req)
}
