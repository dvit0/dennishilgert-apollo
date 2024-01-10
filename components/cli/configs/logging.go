package configs

import (
	"github.com/hashicorp/go-hclog"
	"github.com/spf13/pflag"
)

type LogConfig struct {
	flagBase

	LogLevel      string
	LogColor      bool
	LogForceColor bool
	LogAsJson     bool
}

func NewLoggingConfig() *LogConfig {
	return &LogConfig{}
}

func (c *LogConfig) FlagSet() *pflag.FlagSet {
	if c.initFlagSet() {
		c.flagSet.StringVar(&c.LogLevel, "log-level", "debug", "Log Level")
		c.flagSet.BoolVar(&c.LogAsJson, "log-as-json", false, "Log as JSON")
		c.flagSet.BoolVar(&c.LogColor, "log-color", true, "Log in color")
		c.flagSet.BoolVar(&c.LogForceColor, "log-force-color", false, "Force colored log output")
	}
	return c.flagSet
}

func (c *LogConfig) NewLogger(name string) hclog.Logger {
	loggerColorOption := hclog.ColorOff
	if c.LogColor {
		loggerColorOption = hclog.AutoColor
	}
	if c.LogForceColor {
		loggerColorOption = hclog.ForceColor
	}

	return hclog.New(&hclog.LoggerOptions{
		Name:       name,
		Level:      hclog.LevelFromString(c.LogLevel),
		Color:      loggerColorOption,
		JSONFormat: c.LogAsJson,
	})
}
