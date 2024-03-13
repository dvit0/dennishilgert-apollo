package logger

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type loggingFlags struct {
	Config Config
}

type parsedFlags struct {
	logFlags *loggingFlags
	flagSet  *pflag.FlagSet
}

func ParseFlags() *parsedFlags {
	var f loggingFlags

	fs := pflag.NewFlagSet("logging", pflag.ExitOnError)
	fs.SortFlags = true

	fs.StringVar(&f.Config.AppId, "log-app-id", "", "App id that should be displayed in the logs")
	fs.StringVar(&f.Config.LogLevel, "log-level", "info", "Log level for which the logs should be displayed")
	fs.BoolVar(&f.Config.LogJsonOutput, "log-json-out", false, "Wether the log output should be printed in json format or not")

	return &parsedFlags{
		logFlags: &f,
		flagSet:  fs,
	}
}

func ReadAndApply(command *cobra.Command, logger Logger) {
	appId, err := command.Flags().GetString("log-app-id")
	if err != nil {
		logger.Fatalf("failed to apply logger configuration: %v", err)
	}
	logger.SetAppId(appId)
	logLevel, err := command.Flags().GetString("log-level")
	if err != nil {
		logger.Fatalf("failed to apply logger configuration: %v", err)
	}
	logger.SetLogLevel(LogLevel(logLevel))
	logJsonOut, err := command.Flags().GetBool("log-json-out")
	if err != nil {
		logger.Fatalf("failed to apply logger configuration: %v", err)
	}
	logger.EnableJsonOutput(logJsonOut)
}

func (p *parsedFlags) LoggingFlags() *loggingFlags {
	return p.logFlags
}

func (p *parsedFlags) FlagSet() *pflag.FlagSet {
	return p.flagSet
}
