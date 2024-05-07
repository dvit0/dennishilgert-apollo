package config

import (
	"github.com/dennishilgert/apollo/internal/pkg/configuration"
	"github.com/dennishilgert/apollo/internal/pkg/logger"
	"github.com/spf13/viper"
)

var log = logger.NewLogger("apollo.registry.config")

type Config struct {
	ApiPort                   int
	WorkingDir                string
	PackageWorkerCount        int
	ImageRegistryAddress      string
	MessagingBootstrapServers string
	MessagingWorkerCount      int
	ServiceRegistryAddress    string
	HeartbeatInterval         int
	StorageEndpoint           string
	StorageAccessKeyId        string
	StorageSecretAccessKey    string
}

// Load loads the configuration from the environment.
func Load() (*Config, error) {
	var config Config

	// automatically load environment variables that match
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APOLLO")

	// loading the values from the environment or use default values
	configuration.LoadOrDefault("ApiPort", "APOLLO_API_PORT", 50051)
	configuration.LoadOrDefault("WorkingDir", "APOLLO_WORKING_DIR", nil)
	configuration.LoadOrDefault("PackageWorkerCount", "APOLLO_PACKAGE_WORKER_COUNT", 10)

	configuration.LoadOrDefault("ImageRegistryAddress", "APOLLO_IMAGE_REGISTRY_ADDRESS", nil)

	configuration.LoadOrDefault("MessagingBootstrapServers", "APOLLO_MESSAGING_BOOTSTRAP_SERVERS", nil)
	configuration.LoadOrDefault("MessagingWorkerCount", "APOLLO_MESSAGING_WORKER_COUNT", 10)

	configuration.LoadOrDefault("ServiceRegistryAddress", "APOLLO_SERVICE_REGISTRY_ADDRESS", nil)
	configuration.LoadOrDefault("HeartbeatInterval", "APOLLO_HEARTBEAT_INTERVAL", 3)

	configuration.LoadOrDefault("StorageEndpoint", "APOLLO_STORAGE_ENDPOINT", nil)
	configuration.LoadOrDefault("StorageAccessKeyId", "APOLLO_STORAGE_ACCESS_KEY_ID", nil)
	configuration.LoadOrDefault("StorageSecretAccessKey", "APOLLO_STORAGE_SECRET_ACCESS_KEY", nil)

	// unmarshalling the Config struct
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("unable to unmarshal config: %v", err)
		return nil, err
	}

	return &config, nil
}