package config

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/joho/godotenv"

	"gokick/app/core/database"
	"gokick/app/core/telemetry"
	"gokick/app/internal/shared/config/app"
	"gokick/app/internal/shared/config/glitchtip"
	"gokick/app/internal/shared/config/locale"
	"gokick/app/internal/shared/config/logs"
	"gokick/app/internal/shared/config/otel"
	"gokick/app/internal/shared/config/postgres"
	"gokick/app/internal/shared/config/server"
	"gokick/app/internal/shared/config/smtp"
	"gokick/app/internal/shared/config/tracking"
)

type Config struct {
	App app.Config

	Server server.Config

	Log logs.Config

	GlitchTip glitchtip.Config

	Telemetry telemetry.Config

	Database database.Address

	SMTP smtp.Config

	Tracking tracking.Config

	Locale locale.Config
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err == nil || errors.Is(err, fs.ErrNotExist) {
		return fromEnv()
	}

	return nil, fmt.Errorf("invalid .env: %w", err)
}

func fromEnv() (*Config, error) {
	cfg := &Config{GlitchTip: glitchtip.Read()}

	var err error
	if cfg.App, err = app.Read(); err != nil {
		return nil, err
	}

	if cfg.Server, err = server.Read(); err != nil {
		return nil, err
	}

	if cfg.Log, err = logs.Read(); err != nil {
		return nil, err
	}

	if cfg.Telemetry, err = otel.Read(); err != nil {
		return nil, err
	}

	if cfg.Database, err = postgres.Read(); err != nil {
		return nil, err
	}

	if cfg.SMTP, err = smtp.Read(); err != nil {
		return nil, err
	}

	if cfg.Tracking, err = tracking.Read(); err != nil {
		return nil, err
	}

	if cfg.Locale, err = locale.Read(); err != nil {
		return nil, err
	}

	return cfg, nil
}
