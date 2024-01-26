package main

import "github.com/hashicorp/go-hclog"

const VERSION = "v1.0.0"
const timeFormat = "02-01-2006 15:04:05.000"

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:            "main",
		Level:           hclog.Debug,
		Color:           hclog.AutoColor,
		ColorHeaderOnly: true,
		TimeFormat:      timeFormat,
	})

	logger.Info("apollo worker manager - " + VERSION)
}
