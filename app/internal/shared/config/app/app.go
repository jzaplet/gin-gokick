package app

import (
	"cmp"
	"fmt"
	"net/url"
	"os"
	"runtime/debug"
	"strings"
)

const unversioned = "dev"

type Config struct {
	URL string

	Environment string

	Release string
}

func Read() (Config, error) {
	address, err := parseURL(os.Getenv("APP_URL"))
	if err != nil {
		return Config{}, err
	}

	return Config{
		URL: address,

		Environment: cmp.Or(os.Getenv("APP_ENVIRONMENT"), "production"),

		Release: cmp.Or(os.Getenv("APP_RELEASE"), buildRevision()),
	}, nil
}

func parseURL(value string) (string, error) {
	parsed, err := url.Parse(value)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" || parsed.User != nil || strings.Trim(parsed.Path, "/") != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("APP_URL must be the address of the app without a path, like https://gokick.dev, got %q", value)
	}

	return parsed.Scheme + "://" + parsed.Host, nil
}

func buildRevision() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" && setting.Value != "" {
				return setting.Value
			}
		}
	}

	return unversioned
}
