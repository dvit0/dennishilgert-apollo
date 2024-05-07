package messaging

import "time"

type Config struct {
	AdminOperationTimeout time.Duration
}

// DefaultConfig returns default configuration values.
func DefaultConfig() Config {
	return Config{
		AdminOperationTimeout: time.Second * 60,
	}
}
