package prometheus

import (
	"errors"
	"os"
)

var ErrHalfAccount = errors.New("METRICS_USER and METRICS_PASSWORD must be set together")

type Config struct {
	User string

	Password string
}

func Read() (Config, error) {
	account := Config{User: os.Getenv("METRICS_USER"), Password: os.Getenv("METRICS_PASSWORD")}
	if (account.User == "") != (account.Password == "") {
		return Config{}, ErrHalfAccount
	}

	return account, nil
}
