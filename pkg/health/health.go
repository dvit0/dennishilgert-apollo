package health

import (
	"context"

	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/proto/health/v1"
	"google.golang.org/grpc"
)

type Server interface {
	Register(server *grpc.Server)
}

type healthServer struct {
	health.UnimplementedHealthServer

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
	health.RegisterHealthServer(server, h)
}

// Status returns the health status of the instance.
func (h *healthServer) Status(ctx context.Context, in *health.HealthStatusRequest) (*health.HealthStatusResponse, error) {
	status := health.HealthStatus_HEALTHY
	if !h.healthStatusProvider.Healthy() {
		status = health.HealthStatus_UNHEALTHY
	}
	h.log.Debugf("responding to health status request with health status: %s", status.String())
	response := &health.HealthStatusResponse{
		Status: status,
	}
	return response, nil
}
