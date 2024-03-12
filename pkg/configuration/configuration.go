package configuration

import (
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/spf13/viper"
)

var log = logger.NewLogger("apollo.config")

func LoadOrDefault(configVar string, envVar string, defaultVal any) {
	if defaultVal != nil {
		viper.SetDefault(configVar, defaultVal)
	}
	viper.BindEnv(configVar, envVar)
	if defaultVal == nil {
		if !viper.IsSet(configVar) {
			log.Fatalf("required environment variable %s is not set", envVar)
		}
	}
}
