package bridge

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dennishilgert/apollo/internal/pkg/clients"
	logspb "github.com/dennishilgert/apollo/internal/pkg/proto/logs/v1"
	registrypb "github.com/dennishilgert/apollo/internal/pkg/proto/registry/v1"
	"github.com/dennishilgert/apollo/internal/pkg/registry"
	"github.com/labstack/echo/v4"
)

type LogsBridge interface {
	GetInvocationLogs(c echo.Context) error
}

type logsBridge struct {
	serviceRegistryClient registry.ServiceRegistryClient
}

func NewLogsBridge(serviceRegistryClient registry.ServiceRegistryClient) LogsBridge {
	return &logsBridge{
		serviceRegistryClient: serviceRegistryClient,
	}
}

func (l *logsBridge) GetInvocationLogs(c echo.Context) error {
	client, err := l.client(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to create logs service client: %w", err)
	}
	defer client.Close()

	functionUuid := c.Param("functionUuid")
	functionVersion := c.Param("functionVersion")
	grpcReq := &logspb.FunctionInvocationLogsRequest{
		FunctionUuid:    functionUuid,
		FunctionVersion: functionVersion,
	}

	res, err := client.FunctionInvocationLogs(c.Request().Context(), grpcReq)
	if err != nil {
		return fmt.Errorf("failed to get invocation logs: %w", err)
	}
	return c.JSON(http.StatusOK, res)
}

func (l *logsBridge) client(ctx context.Context) (clients.LogsServiceClient, error) {
	serviceInstance, err := l.serviceRegistryClient.AvailableServiceInstance(ctx, registrypb.InstanceType_LOGS_SERVICE)
	if err != nil {
		return nil, fmt.Errorf("failed to get available service instance: %w", err)
	}

	client, err := clients.NewLogsServiceClient(ctx, fmt.Sprintf("%s:%d", serviceInstance.Host, serviceInstance.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to create logs service client: %w", err)
	}
	return client, nil
}
