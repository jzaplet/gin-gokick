package otel

import (
	"cmp"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"gokick/app/core/telemetry"
)

func Read() (telemetry.Config, error) {
	endpoint, err := parseEndpoint(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	if err != nil {
		return telemetry.Config{}, err
	}

	headers, err := parsePairs("OTEL_EXPORTER_OTLP_HEADERS")
	if err != nil {
		return telemetry.Config{}, err
	}

	attributes, err := parsePairs("OTEL_RESOURCE_ATTRIBUTES")
	if err != nil {
		return telemetry.Config{}, err
	}

	ratio, err := parseRatio(cmp.Or(os.Getenv("OTEL_TRACES_SAMPLER_ARG"), "1"))
	if err != nil {
		return telemetry.Config{}, err
	}

	return telemetry.Config{
		Endpoint: endpoint,

		Headers: headers,

		ServiceName: cmp.Or(os.Getenv("OTEL_SERVICE_NAME"), "app"),

		Attributes: attributes,

		SampleRatio: ratio,
	}, nil
}

func parseEndpoint(value string) (string, error) {
	if value == "" {
		return "", nil
	}

	parsed, err := url.Parse(value)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("OTEL_EXPORTER_OTLP_ENDPOINT must be the base address of an OTLP/HTTP receiver, like http://localhost:4318, got %q", value)
	}

	return value, nil
}

func parsePairs(name string) (map[string]string, error) {
	pairs := map[string]string{}

	value := os.Getenv(name)
	if value == "" {
		return pairs, nil
	}

	for position, item := range strings.Split(value, ",") {
		key, encoded, found := strings.Cut(item, "=")
		key = strings.TrimSpace(key)

		decoded, err := url.PathUnescape(strings.TrimSpace(encoded))
		if found == false || key == "" || err != nil {
			return nil, fmt.Errorf("%s must be key=value pairs separated by commas with percent-encoded values; pair %d is not", name, position+1)
		}

		pairs[key] = decoded
	}

	return pairs, nil
}

func parseRatio(value string) (float64, error) {
	ratio, err := strconv.ParseFloat(value, 64)
	if err != nil || ratio < 0 || ratio > 1 {
		return 0, fmt.Errorf("OTEL_TRACES_SAMPLER_ARG must be a number from 0 to 1, got %q", value)
	}

	return ratio, nil
}
