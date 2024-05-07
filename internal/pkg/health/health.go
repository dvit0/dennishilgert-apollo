package health

import (
	"context"

	"github.com/dennishilgert/apollo/internal/pkg/logger"
	healthpb "github.com/dennishilgert/apollo/internal/pkg/proto/health/v1"
	"google.golang.org/grpc"
)

type Server interface {
	Register(server *grpc.Server)
}

type healthServer struct {
	healthpb.UnimplementedHealthServer

	healthStatusProvider Provider
	log                  logger.Logger
}

// NewHealthServer creates a new Server.
func NewHealthServer(provider Provider, logger logger.Logger) Server {
	return &healthServer{
		healthStatusProvider: provider,
		log:                  logger,
	}
}

// Register registers the health server in an grpc server.
func (h *healthServer) Register(server *grpc.Server) {
	healthpb.RegisterHealthServer(server, h)
}

// Status returns the health status of the instance.
func (h *healthServer) Status(ctx context.Context, in *healthpb.HealthStatusRequest) (*healthpb.HealthStatusResponse, error) {
	status := healthpb.HealthStatus_HEALTHY
	if !h.healthStatusProvider.Healthy() {
		status = healthpb.HealthStatus_UNHEALTHY
	}
	h.log.Debugf("responding to health status request with health status: %s", status.String())
	response := &healthpb.HealthStatusResponse{
		Status: status,
	}
	return response, nil
}
