package config

import (
	"github.com/dennishilgert/apollo/pkg/configuration"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/spf13/viper"
)

var log = logger.NewLogger("apollo.fleet.config")

type Config struct {
	ApiPort                   int
	AgentApiPort              int
	DataPath                  string
	FirecrackerBinaryPath     string
	ImageRegistryAddress      string
	MessagingBootstrapServers string
	MessagingWorkerCount      int
	StorageEndpoint           string
	StorageAccessKeyId        string
	StorageSecretAccessKey    string
	WatchdogCheckInterval     int
	WatchdogWorkerCount       int
	ServiceRegistryAddress    string
	HeartbeatInterval         int
}

// Load loads the configuration from the environment.
func Load() (*Config, error) {
	var config Config

	// automatically load environment variables that match
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APOLLO")

	// loading the values from the environment or use default values
	configuration.LoadOrDefault("ApiPort", "APOLLO_API_PORT", 50051)

	configuration.LoadOrDefault("AgentApiPort", "APOLLO_AGENT_API_PORT", 50051)

	configuration.LoadOrDefault("DataPath", "APOLLO_DATA_PATH", "/data")

	configuration.LoadOrDefault("FirecrackerBinaryPath", "APOLLO_FC_BINARY_PATH", nil)

	configuration.LoadOrDefault("ImageRegistryAddress", "APOLLO_IMAGE_REGISTRY_ADDRESS", nil)

	configuration.LoadOrDefault("MessagingBootstrapServers", "APOLLO_MESSAGING_BOOTSTRAP_SERVERS", nil)
	configuration.LoadOrDefault("MessagingWorkerCount", "APOLLO_MESSAGING_WORKER_COUNT", 10)

	configuration.LoadOrDefault("StorageEndpoint", "APOLLO_STORAGE_ENDPOINT", nil)
	configuration.LoadOrDefault("StorageAccessKeyId", "APOLLO_STORAGE_ACCESS_KEY_ID", nil)
	configuration.LoadOrDefault("StorageSecretAccessKey", "APOLLO_STORAGE_SECRET_ACCESS_KEY", nil)

	configuration.LoadOrDefault("WatchdogCheckInterval", "APOLLO_WATCHDOG_CHECK_INTERVAL", 3)
	configuration.LoadOrDefault("WatchdogWorkerCount", "APOLLO_WATCHDOG_WORKER_COUNT", 10)

	configuration.LoadOrDefault("ServiceRegistryAddress", "APOLLO_SERVICE_REGISTRY_ADDRESS", nil)
	configuration.LoadOrDefault("HeartbeatInterval", "APOLLO_HEARTBEAT_INTERVAL", 3)

	// unmarshalling the Config struct
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("unable to unmarshal config: %v", err)
		return nil, err
	}

	return &config, nil
}
