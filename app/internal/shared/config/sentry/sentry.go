package sentry

import (
	"cmp"
	"os"
)

type Config struct {
	DSN string

	FrontendDSN string

	Environment string

	Release string
}

func Read() Config {
	return Config{
		DSN: os.Getenv("SENTRY_DSN"),

		FrontendDSN: os.Getenv("SENTRY_FRONTEND_DSN"),

		Environment: cmp.Or(os.Getenv("SENTRY_ENVIRONMENT"), "production"),

		Release: os.Getenv("SENTRY_RELEASE"),
	}
}
