package logger

import (
	"fmt"
	"os"

	"github.com/dennishilgert/apollo/internal/pkg/configuration"
	"github.com/spf13/viper"
)

const (
	defaultJsonOutput = false
	defaultLogLevel   = "info"
	undefinedAppId    = ""
)

type Config struct {
	// AppId is the unique id of the Apollo application
	AppId string

	// LogJsonOutput defines the flag to enable JSON formatted log
	LogJsonOutput bool

	// LogLevel defines the level of logging
	LogLevel string
}

// DefaultConfig returns default configuration values.
func DefaultConfig() Config {
	return Config{
		LogJsonOutput: defaultJsonOutput,
		AppId:         undefinedAppId,
		LogLevel:      defaultLogLevel,
	}
}

// LoadConfig loads the configuration from the environment.
func LoadConfig() Config {
	var config Config

	// automatically load environment variables that match
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APOLLO")

	// loading the values from the environment or use default values
	configuration.LoadOrDefault("AppId", "APOLLO_LOG_APP_ID", DefaultConfig().AppId)
	configuration.LoadOrDefault("LogJsonOutput", "APOLLO_LOG_FORMAT_JSON", DefaultConfig().LogJsonOutput)
	configuration.LoadOrDefault("LogLevel", "APOLLO_LOG_LEVEL", DefaultConfig().LogLevel)

	// unmarshalling the Config struct
	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("unable to unmarshal logger config: %v\n", err)
		os.Exit(1)
	}

	return config
}
