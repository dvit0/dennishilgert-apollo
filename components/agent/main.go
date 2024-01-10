package main

import (
	kernel "apollo/agent/pkg/kernel"
	"context"

	//net "apollo/agent/pkg/network"
	grpc "apollo/agent/pkg/grpc"

	"github.com/hashicorp/go-hclog"
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
	// if err := net.Apply(kernelArgs.Ip); err != nil {
	// 	log.Printf("failed to apply network configuration from kernel cmdline: %v", err)
	// 	shutdown(logger)
	// 	return
	// }
	logger.Info("[OK] network configuration applied successfully")

	grpcLogger := logger.Named("grpc")

	grpcLogger.Info("starting grpc server ...")
	grpcServer, err := grpc.NewGrpcServer(grpcLogger)
	if err != nil {
		grpcLogger.Error("error while creating grpc server")
		return
	}
	defer grpcServer.GracefulStop()
	grpcLogger.Info("[OK] grpc server started successfully")

	grpcLogger.Info("sending grpc execute request ...")
	res, err := grpc.Execute(context.Background(), grpcLogger)
	if err != nil {
		grpcLogger.Error("error while sending execute request")
		return
	}
	grpcLogger.Info("response", "body", res.Body)
	grpcLogger.Info("[OK] grpc execute request sent successfully")
}

func shutdown(logger hclog.Logger) {
	logger.Warn("System reboot not implemented yet")
}
