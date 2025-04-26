package config

import (
	"flag"
)

type Flag struct {
	LogHandler     string
	ExcluedEnvFile bool
	EnableTestDb   bool
}

func InitFlags(args []string) *Flag {
	fs := flag.NewFlagSet("Wonk", flag.ContinueOnError)

	logHandler := fs.String("logfmt", "json", "slog logging format")
	excludeEnvFile := fs.Bool("exclude-env", false, "do we need to exlcude reading env file")
	enableTestDb := fs.Bool("test-db", false, "do we need a in memory db for testing")

	fs.Parse(args)

	return &Flag{
		LogHandler:     *logHandler,
		ExcluedEnvFile: *excludeEnvFile,
		EnableTestDb:   *enableTestDb,
	}
}
