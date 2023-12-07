package otel

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// TraceExporter configures and returns a new exporter for OpenTelemetry spans.
func TraceExporter(ctx context.Context, dataset string, options ...Option) (trace.SpanExporter, error) {
	config := defaultTraceConfig()

	if err := populateAndValidateConfig(&config, options...); err != nil {
		return nil, err
	}

	u, err := config.BaseURL().Parse(config.APIEndpoint)
	if err != nil {
		return nil, fmt.Errorf("parse exporter url: %w", err)
	}

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(u.Host),
	}
	if u.Path != "" {
		opts = append(opts, otlptracehttp.WithURLPath(u.Path))
	}
	if u.Scheme == "http" {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	if config.Timeout > 0 {
		opts = append(opts, otlptracehttp.WithTimeout(config.Timeout))
	}

	headers := make(map[string]string)
	if config.Token() != "" {
		headers["Authorization"] = "Bearer " + config.Token()
	}
	if config.OrganizationID() != "" {
		headers["X-Axiom-Org-Id"] = config.OrganizationID()
	}
	if dataset != "" {
		headers["X-Axiom-Dataset"] = dataset
	}
	if len(headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(headers))
	}

	return otlptrace.New(ctx, otlptracehttp.NewClient(opts...))
}

// TracerProvider configures and returns a new OpenTelemetry tracer provider.
func TracerProvider(ctx context.Context, dataset, serviceName, serviceVersion string, options ...Option) (*trace.TracerProvider, error) {
	exporter, err := TraceExporter(ctx, dataset, options...)
	if err != nil {
		return nil, err
	}

	rs, err := resource.Merge(resource.Default(), resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
		semconv.ServiceVersionKey.String(serviceVersion),
		UserAgentAttribute(),
	))
	if err != nil {
		return nil, err
	}

	opts := []trace.TracerProviderOption{
		trace.WithBatcher(exporter, trace.WithMaxQueueSize(10*1024)),
		trace.WithResource(rs),
	}

	return trace.NewTracerProvider(opts...), nil
}

// InitTracing initializes OpenTelemetry tracing with the given service name,
// version and options. If initialization succeeds, the returned cleanup
// function must be called to shut down the tracer provider and flush any
// remaining spans. The error returned by the cleanup function must be checked,
// as well.
func InitTracing(ctx context.Context, dataset, serviceName, serviceVersion string, options ...Option) (func() error, error) {
	tracerProvider, err := TracerProvider(ctx, dataset, serviceName, serviceVersion, options...)
	if err != nil {
		return nil, err
	}

	otel.SetTracerProvider(tracerProvider)

	closeFunc := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		return tracerProvider.Shutdown(ctx)
	}

	return closeFunc, nil
}
