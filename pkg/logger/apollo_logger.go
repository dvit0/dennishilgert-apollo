package logger

import (
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type apolloLogger struct {
	// name defines the name of the logger that is published to log as a scope
	name string

	// logger defines the instance of a logrus logger
	logger *logrus.Entry
}

var ApolloVersion = "unknown"

func newApolloLogger(name string) *apolloLogger {
	newLogger := logrus.New()
	newLogger.SetOutput(os.Stdout)

	al := &apolloLogger{
		name: name,
		logger: newLogger.WithFields(logrus.Fields{
			logFieldScope: name,
			logFieldType:  LogTypeLog,
		}),
	}

	al.EnableJSONOutput(defaultJSONOutput)

	return al
}

// EnableJSONOutput enables JSON formatted output logging.
func (l *apolloLogger) EnableJSONOutput(enabled bool) {
	var formatter logrus.Formatter

	fieldMap := logrus.FieldMap{
		// If time field name is conflicted, logrus adds "fields." prefix.
		// So rename to unused field @time to avoid the confliction.
		logrus.FieldKeyTime:  logFieldTimeStamp,
		logrus.FieldKeyLevel: logFieldLevel,
		logrus.FieldKeyMsg:   logFieldMessage,
	}

	hostname, _ := os.Hostname()
	l.logger.Data = logrus.Fields{
		logFieldScope:     l.logger.Data[logFieldScope],
		logFieldType:      LogTypeLog,
		logFieldInstance:  hostname,
		logFieldApolloVer: ApolloVersion,
	}

	if enabled {
		formatter = &logrus.JSONFormatter{ //nolint: exhaustruct
			TimestampFormat: time.RFC3339Nano,
			FieldMap:        fieldMap,
		}
	} else {
		formatter = &logrus.TextFormatter{ //nolint: exhaustruct
			TimestampFormat: time.RFC3339Nano,
			FieldMap:        fieldMap,
		}
	}

	l.logger.Logger.SetFormatter(formatter)
}

func (l *apolloLogger) LogrusEntry() *logrus.Entry {
	return l.logger
}

func (l *apolloLogger) PrintBanner() {
	l.logger.Log(logrus.InfoLevel, "    _             _ _")
	l.logger.Log(logrus.InfoLevel, "   /_\\  _ __  ___| | |___")
	l.logger.Log(logrus.InfoLevel, "  / _ \\| '_ \\/ _ \\ | / _ \\")
	l.logger.Log(logrus.InfoLevel, " /_/ \\_\\ .__/\\___/_|_\\___/")
	l.logger.Log(logrus.InfoLevel, "       |_|")
}

// SetAppId sets app_id field in the log. Default value is an empty string.
func (l *apolloLogger) SetAppId(id string) {
	l.logger = l.logger.WithField(logFieldAppId, id)
}

func toLogrusLevel(lvl LogLevel) logrus.Level {
	// ignore error because it will never happen
	l, _ := logrus.ParseLevel(string(lvl))
	return l
}

// SetOutputLevel sets the log output level.
func (l *apolloLogger) SetOutputLevel(outputLevel LogLevel) {
	l.logger.Logger.SetLevel(toLogrusLevel(outputLevel))
}

// IsOutputLevelEnabled returns true if the logger will output this LogLevel.
func (l *apolloLogger) IsOutputLevelEnabled(level LogLevel) bool {
	return l.logger.Logger.IsLevelEnabled(toLogrusLevel(level))
}

// SetOutput sets the destination for the logs.
func (l *apolloLogger) SetOutput(dst io.Writer) {
	l.logger.Logger.SetOutput(dst)
}

// WithLogType specify the log_type field in log. Default value is LogTypeLog.
func (l *apolloLogger) WithLogType(logType string) Logger {
	return &apolloLogger{
		name:   l.name,
		logger: l.logger.WithField(logFieldType, logType),
	}
}

// WithFields returns a logger with the added structured fields.
func (l *apolloLogger) WithFields(fields map[string]any) Logger {
	return &apolloLogger{
		name:   l.name,
		logger: l.logger.WithFields(fields),
	}
}

// Info logs a message at level Info.
func (l *apolloLogger) Info(args ...interface{}) {
	l.logger.Log(logrus.InfoLevel, args...)
}

// Infof logs a formatted message at level Info.
func (l *apolloLogger) Infof(format string, args ...interface{}) {
	l.logger.Logf(logrus.InfoLevel, format, args...)
}

// Debug logs a message at level Debug.
func (l *apolloLogger) Debug(args ...interface{}) {
	l.logger.Log(logrus.DebugLevel, args...)
}

// Debugf logs a formatted message at level Debug.
func (l *apolloLogger) Debugf(format string, args ...interface{}) {
	l.logger.Logf(logrus.DebugLevel, format, args...)
}

// Warn logs a message at level Warn.
func (l *apolloLogger) Warn(args ...interface{}) {
	l.logger.Log(logrus.WarnLevel, args...)
}

// Warnf logs a formatted message at level Warn.
func (l *apolloLogger) Warnf(format string, args ...interface{}) {
	l.logger.Logf(logrus.WarnLevel, format, args...)
}

// Error logs a message at level Error.
func (l *apolloLogger) Error(args ...interface{}) {
	l.logger.Log(logrus.ErrorLevel, args...)
}

// Errorf logs a formatted message at level Error.
func (l *apolloLogger) Errorf(format string, args ...interface{}) {
	l.logger.Logf(logrus.ErrorLevel, format, args...)
}

// Fatal logs a message at level Fatal then the process will exit with status set to 1.
func (l *apolloLogger) Fatal(args ...interface{}) {
	l.logger.Fatal(args...)
}

// Fatalf logs a formatted message at level Fatal then the process will exit with status set to 1.
func (l *apolloLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}
