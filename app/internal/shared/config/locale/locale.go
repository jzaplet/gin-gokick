package locale

import (
	"errors"
	"os"
	"strings"
)

var ErrMissingDefault = errors.New("DEFAULT_LOCALE is required")

type Config struct {
	Default string

	Locales []string
}

func Read() (Config, error) {
	cfg := Config{Default: os.Getenv("DEFAULT_LOCALE")}
	if cfg.Default == "" {
		return Config{}, ErrMissingDefault
	}

	cfg.Locales = []string{cfg.Default}
	if value := os.Getenv("LOCALES"); value != "" {
		cfg.Locales = strings.Split(value, ",")
	}

	return cfg, nil
}
