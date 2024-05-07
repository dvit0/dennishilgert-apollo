package bridge

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dennishilgert/apollo/internal/app/gateway/bridge/clients"
	frontendpb "github.com/dennishilgert/apollo/internal/pkg/proto/frontend/v1"
	registrypb "github.com/dennishilgert/apollo/internal/pkg/proto/registry/v1"
	"github.com/dennishilgert/apollo/internal/pkg/registry"
	"github.com/labstack/echo/v4"
)

type FrontendBridge interface {
	CreateFunction(c echo.Context) error
}

type frontendBridge struct {
	serviceRegistryClient registry.ServiceRegistryClient
}

func NewFrontendBridge(serviceRegistryClient registry.ServiceRegistryClient) FrontendBridge {
	return &frontendBridge{
		serviceRegistryClient: serviceRegistryClient,
	}
}

func (f frontendBridge) CreateFunction(c echo.Context) error {
	client, err := f.client(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("failed to create client: %v", err)})
	}
	req := &frontendpb.CreateFunctionRequest{
		Name:                c.FormValue("name"),
		RuntimeName:         c.FormValue("runtime_name"),
		RuntimeVersion:      c.FormValue("runtime_version"),
		RuntimeArchitecture: c.FormValue("runtime_architecture"),
	}
	res, err := client.CreateFunction(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("failed to create function: %v", err)})
	}
	return c.JSON(http.StatusOK, res)
}

func (f *frontendBridge) client(ctx context.Context) (clients.FrontendClient, error) {
	serviceInstance, err := f.serviceRegistryClient.AvailableServiceInstance(ctx, registrypb.InstanceType_FRONTEND)
	if err != nil {
		return nil, fmt.Errorf("failed to get available service instance: %w", err)
	}

	client, err := clients.NewFrontendClient(ctx, fmt.Sprintf("%s:%d", serviceInstance.Host, serviceInstance.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to create frontend client: %w", err)
	}
	return client, nil
}
