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
}

func Load() (*Config, error) {
	var config Config

	// automatically load environment variables that match
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APOLLO")

	// loading the values from the environment or use default values
	configuration.LoadOrDefault("ApiPort", "APOLLO_API_PORT", 50051)
	configuration.LoadOrDefault("DataPath", "APOLLO_DATA_PATH", "/data")
	configuration.LoadOrDefault("FirecrackerBinaryPath", "APOLLO_FC_BINARY_PATH", nil)
	configuration.LoadOrDefault("DockerImageRegistryUrl", "APOLLO_DOCKER_IMAGE_REGISTRY_URL", nil)
	configuration.LoadOrDefault("WatchdogCheckInterval", "APOLLO_WATCHDOG_CHECK_INTERVAL", 5)
	configuration.LoadOrDefault("WatchdogWorkerCount", "APOLLO_WATCHDOG_WORKER_COUNT", 10)
	configuration.LoadOrDefault("AgentApiPort", "APOLLO_AGENT_API_PORT", 50051)

	// unmarshalling the Config struct
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("unable to unmarshal config: %v", err)
		return nil, err
	}

	return &config, nil
}
