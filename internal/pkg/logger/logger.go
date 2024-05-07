package logger

import (
	"context"
	"io"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

const (
	// LogTypeLog defines the default log type.
	LogTypeLog = "log"

	// LogTypeRequest defines the request log type.
	LogTypeRequest = "request"

	// Field names that define Apollo log schema.
	logFieldTimeStamp = "time"
	logFieldLevel     = "level"
	logFieldType      = "type"
	logFieldScope     = "scope"
	logFieldMessage   = "msg"
	logFieldInstance  = "instance"
	logFieldApolloVer = "ver"
	logFieldAppId     = "app_id"
)

type logContextKeyType struct{}

// logContextKey defines how we find Loggers in a context.Context.
var logContextKey = logContextKeyType{}

// LogLevel defines the Logger level type.
type LogLevel string

const (
	// DebugLevel is used to log verbose messages.
	DebugLevel LogLevel = "debug"

	// InfoLevel is the default log level.
	InfoLevel LogLevel = "info"

	// WarnLevel is used to log messages about possible issues.
	WarnLevel LogLevel = "warn"

	// ErrorLevel is used to log errors.
	ErrorLevel LogLevel = "error"

	// FatalLevel os used to log fatal messages. The system shuts down after logging the message.
	FatalLevel LogLevel = "fatal"

	// UndefinedLevel is used for an undefined log level.
	UndefinedLevel LogLevel = "undefined"
)

var (
	globalLoggers     = map[string]Logger{}
	globalLoggersLock = sync.RWMutex{}
	defaultOpLogger   = &nopLogger{}
)

var config Config = LoadConfig()

type Logger interface {
	// EnableJsonOutput enables JSON formatted output log.
	EnableJsonOutput(enabled bool)

	// Logger returns the logger instance.
	LogrusEntry() *logrus.Entry

	// SetAppId sets app_id field in the log. Default value is empty string.
	SetAppId(id string)

	// SetLogLevel sets the log level.
	SetLogLevel(logLevel LogLevel)

	// LogLevel returns the log level of the logger.
	LogLevel() string

	// SetOutput sets the destination for the logs.
	SetOutput(dst io.Writer)

	// IsLogLevelEnabled returns true if the logger will log this LogLevel.
	IsLogLevelEnabled(level LogLevel) bool

	// WithLogType specifies the log_type field in log. Default value is LogTypeLog
	WithLogType(logType string) Logger

	// WithFields returns a logger with the added structured fields.
	WithFields(fields map[string]any) Logger

	// Info logs a message at level Info.
	Info(args ...interface{})
	// Infof logs a message at level Info.
	Infof(format string, args ...interface{})
	// Debug logs a message at level Debug.
	Debug(args ...interface{})
	// Debugf logs a message at level Debug.
	Debugf(format string, args ...interface{})
	// Warn logs a message at level Warn.
	Warn(args ...interface{})
	// Warnf logs a message at level Warn.
	Warnf(format string, args ...interface{})
	// Error logs a message at level Error.
	Error(args ...interface{})
	// Errorf logs a message at level Error.
	Errorf(format string, args ...interface{})
	// Fatal logs a message at level Fatal then the process will exit with status set to 1.
	Fatal(args ...interface{})
	// Fatalf logs a message at level Fatal then the process will exit with status set to 1.
	Fatalf(format string, args ...interface{})
}

// toLogLevel converts to LogLevel.
func toLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "fatal":
		return FatalLevel
	}

	// Unsupported log level by apollo.
	return UndefinedLevel
}

// NewLogger creates new Logger instance.
func NewLogger(name string) Logger {
	globalLoggersLock.Lock()
	defer globalLoggersLock.Unlock()

	logger, ok := globalLoggers[name]
	if !ok {
		logger = newApolloLogger(name)

		// apply logger config
		logger.SetAppId(config.AppId)
		logger.SetLogLevel(toLogLevel(config.LogLevel))
		logger.EnableJsonOutput(config.LogJsonOutput)

		globalLoggers[name] = logger
	}

	return logger
}

// NewContext returns a new Context, derived from ctx, which carries the
// provided Logger.
func NewContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, logContextKey, logger)
}

// FromContextOrDiscard returns a Logger from ctx.  If no Logger is found, this
// returns a Logger that discards all log messages.
func FromContextOrDefault(ctx context.Context) Logger {
	if v, ok := ctx.Value(logContextKey).(Logger); ok {
		return v
	}

	return defaultOpLogger
}
