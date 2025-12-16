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

	// Keep in sync with
	// https://github.com/open-telemetry/opentelemetry-go/blob/main/sdk/resource/builtin.go#L16.
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	"github.com/axiomhq/axiom-go/internal/version"
)

var userAgent string

func init() {
	userAgent = "axiom-go"
	if v := version.Get(); v != "" {
		userAgent += fmt.Sprintf("/%s", v)
	}
}

// TraceExporter configures and returns a new exporter for OpenTelemetry spans.
func TraceExporter(ctx context.Context, dataset string, options ...TraceOption) (trace.SpanExporter, error) {
	config := defaultTraceConfig()

	// Apply supplied options.
	for _, option := range options {
		if option == nil {
			continue
		} else if err := option(&config); err != nil {
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
func TracerProvider(ctx context.Context, dataset, serviceName, serviceVersion string, options ...TraceOption) (*trace.TracerProvider, error) {
	exporter, err := TraceExporter(ctx, dataset, options...)
	if err != nil {
		return nil, err
	}

	rs, err := resource.Merge(resource.Default(), resource.NewWithAttributes(
		// HINT(lukasmalkmus): [resource.Merge] will use the schema URL from the
		// first resource, which is what we want to achieve here.
		"",
		semconv.ServiceNameKey.String(serviceName),
		semconv.ServiceVersionKey.String(serviceVersion),
		semconv.UserAgentOriginal(userAgent),
	))
	if err != nil {
		return nil, err
	}

	opts := []trace.TracerProviderOption{
		trace.WithBatcher(exporter, trace.WithMaxQueueSize(1024*10)),
		trace.WithResource(rs),
	}

	return trace.NewTracerProvider(opts...), nil
}

// InitTracing initializes OpenTelemetry tracing with the given service name,
// version and options. If initialization succeeds, the returned cleanup
// function must be called to shut down the tracer provider and flush any
// remaining spans. The error returned by the cleanup function must be checked,
// as well.
func InitTracing(ctx context.Context, dataset, serviceName, serviceVersion string, options ...TraceOption) (func() error, error) {
	tracerProvider, err := TracerProvider(ctx, dataset, serviceName, serviceVersion, options...)
	if err != nil {
		return nil, err
	}

	otel.SetTracerProvider(tracerProvider)

	closeFunc := func() error {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second*15)
		defer cancel()

		return tracerProvider.Shutdown(ctx)
	}

	return closeFunc, nil
}
