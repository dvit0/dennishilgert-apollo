package config

import (
	"github.com/dennishilgert/apollo/pkg/configuration"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/spf13/viper"
)

var log = logger.NewLogger("apollo.agent.config")

type Config struct {
	ApiPort                int
	DataPath               string
	FirecrackerBinaryPath  string
	DockerImageRegistryUrl string
	WatchdogCheckInterval  int
	WatchdogWorkerCount    int
	AgentApiPort           int
	Logger                 logger.Config
}

func Load() (*Config, error) {
	var config Config

	// automatically load environment variables that match
	viper.AutomaticEnv()

	// loading the values from the environment or use default values
	configuration.LoadOrDefault("Logger.AppId", "LOG_APP_ID", logger.DefaultConfig().AppId)
	configuration.LoadOrDefault("Logger.LogJsonOutput", "LOG_FORMAT_JSON", logger.DefaultConfig().LogJsonOutput)
	configuration.LoadOrDefault("Logger.LogLevel", "LOG_LEVEL", logger.DefaultConfig().LogLevel)

	configuration.LoadOrDefault("ApiPort", "API_PORT", 50051)
	configuration.LoadOrDefault("DataPath", "DATA_PATH", "/data")
	configuration.LoadOrDefault("FirecrackerBinaryPath", "FC_BINARY_PATH", nil)
	configuration.LoadOrDefault("DockerImageRegistryUrl", "DOCKER_IMAGE_REGISTRY_URL", nil)
	configuration.LoadOrDefault("WatchdogCheckInterval", "WATCHDOG_CHECK_INTERVAL", 5)
	configuration.LoadOrDefault("WatchdogWorkerCount", "WATCHDOG_WORKER_COUNT", 10)
	configuration.LoadOrDefault("AgentApiPort", "AGENT_API_PORT", 50051)

	// unmarshalling the Config struct
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Unable to unmarshal config: %v", err)
		return nil, err
	}

	return &config, nil
}
