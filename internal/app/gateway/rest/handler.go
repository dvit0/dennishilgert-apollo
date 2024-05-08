package rest

import (
	"errors"
	"net/http"

	"github.com/dennishilgert/apollo/internal/app/gateway/bridge"
	"github.com/dennishilgert/apollo/internal/app/gateway/rest/models"
	"github.com/dennishilgert/apollo/internal/pkg/registry"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type RestHandler interface {
	RegisterHandlers(e *echo.Echo)
}

type restHandler struct {
	serviceRegistryClient registry.ServiceRegistryClient
	frontendBridge        bridge.FrontendBridge
}

// NewRestHandler creates a new RestHandler.
func NewRestHandler(serviceRegistryClient registry.ServiceRegistryClient) RestHandler {
	frontendBridge := bridge.NewFrontendBridge(serviceRegistryClient)

	return &restHandler{
		serviceRegistryClient: serviceRegistryClient,
		frontendBridge:        frontendBridge,
	}
}

// RegisterHandlers registers the REST API handlers.
func (r *restHandler) RegisterHandlers(e *echo.Echo) {
	apiV1 := e.Group("/api/v1")

	apiV1.Use(RequestInterceptor)

	apiV1.GET("/hello", func(c echo.Context) error {
		return c.String(200, "Hello, World!")
	})

	apiV1.POST("/functions", r.frontendBridge.CreateFunction, RequestValidator(func() interface{} {
		return new(models.CreateFunctionRequest)
	}))
}

// RequestInterceptor creates a middleware for handling errors.
func RequestInterceptor(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Execute request handler
		err := next(c)
		if err != nil {
			httpError, ok := err.(*echo.HTTPError)
			if !ok {
				// If it's not an echo.HTTPError, wrap it for consistency
				return c.JSON(http.StatusInternalServerError, map[string]interface{}{
					"status":  "Internal Server Error",
					"message": errors.Unwrap(err).Error(),
					"error":   err.Error(),
				})
			}
			// Use the message from HTTPError to maintain custom error formatting
			return c.JSON(httpError.Code, httpError.Message)
		}
		return nil
	}
}

// RequestValidator creates a middleware for validating requests.
func RequestValidator(factoryFunc func() interface{}) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := factoryFunc()
			// Bind and validate the request.
			if err := validateRequest(c, req); err != nil {
				// Directly return the error to halt processing and not call next.
				return err
			}
			// Proceed to the next handler if validation succeeds.
			return next(c)
		}
	}
}

// validateRequest binds and validates the request.
func validateRequest(c echo.Context, req interface{}) error {
	validate := validator.New()

	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format: "+err.Error())
	}
	if err := validate.Struct(req); err != nil {
		// Check if the errors are ValidationErrors and return them directly.
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			return echo.NewHTTPError(http.StatusBadRequest, validationErrorResponse(validationErrors))
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Error during request validation: "+err.Error())
	}
	return nil
}

// validationErrorResponse formats validation errors for the response.
func validationErrorResponse(err validator.ValidationErrors) map[string]interface{} {
	errorsMap := make(map[string]string)
	for _, e := range err {
		errorsMap[e.Field()] = e.Translate(nil) // Replace `nil` with your translator if configured
	}

	return map[string]interface{}{
		"status":  "Bad Request",
		"message": "Validation Error",
		"errors":  errorsMap,
	}
}
