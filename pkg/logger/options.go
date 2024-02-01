package logger

import "fmt"

const (
	defaultJSONOutput  = false
	defaultOutputLevel = "info"
	undefinedAppId     = ""
)

type Options struct {
	// appID is the unique id of the Apollo application
	appId string

	// JSONFormatEnabled defines the flag to enable JSON formatted log
	JSONFormatEnabled bool

	// OutputLevel defines the level of logging
	OutputLevel string
}

func (o *Options) SetOutputLevel(level string) error {
	if toLogLevel(level) == UndefinedLevel {
		return fmt.Errorf("undefined Log Output Level: %s", level)
	}
	o.OutputLevel = level
	return nil
}

// SetAppID sets Application ID.
func (o *Options) SetAppID(id string) {
	o.appId = id
}

// AttachCmdFlags attaches log options to the command flags.
func (o *Options) AttachCmdFlags(
	stringVar func(p *string, name string, value string, usage string),
	boolVar func(p *bool, name string, value bool, usage string),
) {
	if stringVar != nil {
		stringVar(
			&o.OutputLevel,
			"log-level",
			defaultOutputLevel,
			"Options are debug, info, warn, error, or fatal (default info)")
	}
	if boolVar != nil {
		boolVar(
			&o.JSONFormatEnabled,
			"log-as-json",
			defaultJSONOutput,
			"print log as JSON (default false)")
	}
}

// DefaultOptions returns default values of Options.
func DefaultOptions() Options {
	return Options{
		JSONFormatEnabled: defaultJSONOutput,
		appId:             undefinedAppId,
		OutputLevel:       defaultOutputLevel,
	}
}

// ApplyOptionsToLoggers applys options to all registered loggers.
func ApplyOptionsToLoggers(options *Options) error {
	internalLoggers := getLoggers()

	// apply formatting options first
	for _, v := range internalLoggers {
		v.EnableJSONOutput(options.JSONFormatEnabled)

		if options.appId != undefinedAppId {
			v.SetAppId(options.appId)
		}
	}

	daprLogLevel := toLogLevel(options.OutputLevel)
	if daprLogLevel == UndefinedLevel {
		return fmt.Errorf("invalid value for --log-level: %s", options.OutputLevel)
	}

	for _, v := range internalLoggers {
		v.SetOutputLevel(daprLogLevel)
	}
	return nil
}
