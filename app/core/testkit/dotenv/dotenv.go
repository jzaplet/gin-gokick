package dotenv

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
)

var errUnset = errors.New("is set neither in the environment nor in .env; copy it from .env.example")

var errNoDotenv = errors.New("no .env next to go.mod; copy .env.example")

var errBadDotenv = errors.New(".env cannot be parsed")

var errNoModule = errors.New("no go.mod above the working directory")

func Load(tb testing.TB) {
	tb.Helper()

	dir, err := os.Getwd()
	if err != nil {
		tb.Fatal(err)
	}

	root, err := moduleRoot(dir)
	if err != nil {
		tb.Fatal(err)
	}

	env, err := godotenv.Read(filepath.Join(root, ".env"))
	if errors.Is(err, fs.ErrNotExist) {
		return
	}

	if err != nil {
		tb.Fatal(errBadDotenv)
	}

	for key, value := range env {
		if _, set := os.LookupEnv(key); set == false {
			tb.Setenv(key, value)
		}
	}
}

func Lookup(dir, first string, rest ...string) (map[string]string, error) {
	get := os.Getenv
	if get(first) == "" {
		env, err := read(dir)
		if err != nil {
			return nil, err
		}

		if env[first] == "" {
			return nil, fmt.Errorf("%s %w", first, errUnset)
		}

		get = func(key string) string { return env[key] }
	}

	values := map[string]string{}
	for _, key := range append([]string{first}, rest...) {
		values[key] = get(key)
	}

	return values, nil
}

func read(dir string) (map[string]string, error) {
	root, err := moduleRoot(dir)
	if err != nil {
		return nil, err
	}

	env, err := godotenv.Read(filepath.Join(root, ".env"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, errNoDotenv
	}

	if err != nil {
		return nil, errBadDotenv
	}

	return env, nil
}

func moduleRoot(dir string) (string, error) {
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errNoModule
		}

		dir = parent
	}
}
