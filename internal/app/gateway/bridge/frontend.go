package bridge

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/dennishilgert/apollo/internal/app/gateway/bridge/models"
	"github.com/dennishilgert/apollo/internal/pkg/clients"
	frontendpb "github.com/dennishilgert/apollo/internal/pkg/proto/frontend/v1"
	registrypb "github.com/dennishilgert/apollo/internal/pkg/proto/registry/v1"
	"github.com/dennishilgert/apollo/internal/pkg/registry"
	"github.com/labstack/echo/v4"
	"google.golang.org/protobuf/types/known/structpb"
)

type FrontendBridge interface {
	InvokeFunction(c echo.Context) error
	AddKernel(c echo.Context) error
	AddRuntime(c echo.Context) error
	CreateFunction(c echo.Context) error
	FunctionCodeUploadUrl(c echo.Context) error
}

type frontendBridge struct {
	serviceRegistryClient registry.ServiceRegistryClient
}

func NewFrontendBridge(serviceRegistryClient registry.ServiceRegistryClient) FrontendBridge {
	return &frontendBridge{
		serviceRegistryClient: serviceRegistryClient,
	}
}

func (f *frontendBridge) InvokeFunction(c echo.Context) error {
	client, err := f.client(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to create frontend client: %w", err)
	}
	defer client.Close()

	// Extract trigger ID from the hostname or the X-Trigger-ID header.
	url := c.Request().URL
	triggerId, err := ExtractTriggerIdFromHostname(url.Hostname())
	if err != nil {
		triggerId = c.Request().Header.Get("X-Trigger-ID")
		if triggerId == "" {
			return fmt.Errorf("failed to extract trigger ID: %w", err)
		}
	}

	// Extract the headers.
	headers := make(map[string]string)
	for key, values := range c.Request().Header {
		if len(values) == 1 {
			headers[key] = values[0]
		} else if len(values) < 1 {
			headers[key] = ""
		} else if len(values) > 1 {
			headers[key] = strings.Join(values, ",")
		}
	}

	// Extract the query parameters.
	params := make(map[string]string)
	queryParams := c.QueryParams()
	for key, values := range queryParams {
		if len(values) == 1 {
			params[key] = values[0]
		} else if len(values) < 1 {
			params[key] = ""
		} else if len(values) > 1 {
			params[key] = strings.Join(values, ",")
		}
	}

	// Bind the incoming JSON body to the map. Echo automatically handles the JSON parsing.
	payload := new(map[string]interface{})
	if err := BindBody(c, payload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("failed to bind request body: %s", err.Error()))
	}
	payloadStruct, err := structpb.NewStruct(*payload)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("failed to map payload to struct: %s", err.Error()))
	}

	grpcReq := &frontendpb.InvokeFunctionRequest{
		HttpTriggerId: triggerId,
		Event: &frontendpb.EventSpecs{
			Type:     "http-trigger",
			SourceIp: c.RealIP(),
			Headers:  headers,
			Params:   params,
			Payload:  payloadStruct,
		},
	}

	res, err := client.InvokeFunction(c.Request().Context(), grpcReq)
	if err != nil {
		return fmt.Errorf("failed to invoke function: %w", err)
	}
	return c.JSON(http.StatusOK, res)
}

func (f *frontendBridge) AddKernel(c echo.Context) error {
	client, err := f.client(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to create frontend client: %w", err)
	}
	defer client.Close()

	req := new(models.AddKernelRequest)
	if err := BindBody(c, req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("failed to bind request body: %s", err.Error()))
	}
	grpcReq := &frontendpb.AddKernelRequest{
		Kernel: &frontendpb.KernelSpecs{
			Name:         req.Name,
			Version:      req.Version,
			Architecture: req.Architecture,
		},
	}

	res, err := client.AddKernel(c.Request().Context(), grpcReq)
	if err != nil {
		return fmt.Errorf("failed to add kernel: %w", err)
	}
	return c.JSON(http.StatusOK, res)
}

func (f *frontendBridge) AddRuntime(c echo.Context) error {
	client, err := f.client(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to create frontend client: %w", err)
	}
	defer client.Close()

	req := new(models.AddRuntimeRequest)
	if err := BindBody(c, req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("failed to bind request body: %s", err.Error()))
	}
	if req.Environment == nil {
		req.Environment = []string{}
	}
	grpcReq := &frontendpb.AddRuntimeRequest{
		Runtime: &frontendpb.RuntimeSpecs{
			Name:               req.Name,
			Version:            req.Version,
			BinaryPath:         req.BinaryPath,
			BinaryArgs:         req.BinaryArgs,
			Environment:        req.Environment,
			DisplayName:        req.DisplayName,
			DefaultHandler:     req.DefaultHandler,
			DefaultMemoryLimit: req.DefaultMemoryLimit,
			DefaultVCpuCores:   req.DefaultVCpuCores,
			KernelName:         req.KernelName,
			KernelVersion:      req.KernelVersion,
			KernelArchitecture: req.KernelArchitecture,
		},
	}

	res, err := client.AddRuntime(c.Request().Context(), grpcReq)
	if err != nil {
		return fmt.Errorf("failed to add runtime: %w", err)
	}
	return c.JSON(http.StatusOK, res)
}

func (f *frontendBridge) CreateFunction(c echo.Context) error {
	client, err := f.client(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to create frontend client: %w", err)
	}
	defer client.Close()

	req := new(models.CreateFunctionRequest)
	if err := BindBody(c, req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("failed to bind request body: %s", err.Error()))
	}
	grpcReq := &frontendpb.CreateFunctionRequest{
		Name:                req.Name,
		RuntimeName:         req.RuntimeName,
		RuntimeVersion:      req.RuntimeVersion,
		RuntimeArchitecture: req.RuntimeArchitecture,
	}

	res, err := client.CreateFunction(c.Request().Context(), grpcReq)
	if err != nil {
		return fmt.Errorf("failed to create function: %w", err)
	}
	return c.JSON(http.StatusOK, res)
}

func (f *frontendBridge) FunctionCodeUploadUrl(c echo.Context) error {
	client, err := f.client(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to create frontend client: %w", err)
	}
	defer client.Close()

	functionUuid := c.Param("functionUuid")
	grpcReq := &frontendpb.FunctionCodeUploadUrlRequest{
		FunctionUuid: functionUuid,
	}

	res, err := client.FunctionCodeUploadUrl(c.Request().Context(), grpcReq)
	if err != nil {
		return fmt.Errorf("failed to get function code upload url: %w", err)
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
