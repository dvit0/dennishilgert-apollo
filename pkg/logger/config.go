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

// SetAppID sets Application ID.
func (c *Config) SetAppID(id string) {
	c.AppId = id
}

// AttachCmdFlags attaches log config to the command flags.
func (c *Config) AttachCmdFlags(
	stringVar func(p *string, name string, value string, usage string),
	boolVar func(p *bool, name string, value bool, usage string),
) {
	if stringVar != nil {
		stringVar(
			&c.LogLevel,
			"log-level",
			defaultLogLevel,
			"Options are debug, info, warn, error, or fatal (default info)")
	}
	if boolVar != nil {
		boolVar(
			&c.LogJsonOutput,
			"log-as-json",
			defaultJsonOutput,
			"print log as JSON (default false)")
	}
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
