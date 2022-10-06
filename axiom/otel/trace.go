package otel

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"

	"github.com/axiomhq/axiom-go/internal/version"
)

// UserAgentAttribute returns a new OpenTelemetry axiom-go user agent attribute.
func UserAgentAttribute() attribute.KeyValue {
	userAgent := "axiom-go"
	if v := version.Get(); v != "" {
		userAgent += fmt.Sprintf("/%s", v)
	}
	return semconv.HTTPUserAgentKey.String(userAgent)
}

// TraceExporter configures and returns a new exporter for OpenTelemetry spans.
func TraceExporter(ctx context.Context, options ...TraceOption) (trace.SpanExporter, error) {
	config := defaultTraceConfig()

	// Apply supplied options.
	for _, option := range options {
		if err := option(&config); err != nil {
			return nil, err
		}
	}

	// Make sure to populate remaining fields from the environment, if not
	// explicitly disabled.
	if !config.NoEnv {
		if err := config.IncorporateEnvironment(); err != nil {
			return nil, err
		}
	}

	if err := config.Validate(); err != nil {
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
	if config.AccessToken() != "" || config.OrganizationID() != "" {
		headers := make(map[string]string, 2)

		if config.AccessToken() != "" {
			headers["Authorization"] = "Bearer " + config.AccessToken()
		}
		if config.OrganizationID() != "" {
			headers["X-Axiom-Org-Id"] = config.OrganizationID()
		}

		if len(headers) > 0 {
			opts = append(opts, otlptracehttp.WithHeaders(headers))
		}
	}
	if config.Timeout > 0 {
		opts = append(opts, otlptracehttp.WithTimeout(config.Timeout))
	}

	return otlptrace.New(ctx, otlptracehttp.NewClient(opts...))
}

// TracerProvider configures and returns a new OpenTelemetry tracer provider.
func TracerProvider(ctx context.Context, serviceName, serviceVersion string, options ...TraceOption) (*trace.TracerProvider, error) {
	exporter, err := TraceExporter(ctx, options...)
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
func InitTracing(ctx context.Context, serviceName, serviceVersion string, options ...TraceOption) (func() error, error) {
	tracerProvider, err := TracerProvider(ctx, serviceName, serviceVersion, options...)
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
