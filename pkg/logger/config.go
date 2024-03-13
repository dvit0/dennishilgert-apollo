package logger

import "fmt"

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

func (c *Config) SetLogLevel(level string) error {
	if toLogLevel(level) == UndefinedLevel {
		return fmt.Errorf("undefined Log Output Level: %s", level)
	}
	c.LogLevel = level
	return nil
}

// SetAppId sets application id.
func (c *Config) SetAppId(id string) {
	c.AppId = id
}

// DefaultConfig returns default configuration values.
func DefaultConfig() Config {
	return Config{
		LogJsonOutput: defaultJsonOutput,
		AppId:         undefinedAppId,
		LogLevel:      defaultLogLevel,
	}
}

// ApplyConfigToLoggers applys options to all registered loggers.
func ApplyConfigToLoggers(config *Config) error {
	internalLoggers := getLoggers()

	// apply formatting options first
	for _, v := range internalLoggers {
		v.EnableJsonOutput(config.LogJsonOutput)

		if config.AppId != undefinedAppId {
			v.SetAppId(config.AppId)
		}
	}

	logLevel := toLogLevel(config.LogLevel)
	if logLevel == UndefinedLevel {
		return fmt.Errorf("invalid value for --log-level: %s", config.LogLevel)
	}

	for _, v := range internalLoggers {
		v.SetLogLevel(logLevel)
	}
	return nil
}
