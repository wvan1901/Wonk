package config

import (
	"flag"
)

type Flag struct {
	LogHandler string
}

func InitFlags() *Flag {
	logHandler := flag.String("logfmt", "json", "slog logging format")

	flag.Parse()

	return &Flag{
		LogHandler: *logHandler,
	}
}
