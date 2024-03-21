package messaging

import "time"

type Config struct {
	AdminOperationTimeout time.Duration
}

func DefaultConfig() Config {
	return Config{
		AdminOperationTimeout: time.Second * 60,
	}
}
