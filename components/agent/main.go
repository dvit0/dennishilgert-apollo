package main

import (
	kernel "apollo/agent/pkg/kernel"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	grpc "apollo/agent/pkg/grpc"
	net "apollo/agent/pkg/network"

	"github.com/hashicorp/go-hclog"
	"github.com/labstack/echo/v4"
)

const timeFormat = "02-01-2006 15:04:05.000"

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:            "main",
		Level:           hclog.Debug,
		Color:           hclog.AutoColor,
		ColorHeaderOnly: true,
		TimeFormat:      timeFormat,
	})

	logger.Info("running mount | grep proc")
	out, err := exec.Command("/bin/mount", "|", "grep", "proc").Output()
	if err != nil {
		logger.Error("failed to execute mount | grep proc command", "reason", err)
	}
	logger.Info("Output", string(out))

	logger.Info("running mount -t proc proc /proc")
	out, err = exec.Command("/bin/mount", "-t", "proc", "proc", "/proc").Output()
	if err != nil {
		logger.Error("failed to execute mount -t proc proc /proc command", "reason", err)
	}
	logger.Info("Output", string(out))

	logger.Info("running ls -al /proc")
	out, err = exec.Command("/bin/ls", "-al", "/proc").Output()
	if err != nil {
		logger.Error("failed to execute ls -al /proc command", "reason", err)
	}
	logger.Info("Output", string(out))

	logger.Info("reading kernel command line ...")
	kernelArgs, err := kernel.ReadCmdLine(logger.Named("kernel"))
	if err != nil {
		logger.Error("failed to read arguments from kernel command line")
		shutdown(logger)
		return
	}
	logger.Info("[OK] kernel command line arguments read successfully")

	logger.Debug("network configuration to apply", "net-conf", *kernelArgs.Ip)
	logger.Info("applying network configuration ...")
	if err := net.Apply(kernelArgs.Ip); err != nil {
		log.Printf("failed to apply network configuration from kernel cmdline: %v", err)
		shutdown(logger)
		return
	}
	logger.Info("[OK] network configuration applied successfully")

	logger.Info("running ip a")
	out, err = exec.Command("/bin/ip", "a").Output()
	if err != nil {
		logger.Error("failed to execute ip a command", "reason", err)
	}
	logger.Info("Output", string(out))

	logger.Info("running ip route")
	out, err = exec.Command("/bin/ip", "route").Output()
	if err != nil {
		logger.Error("failed to execute ip route command", "reason", err)
	}
	logger.Info("Output", string(out))

	// ==========

	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello World!")
	})

	e.GET("/shutdown", func(c echo.Context) error {
		logger.Info("received shutdown request")
		shutdown(logger)
		return c.String(http.StatusOK, "Shutting down now")
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
	// cmd := exec.Command("/sbin/reboot")
	// var stdOut bytes.Buffer
	// var stdErr bytes.Buffer
	// cmd.Stdout = &stdOut
	// cmd.Stderr = &stdErr
	// if err := cmd.Run(); err != nil {
	// 	logger.Error("failed to execute the reboot command", "reason", err, "stack", stdErr.String())
	// }
	// logger.Info("Output", stdOut.String())
}
