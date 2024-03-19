package configuration

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func LoadOrDefault(configVar string, envVar string, defaultVal any) {
	if defaultVal != nil {
		viper.SetDefault(configVar, defaultVal)
	}
	viper.BindEnv(configVar, envVar)
	if defaultVal == nil {
		if !viper.IsSet(configVar) {
			fmt.Printf("required environment variable is not set: %s\n", envVar)
			os.Exit(1)
		}
	}
}
