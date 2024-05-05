package config

import (
	"github.com/dennishilgert/apollo/pkg/configuration"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/spf13/viper"
)

var log = logger.NewLogger("apollo.agent.config")

type Config struct {
	WorkerUuid         string
	FunctionIdentifier string
	RunnerUuid         string
	RuntimeHandler     string
	RuntimeBinaryPath  string
	RuntimeBinaryArgs  string

	ApiPort                   int
	MessagingBootstrapServers string
}

// Load loads the configuration from the environment.
func Load() (*Config, error) {
	var config Config

	// Automatically load environment variables that match.
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APOLLO")

	// Loading the application specific values from the environment.
	configuration.LoadOrDefault("WorkerUuid", "APOLLO_WORKER_UUID", nil)
	configuration.LoadOrDefault("FunctionIdentifier", "APOLLO_FUNCTION_IDENTIFIER", nil)
	configuration.LoadOrDefault("RunnerUuid", "APOLLO_RUNNER_UUID", nil)
	configuration.LoadOrDefault("RuntimeHandler", "APOLLO_RUNTIME_HANDLER", nil)
	configuration.LoadOrDefault("RuntimeBinaryPath", "APOLLO_RUNTIME_BINARY_PATH", nil)
	configuration.LoadOrDefault("RuntimeBinaryArgs", "APOLLO_RUNTIME_BINARY_ARGS", nil)

	// Loading the values from the environment or use default values.
	configuration.LoadOrDefault("ApiPort", "APOLLO_API_PORT", 50051)

	configuration.LoadOrDefault("MessagingBootstrapServers", "APOLLO_MESSAGING_BOOTSTRAP_SERVERS", nil)

	// Unmarshalling the Config struct.
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("unable to unmarshal config: %v", err)
		return nil, err
	}

	return &config, nil
}
