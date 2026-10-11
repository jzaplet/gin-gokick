package otel

import (
	"reflect"
	"strings"
	"testing"

	"gokick/app/core/telemetry"
)

func emptyEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{"OTEL_EXPORTER_OTLP_ENDPOINT", "OTEL_EXPORTER_OTLP_HEADERS", "OTEL_SERVICE_NAME", "OTEL_RESOURCE_ATTRIBUTES", "OTEL_TRACES_SAMPLER_ARG"} {
		t.Setenv(key, "")
	}
}

func TestDefaults(t *testing.T) {
	emptyEnv(t)

	cfg, err := Read()
	if want := (telemetry.Config{
		Headers: map[string]string{},

		ServiceName: "app",

		Attributes: map[string]string{},

		SampleRatio: 1,
	}); err != nil || reflect.DeepEqual(cfg, want) == false {
		t.Errorf("got %+v, %v", cfg, err)
	}
}

func TestReadsTheEnvironment(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel:4318")
	t.Setenv("OTEL_EXPORTER_OTLP_HEADERS", "Authorization=Basic%20a2V5, X-Scope = team")
	t.Setenv("OTEL_SERVICE_NAME", "worker")
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "service.namespace=shop,service.instance.id=app-1")
	t.Setenv("OTEL_TRACES_SAMPLER_ARG", "0.25")

	cfg, err := Read()
	if err != nil {
		t.Fatal(err)
	}

	want := telemetry.Config{
		Endpoint: "http://otel:4318",

		Headers: map[string]string{"Authorization": "Basic a2V5", "X-Scope": "team"},

		ServiceName: "worker",

		Attributes: map[string]string{"service.namespace": "shop", "service.instance.id": "app-1"},

		SampleRatio: 0.25,
	}
	if reflect.DeepEqual(cfg, want) == false {
		t.Errorf("got %+v, want %+v", cfg, want)
	}
}

func TestRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name, key, value string
	}{
		{"endpoint without a scheme", "OTEL_EXPORTER_OTLP_ENDPOINT", "otel:4318"},
		{"endpoint of another scheme", "OTEL_EXPORTER_OTLP_ENDPOINT", "grpc://otel:4317"},
		{"endpoint with a query", "OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel:4318/?a=1"},
		{"header without a value", "OTEL_EXPORTER_OTLP_HEADERS", "Authorization"},
		{"header with a broken escape", "OTEL_EXPORTER_OTLP_HEADERS", "Authorization=secret%zz"},
		{"attribute without a key", "OTEL_RESOURCE_ATTRIBUTES", "=shop"},
		{"ratio not a number", "OTEL_TRACES_SAMPLER_ARG", "all"},
		{"ratio above one", "OTEL_TRACES_SAMPLER_ARG", "1.5"},
		{"ratio below zero", "OTEL_TRACES_SAMPLER_ARG", "-0.1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emptyEnv(t)
			t.Setenv(tt.key, tt.value)

			if _, err := Read(); err == nil || strings.Contains(err.Error(), tt.key) == false {
				t.Errorf("got %v", err)
			}
		})
	}
}

func TestErrorsLeaveTheHeaderValuesOut(t *testing.T) {
	emptyEnv(t)
	t.Setenv("OTEL_EXPORTER_OTLP_HEADERS", "X-Scope=team,Authorization=secret-key%zz")

	if _, err := Read(); err == nil || strings.Contains(err.Error(), "secret") {
		t.Errorf("got %v", err)
	}
}
