package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	function "apollo/agent/pkg/function"
	handler "apollo/agent/pkg/function/invoker"
	grpc "apollo/agent/pkg/grpc"

	"github.com/hashicorp/go-hclog"
	"github.com/labstack/echo/v4"
)

const VERSION = "v1.0.0"
const timeFormat = "02-01-2006 15:04:05.000"

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:            "main",
		Level:           hclog.Debug,
		Color:           hclog.AutoColor,
		ColorHeaderOnly: true,
		TimeFormat:      timeFormat,
	})

	logger.Info("apollo agent - " + VERSION)

	// ========== temporary section

	e := echo.New()
	e.HideBanner = true

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello World!")
	})

	e.GET("/shutdown", func(c echo.Context) error {
		logger.Info("received shutdown request")
		shutdown(logger)
		return c.String(http.StatusOK, "Shutting down now")
	})

	e.GET("/run", func(c echo.Context) error {
		logger.Info("received run request")

		response := map[string]interface{}{}

		timestamp := time.Now()

		fnCtx, fnCtxCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer fnCtxCancel()

		result, logs, err := handler.Invoke(fnCtx, logger.Named("invoker"), function.Config{Runtime: "node", RuntimeVersion: "20"}, function.Context{Handler: "index.handler"}, function.Event{})

		duration := fmt.Sprintf("%dms", time.Since(timestamp).Milliseconds())
		response["duration"] = duration
		response["logs"] = logs

		if err != nil {
			response["message"] = "error"
			response["error"] = err.Error()
			return c.JSONPretty(http.StatusInternalServerError, response, "  ")
		}

		response["message"] = "ok"
		response["result"] = result

		return c.JSONPretty(http.StatusOK, response, "  ")
	})

	errorChan := make(chan error, 1)

	go func() {
		if err := e.Start(":3000"); err != nil {
			logger.Error("failed to serve http server", "reason", err)
			errorChan <- err
		}
		close(errorChan)
	}()

	logger.Named("http").Info("http server is listening on :3000")

	// ==========

	grpcLogger := logger.Named("grpc")

	grpcLogger.Info("starting grpc server ...")
	grpcServer, err := grpc.NewGrpcServer(grpcLogger)
	if err != nil {
		grpcLogger.Error("error while creating grpc server")
		shutdown(logger)
	}
	defer grpcServer.GracefulStop()
	grpcLogger.Info("[OK] grpc server started successfully")

	// sending a grpc request to the local server
	// grpcLogger.Info("sending grpc execute request ...")
	// res, err := grpc.Execute(context.Background(), grpcLogger)
	// if err != nil {
	// 	grpcLogger.Error("error while sending execute request")
	// 	shutdown(logger)
	// }
	// grpcLogger.Info("response", "body", res.Body)
	// grpcLogger.Info("[OK] grpc execute request sent successfully")

	doneChan := make(chan os.Signal, 1)
	signal.Notify(doneChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	logger.Info("agent is up and running - waiting for requests")

	<-doneChan

	logger.Info("received signal to shutdown")
	shutdown(logger)
}

func shutdown(logger hclog.Logger) {
	logger.Info("shutting down now")

	err := syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
	if err != nil {
		logger.Error("failed to execute the reboot command", "reason", err)
	}
}
