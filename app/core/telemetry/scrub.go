package telemetry

import (
	"context"
	"net/url"
	"slices"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var personal = []attribute.Key{
	"client.address",
	"network.peer.address",
	"network.peer.port",
	"url.path",
	"http.client_ip",
	"http.target",
	"net.peer.ip",
	"net.sock.peer.addr",
}

var addresses = []attribute.Key{"url.full", "http.url"}

var secretInURL = []string{"api.telegram.org"}

type scrubbingExporter struct {
	sdktrace.SpanExporter
}

type scrubbedSpan struct {
	sdktrace.ReadOnlySpan

	attributes []attribute.KeyValue
}

func (e scrubbingExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	scrubbed := make([]sdktrace.ReadOnlySpan, 0, len(spans))
	for _, span := range spans {
		scrubbed = append(scrubbed, scrubbedSpan{ReadOnlySpan: span, attributes: scrub(span.Attributes())})
	}

	return e.SpanExporter.ExportSpans(ctx, scrubbed)
}

func (s scrubbedSpan) Attributes() []attribute.KeyValue {
	return s.attributes
}

func scrub(attributes []attribute.KeyValue) []attribute.KeyValue {
	kept := make([]attribute.KeyValue, 0, len(attributes))
	for _, attr := range attributes {
		switch {
		case slices.Contains(personal, attr.Key):
		case slices.Contains(addresses, attr.Key):
			kept = append(kept, attribute.String(string(attr.Key), withoutSecret(attr.Value.AsString())))
		default:
			kept = append(kept, attr)
		}
	}

	return kept
}

func withoutSecret(address string) string {
	parsed, err := url.Parse(address)
	if err != nil {
		return ""
	}

	if slices.Contains(secretInURL, parsed.Hostname()) {
		return parsed.Scheme + "://" + parsed.Host
	}

	parsed.User = nil

	return parsed.String()
}
