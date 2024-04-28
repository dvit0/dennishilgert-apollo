package config

import (
	"github.com/dennishilgert/apollo/pkg/configuration"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/spf13/viper"
)

var log = logger.NewLogger("apollo.worker.config")

type Config struct {
	ApiPort                   int
	MessagingBootstrapServers string
	MessagingWorkerCount      int
	CacheAddress              string
	CacheUsername             string
	CachePassword             string
	CacheDatabase             int
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

	configuration.LoadOrDefault("CacheAddress", "APOLLO_CACHE_ADDRESS", nil)
	configuration.LoadOrDefault("CacheUsername", "APOLLO_CACHE_USERNAME", "")
	configuration.LoadOrDefault("CachePassword", "APOLLO_CACHE_PASSWORD", "")
	configuration.LoadOrDefault("CacheDatabase", "APOLLO_CACHE_DATABASE", 0)

	// unmarshalling the Config struct
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("unable to unmarshal config: %v", err)
		return nil, err
	}

	return &config, nil
}
