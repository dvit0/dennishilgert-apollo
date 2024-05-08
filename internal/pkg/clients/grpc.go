package clients

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/dennishilgert/apollo/internal/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

var log = logger.NewLogger("apollo.grpc")

type GrpcClient interface {
	ClientConn() *grpc.ClientConn
	EstablishConnection(ctx context.Context) error
	CloseConnection()
}

type grpcClient struct {
	address    string
	clientConn *grpc.ClientConn
}

func NewGrpcClient(address string) GrpcClient {
	return &grpcClient{
		address: address,
	}
}

func (g *grpcClient) ClientConn() *grpc.ClientConn {
	return g.clientConn
}

func (g *grpcClient) EstablishConnection(ctx context.Context) error {
	log.Info("establishing a connection to the service")

	if g.clientConn != nil {
		if g.connectionAlive() {
			log.Debug("connection alive - no need to establish a new connection to the service")
			return nil
		}
		// Try to close the existing connection and ignore the possible error.
		g.clientConn.Close()
	}

	log.Debugf("connecting to service with address: %s", g.address)

	const retrySeconds = 3     // trying to connect for a period of 3 seconds
	const retriesPerSecond = 2 // trying to connect 2 times per second
	for i := 0; i < (retrySeconds * retriesPerSecond); i++ {
		conn, err := grpc.DialContext(ctx, g.address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
		if err == nil {
			g.clientConn = conn
			log.Info("connection to service established")
			return nil
		} else {
			if conn != nil {
				conn.Close()
			}
			log.Errorf("failed to establish connection to service - reason: %v", err)
		}
		// Wait before retrying, but stop if context is done.
		select {
		case <-ctx.Done():
			return fmt.Errorf("context done before connection to service could be established: %w", ctx.Err())
		case <-time.After(time.Duration(math.Round(1000/retriesPerSecond)) * time.Millisecond): // retry delay
			continue
		}
	}
	return fmt.Errorf("failed to establish connection to service after %d seconds", retrySeconds)
}

// CloseConnection closes the grpc client connection.
func (g *grpcClient) CloseConnection() {
	if g.clientConn != nil {
		g.clientConn.Close()
	}
}

// connectionAlive returns if a grpc client connection is still alive.
func (g *grpcClient) connectionAlive() bool {
	if g.clientConn == nil {
		return false
	}
	return (g.clientConn.GetState() == connectivity.Ready) || (g.clientConn.GetState() == connectivity.Idle)
}
