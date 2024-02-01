package logger

import (
	"io"

	"github.com/sirupsen/logrus"
)

type nopLogger struct{}

// EnableJSONOutput enables JSON formatted output logging.
func (n *nopLogger) EnableJSONOutput(_ bool) {}

func (n *nopLogger) LogrusEntry() *logrus.Entry {
	return nil
}

// SetAppID sets app_id field in the log. nopLogger value is an empty string.
func (n *nopLogger) SetAppId(_ string) {}

// SetOutputLevel sets the log output level.
func (n *nopLogger) SetOutputLevel(_ LogLevel) {}

// SetOutput sets the destination for the logs.
func (n *nopLogger) SetOutput(_ io.Writer) {}

// IsOutputLevelEnabled returns true if the logger will output this LogLevel.
func (n *nopLogger) IsOutputLevelEnabled(_ LogLevel) bool { return true }

// WithLogType specify the log_type field in log. nopLogger value is LogTypeLog.
func (n *nopLogger) WithLogType(_ string) Logger {
	return n
}

// WithFields returns a logger with the added structured fields.
func (n *nopLogger) WithFields(_ map[string]any) Logger {
	return n
}

// Info logs a message at level Info.
func (n *nopLogger) Info(_ ...interface{}) {}

// Infof logs a formatted message at level Info.
func (n *nopLogger) Infof(_ string, _ ...interface{}) {}

// Debug logs a message at level Debug.
func (n *nopLogger) Debug(_ ...interface{}) {}

// Debugf logs a formatted message at level Debug.
func (n *nopLogger) Debugf(_ string, _ ...interface{}) {}

// Warn logs a message at level Warn.
func (n *nopLogger) Warn(_ ...interface{}) {}

// Warnf logs a formatted message at level Warn.
func (n *nopLogger) Warnf(_ string, _ ...interface{}) {}

// Error logs a message at level Error.
func (n *nopLogger) Error(_ ...interface{}) {}

// Errorf logs a formatted message at level Error.
func (n *nopLogger) Errorf(_ string, _ ...interface{}) {}

// Fatal logs a message at level Fatal then the process will exit with status set to 1.
func (n *nopLogger) Fatal(_ ...interface{}) {}

// Fatalf logs a formatted message at level Fatal then the process will exit with status set to 1.
func (n *nopLogger) Fatalf(_ string, _ ...interface{}) {}
