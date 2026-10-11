package collector

import (
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	"google.golang.org/protobuf/proto"
)

type Span struct {
	Name string

	TraceID string

	ParentSpanID string

	Attributes map[string]string

	Resource map[string]string
}

type Request struct {
	Path string

	Header http.Header
}

type Collector struct {
	URL string

	mu sync.Mutex

	requests []Request

	spans []Span
}

func Start(tb testing.TB) *Collector {
	tb.Helper()

	collector := &Collector{}
	server := httptest.NewServer(http.HandlerFunc(collector.receive))
	tb.Cleanup(server.Close)
	collector.URL = server.URL

	return collector
}

func (c *Collector) Requests() []Request {
	c.mu.Lock()
	defer c.mu.Unlock()

	return append([]Request(nil), c.requests...)
}

func (c *Collector) Spans() []Span {
	c.mu.Lock()
	defer c.mu.Unlock()

	return append([]Span(nil), c.spans...)
}

func (c *Collector) receive(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	var export coltracepb.ExportTraceServiceRequest
	if err := proto.Unmarshal(body, &export); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.requests = append(c.requests, Request{Path: r.URL.Path, Header: r.Header.Clone()})

	for _, resourceSpans := range export.GetResourceSpans() {
		resource := values(resourceSpans.GetResource().GetAttributes())
		for _, scopeSpans := range resourceSpans.GetScopeSpans() {
			for _, span := range scopeSpans.GetSpans() {
				c.spans = append(c.spans, Span{
					Name: span.GetName(),

					TraceID: hex.EncodeToString(span.GetTraceId()),

					ParentSpanID: hex.EncodeToString(span.GetParentSpanId()),

					Attributes: values(span.GetAttributes()),

					Resource: resource,
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/x-protobuf")
	w.WriteHeader(http.StatusOK)
}

func values(attributes []*commonpb.KeyValue) map[string]string {
	result := make(map[string]string, len(attributes))
	for _, attr := range attributes {
		value := attr.GetValue()
		switch value.GetValue().(type) {
		case *commonpb.AnyValue_IntValue:
			result[attr.GetKey()] = strconv.FormatInt(value.GetIntValue(), 10)
		case *commonpb.AnyValue_BoolValue:
			result[attr.GetKey()] = strconv.FormatBool(value.GetBoolValue())
		case *commonpb.AnyValue_DoubleValue:
			result[attr.GetKey()] = strconv.FormatFloat(value.GetDoubleValue(), 'g', -1, 64)
		default:
			result[attr.GetKey()] = value.GetStringValue()
		}
	}

	return result
}
