package rest

import (
	"github.com/dennishilgert/apollo/internal/app/gateway/bridge"
	"github.com/dennishilgert/apollo/internal/pkg/registry"
	"github.com/labstack/echo/v4"
)

type RestHandler interface {
	RegisterHandlers(e *echo.Echo)
}

type restHandler struct {
	serviceRegistryClient registry.ServiceRegistryClient
	frontendBridge        bridge.FrontendBridge
}

func NewRestHandler(serviceRegistryClient registry.ServiceRegistryClient) RestHandler {
	return &restHandler{
		serviceRegistryClient: serviceRegistryClient,
	}
}

func (r *restHandler) RegisterHandlers(e *echo.Echo) {
	apiV1 := e.Group("/api/v1")

	apiV1.POST("/functions", r.frontendBridge.CreateFunction)
}
