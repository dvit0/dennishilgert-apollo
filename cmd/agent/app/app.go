package app

import (
	"github.com/dennishilgert/apollo/cmd/agent/config"
	"github.com/dennishilgert/apollo/internal/app/agent"
	"github.com/dennishilgert/apollo/internal/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/internal/pkg/logger"
	"github.com/dennishilgert/apollo/internal/pkg/signals"
	"github.com/dennishilgert/apollo/internal/pkg/utils"
	"github.com/joho/godotenv"
)

var log = logger.NewLogger("apollo.agent")

// Run starts the agent.
func Run() {
	// Load environment variables from .env file for local development.
	godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("starting apollo agent -- version %s", "TO_BE_IMPLEMENTED")
	log.Infof("log level set to: %s", log.LogLevel())

	runtimeConfig, err := utils.DeserializeJson(cfg.RuntimeConfiguration)
	if err != nil {
		log.Fatalf("error while deserializing runtime configuration: %v", err)
	}
	runtimeHandler, ok := runtimeConfig["handler"].(string)
	if !ok {
		log.Fatalf("error while reading runtime handler from configuration")
	}
	runtimeBinaryPath, ok := runtimeConfig["binaryPath"].(string)
	if !ok {
		log.Fatalf("error while reading runtime binary path from configuration")
	}
	if rawBinaryArgs, ok := runtimeConfig["binaryArgs"].([]interface{}); ok {
		binaryArgs := make([]string, len(rawBinaryArgs))
		for i, v := range rawBinaryArgs {
			binaryArgs[i] = v.(string)
		}
		runtimeConfig["binaryArgs"] = binaryArgs
	} else {
		log.Fatalf("error while reading runtime binary arguments from configuration")
	}
	if rawEnvironment, ok := runtimeConfig["environment"].([]interface{}); ok {
		environment := make([]string, len(rawEnvironment))
		for i, v := range rawEnvironment {
			environment[i] = v.(string)
		}
		runtimeConfig["environment"] = environment
	} else {
		log.Fatalf("error while reading runtime environment from configuration")
	}

	runtimeBinaryArgs := runtimeConfig["binaryArgs"].([]string)
	runtimeEnvironment := runtimeConfig["environment"].([]string)

	ctx := signals.Context()
	agent, err := agent.NewAgent(
		ctx,
		agent.Options{
			WorkerUuid:                cfg.WorkerUuid,
			FunctionIdentifier:        cfg.FunctionIdentifier,
			RunnerUuid:                cfg.RunnerUuid,
			RuntimeHandler:            runtimeHandler,
			RuntimeBinaryPath:         runtimeBinaryPath,
			RuntimeBinaryArgs:         runtimeBinaryArgs,
			RuntimeEnvironment:        runtimeEnvironment,
			ApiPort:                   cfg.ApiPort,
			MessagingBootstrapServers: cfg.MessagingBootstrapServers,
		},
	)
	if err != nil {
		log.Fatalf("error while creating agent: %v", err)
	}

	err = runner.NewRunnerManager(
		agent.Run,
	).Run(ctx)
	if err != nil {
		log.Fatalf("error while running agent: %v", err)
	}

	log.Info("agent shut down gracefully")
}
