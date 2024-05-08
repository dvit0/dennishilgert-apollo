package rest

import (
	"context"
	"fmt"
	"time"

	"github.com/dennishilgert/apollo/internal/pkg/registry"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Options struct {
	ApiPort int
}

type RestServer interface {
	Run() error
	Shutdown() error
}

type restServer struct {
	apiPort int
	e       *echo.Echo
}

func NewRestServer(serviceRegistryClient registry.ServiceRegistryClient, opts Options) RestServer {
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())

	restHandler := NewRestHandler(serviceRegistryClient)
	restHandler.RegisterHandlers(e)

	return &restServer{
		apiPort: opts.ApiPort,
		e:       e,
	}
}

func (s *restServer) Run() error {
	if err := s.e.Start(fmt.Sprintf(":%d", s.apiPort)); err != nil {
		return fmt.Errorf("error while starting rest server: %w", err)
	}
	return nil
}

func (s *restServer) Shutdown() error {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer ctxCancel()
	if err := s.e.Shutdown(ctx); err != nil {
		return fmt.Errorf("error while shutting down rest server: %w", err)
	}
	return nil
}
