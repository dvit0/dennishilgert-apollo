package config

import (
	"github.com/dennishilgert/apollo/internal/pkg/configuration"
	"github.com/dennishilgert/apollo/internal/pkg/logger"
	"github.com/spf13/viper"
)

var log = logger.NewLogger("apollo.config")

type Config struct {
	ApiPort                   int
	MessagingBootstrapServers string
	MessagingWorkerCount      int
	ServiceRegistryAddress    string
	HeartbeatInterval         int
	DatabaseHost              string
	DatabasePort              int
	DatabaseUsername          string
	DatabasePassword          string
	DatabaseDb                string
	DatabaseAuthDb            string
}

// Load loads the configuration from the environment.
func Load() (*Config, error) {
	var config Config

	// automatically load environment variables that match
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APOLLO")

	// loading the values from the environment or use default values
	configuration.LoadOrDefault("ApiPort", "APOLLO_API_PORT", 50051)

	configuration.LoadOrDefault("MessagingBootstrapServers", "APOLLO_MESSAGING_BOOTSTRAP_SERVERS", nil)
	configuration.LoadOrDefault("MessagingWorkerCount", "APOLLO_MESSAGING_WORKER_COUNT", 10)

	configuration.LoadOrDefault("ServiceRegistryAddress", "APOLLO_SERVICE_REGISTRY_ADDRESS", nil)
	configuration.LoadOrDefault("HeartbeatInterval", "APOLLO_HEARTBEAT_INTERVAL", 3)

	configuration.LoadOrDefault("DatabaseHost", "APOLLO_DATABASE_HOST", nil)
	configuration.LoadOrDefault("DatabasePort", "APOLLO_DATABASE_PORT", 5432)
	configuration.LoadOrDefault("DatabaseUsername", "APOLLO_DATABASE_USERNAME", nil)
	configuration.LoadOrDefault("DatabasePassword", "APOLLO_DATABASE_PASSWORD", nil)
	configuration.LoadOrDefault("DatabaseDb", "APOLLO_DATABASE_DB", "apollo")
	configuration.LoadOrDefault("DatabaseAuthDb", "APOLLO_DATABASE_AUTH_DB", "admin")

	// unmarshalling the Config struct
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("unable to unmarshal config: %v", err)
		return nil, err
	}

	return &config, nil
}
