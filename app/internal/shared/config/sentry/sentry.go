package sentry

import "os"

type Config struct {
	DSN string

	FrontendDSN string
}

func Read() Config {
	return Config{DSN: os.Getenv("SENTRY_DSN"), FrontendDSN: os.Getenv("SENTRY_FRONTEND_DSN")}
}
