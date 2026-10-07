package postgres

import (
	"errors"
	"fmt"
	"os"

	"gokick/app/core/database"
)

var ErrMissing = errors.New("is required")

func Read() (database.Address, error) {
	for _, key := range []string{"DB_HOST", "DB_NAME", "DB_USERNAME", "DB_PASSWORD"} {
		if os.Getenv(key) == "" {
			return database.Address{}, fmt.Errorf("%s %w", key, ErrMissing)
		}
	}

	port, err := database.ParsePort(os.Getenv("DB_PORT"))
	if err != nil {
		return database.Address{}, fmt.Errorf("DB_PORT %w, got %q", err, os.Getenv("DB_PORT"))
	}

	params, err := database.ParseParams(os.Getenv("DB_PARAMS"))
	if err != nil {
		return database.Address{}, fmt.Errorf("DB_PARAMS %w", err)
	}

	return database.Address{
		Host: os.Getenv("DB_HOST"),

		Port: port,

		Name: os.Getenv("DB_NAME"),

		Username: os.Getenv("DB_USERNAME"),

		Password: os.Getenv("DB_PASSWORD"),

		Params: params,
	}, nil
}
