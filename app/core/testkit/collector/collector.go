package collector

import (
	"compress/gzip"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	colmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
	"google.golang.org/protobuf/proto"
)

type Span struct {
	Name string

	TraceID string

	SpanID string

	ParentSpanID string

	Attributes map[string]string

	Resource map[string]string
}

type Log struct {
	Body string

	Severity string

	TraceID string

	SpanID string

	Attributes map[string]string

	Resource map[string]string
}

type Point struct {
	Metric string

	Attributes map[string]string

	Count uint64

	Value float64

	Delta bool

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

	points []Point

	logs []Log
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

func (c *Collector) Logs() []Log {
	c.mu.Lock()
	defer c.mu.Unlock()

	return append([]Log(nil), c.logs...)
}

func (c *Collector) Latest(metric string) []Point {
	c.mu.Lock()
	defer c.mu.Unlock()

	var latest []Point

	series := map[string]int{}

	for _, point := range c.points {
		if point.Metric != metric {
			continue
		}

		key := fmt.Sprint(point.Resource, point.Attributes)
		if index, found := series[key]; found {
			latest[index] = point

			continue
		}

		series[key] = len(latest)
		latest = append(latest, point)
	}

	return latest
}

func (c *Collector) Metrics() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	var names []string

	seen := map[string]bool{}
	for _, point := range c.points {
		if seen[point.Metric] == false {
			seen[point.Metric] = true
			names = append(names, point.Metric)
		}
	}

	return names
}

func (c *Collector) receive(w http.ResponseWriter, r *http.Request) {
	body, err := read(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	switch {
	case strings.HasSuffix(r.URL.Path, "/v1/metrics"):
		err = c.receiveMetrics(body)
	case strings.HasSuffix(r.URL.Path, "/v1/logs"):
		err = c.receiveLogs(body)
	default:
		err = c.receiveSpans(body)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	c.mu.Lock()
	c.requests = append(c.requests, Request{Path: r.URL.Path, Header: r.Header.Clone()})
	c.mu.Unlock()

	w.Header().Set("Content-Type", "application/x-protobuf")
	w.WriteHeader(http.StatusOK)
}

func read(r *http.Request) ([]byte, error) {
	if r.Header.Get("Content-Encoding") != "gzip" {
		return io.ReadAll(r.Body)
	}

	unzipped, err := gzip.NewReader(r.Body)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(unzipped)
}

func (c *Collector) receiveSpans(body []byte) error {
	var export coltracepb.ExportTraceServiceRequest
	if err := proto.Unmarshal(body, &export); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, resourceSpans := range export.GetResourceSpans() {
		resource := values(resourceSpans.GetResource().GetAttributes())
		for _, scopeSpans := range resourceSpans.GetScopeSpans() {
			for _, span := range scopeSpans.GetSpans() {
				c.spans = append(c.spans, Span{
					Name: span.GetName(),

					TraceID: hex.EncodeToString(span.GetTraceId()),

					SpanID: hex.EncodeToString(span.GetSpanId()),

					ParentSpanID: hex.EncodeToString(span.GetParentSpanId()),

					Attributes: values(span.GetAttributes()),

					Resource: resource,
				})
			}
		}
	}

	return nil
}

func (c *Collector) receiveLogs(body []byte) error {
	var export collogspb.ExportLogsServiceRequest
	if err := proto.Unmarshal(body, &export); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, resourceLogs := range export.GetResourceLogs() {
		resource := values(resourceLogs.GetResource().GetAttributes())
		for _, scopeLogs := range resourceLogs.GetScopeLogs() {
			for _, record := range scopeLogs.GetLogRecords() {
				c.logs = append(c.logs, Log{
					Body: record.GetBody().GetStringValue(),

					Severity: record.GetSeverityText(),

					TraceID: hex.EncodeToString(record.GetTraceId()),

					SpanID: hex.EncodeToString(record.GetSpanId()),

					Attributes: values(record.GetAttributes()),

					Resource: resource,
				})
			}
		}
	}

	return nil
}

func (c *Collector) receiveMetrics(body []byte) error {
	var export colmetricspb.ExportMetricsServiceRequest
	if err := proto.Unmarshal(body, &export); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, resourceMetrics := range export.GetResourceMetrics() {
		resource := values(resourceMetrics.GetResource().GetAttributes())
		for _, scopeMetrics := range resourceMetrics.GetScopeMetrics() {
			for _, metric := range scopeMetrics.GetMetrics() {
				c.points = append(c.points, points(metric, resource)...)
			}
		}
	}

	return nil
}

func points(metric *metricspb.Metric, resource map[string]string) []Point {
	var found []Point

	delta := metricspb.AggregationTemporality_AGGREGATION_TEMPORALITY_DELTA

	for _, point := range metric.GetHistogram().GetDataPoints() {
		found = append(found, Point{
			Metric: metric.GetName(),

			Attributes: values(point.GetAttributes()),

			Count: point.GetCount(),

			Delta: metric.GetHistogram().GetAggregationTemporality() == delta,

			Resource: resource,
		})
	}

	numbers := append(metric.GetSum().GetDataPoints(), metric.GetGauge().GetDataPoints()...)
	for _, point := range numbers {
		value := point.GetAsDouble()
		if _, integer := point.GetValue().(*metricspb.NumberDataPoint_AsInt); integer {
			value = float64(point.GetAsInt())
		}

		found = append(found, Point{
			Metric: metric.GetName(),

			Attributes: values(point.GetAttributes()),

			Value: value,

			Delta: metric.GetSum().GetAggregationTemporality() == delta,

			Resource: resource,
		})
	}

	return found
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
