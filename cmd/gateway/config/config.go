package config

import (
	"github.com/dennishilgert/apollo/internal/pkg/configuration"
	"github.com/dennishilgert/apollo/internal/pkg/logger"
	"github.com/spf13/viper"
)

var log = logger.NewLogger("apollo.config")

type Config struct {
	ApiPort                int
	ServiceRegistryAddress string
}

// Load loads the configuration from the environment.
func Load() (*Config, error) {
	var config Config

	// automatically load environment variables that match
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APOLLO")

	// loading the values from the environment or use default values
	configuration.LoadOrDefault("ApiPort", "APOLLO_API_PORT", 80)

	configuration.LoadOrDefault("ServiceRegistryAddress", "APOLLO_SERVICE_REGISTRY_ADDRESS", nil)

	// unmarshalling the Config struct
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("unable to unmarshal config: %v", err)
		return nil, err
	}

	return &config, nil
}
