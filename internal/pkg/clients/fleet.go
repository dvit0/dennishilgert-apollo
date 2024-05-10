package clients

import (
	"context"
	"fmt"

	fleetpb "github.com/dennishilgert/apollo/internal/pkg/proto/fleet/v1"
	sharedpb "github.com/dennishilgert/apollo/internal/pkg/proto/shared/v1"
)

type FleetManagerClient interface {
	Close()
	DeinitializeFunction(ctx context.Context, req *fleetpb.DeinitializeFunctionRequest) (*sharedpb.EmptyResponse, error)
}

type fleetManagerClient struct {
	grpcClient GrpcClient
	client     fleetpb.FleetManagerClient
}

func NewFleetManagerClient(ctx context.Context, address string) (FleetManagerClient, error) {
	grpcClient := NewGrpcClient(address)
	if err := grpcClient.EstablishConnection(ctx); err != nil {
		return nil, fmt.Errorf("failed to establish connection to the service: %w", err)
	}

	client := fleetpb.NewFleetManagerClient(grpcClient.ClientConn())

	return &fleetManagerClient{
		grpcClient: grpcClient,
		client:     client,
	}, nil
}

func (f *fleetManagerClient) Close() {
	f.grpcClient.CloseConnection()
}

func (f *fleetManagerClient) DeinitializeFunction(ctx context.Context, req *fleetpb.DeinitializeFunctionRequest) (*sharedpb.EmptyResponse, error) {
	return f.client.DeinitializeFunction(ctx, req)
}
