package server

import (
	"cmp"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const defaultPort = 8020

type Config struct {
	Port uint16

	GinMode string

	TrustedProxies []string

	URL string
}

func Read() (Config, error) {
	port, err := parsePort(cmp.Or(os.Getenv("PORT"), strconv.Itoa(defaultPort)))
	if err != nil {
		return Config{}, err
	}

	mode := cmp.Or(os.Getenv("GIN_MODE"), gin.ReleaseMode)
	switch mode {
	case gin.DebugMode, gin.ReleaseMode, gin.TestMode:
	default:
		return Config{}, fmt.Errorf("GIN_MODE must be debug, release or test, got %q", mode)
	}

	proxies, err := parseProxies(os.Getenv("TRUSTED_PROXIES"))
	if err != nil {
		return Config{}, err
	}

	address, err := parseURL(os.Getenv("APP_URL"))
	if err != nil {
		return Config{}, err
	}

	return Config{Port: port, GinMode: mode, TrustedProxies: proxies, URL: address}, nil
}

func (c Config) Addr() string {
	return ":" + strconv.FormatUint(uint64(c.Port), 10)
}

func parsePort(value string) (uint16, error) {
	port, err := strconv.ParseUint(value, 10, 16)
	if err != nil || port == 0 {
		return 0, fmt.Errorf("PORT must be a number from 1 to 65535, got %q", value)
	}

	return uint16(port), nil
}

func parseURL(value string) (string, error) {
	parsed, err := url.Parse(value)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" || parsed.User != nil || strings.Trim(parsed.Path, "/") != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("APP_URL must be the address of the app without a path, like https://gokick.dev, got %q", value)
	}

	return parsed.Scheme + "://" + parsed.Host, nil
}

func parseProxies(value string) ([]string, error) {
	var proxies []string

	for entry := range strings.SplitSeq(value, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		if isIPOrCIDR(entry) == false {
			return nil, fmt.Errorf("TRUSTED_PROXIES: %q is neither an IP address nor a CIDR", entry)
		}

		proxies = append(proxies, entry)
	}

	return proxies, nil
}

func isIPOrCIDR(value string) bool {
	if strings.Contains(value, "/") {
		_, _, err := net.ParseCIDR(value)

		return err == nil
	}

	return net.ParseIP(value) != nil
}
