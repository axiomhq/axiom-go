package otel

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// MetricExporter configures and returns a new exporter for OpenTelemetry spans.
func MetricExporter(ctx context.Context, dataset string, options ...Option) (metric.Exporter, error) {
	config := defaultMetricConfig()

	if err := populateAndValidateConfig(&config, options...); err != nil {
		return nil, err
	}

	u, err := config.BaseURL().Parse(config.APIEndpoint)
	if err != nil {
		return nil, fmt.Errorf("parse exporter url: %w", err)
	}

	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(u.Host),
	}
	if u.Path != "" {
		opts = append(opts, otlpmetrichttp.WithURLPath(u.Path))
	}
	if u.Scheme == "http" {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}
	if config.Timeout > 0 {
		opts = append(opts, otlpmetrichttp.WithTimeout(config.Timeout))
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
		opts = append(opts, otlpmetrichttp.WithHeaders(headers))
	}

	return otlpmetrichttp.New(ctx, opts...)
}

// MeterProvider configures and returns a new OpenTelemetry meter provider.
func MeterProvider(ctx context.Context, dataset, serviceName, serviceVersion string, options ...Option) (*metric.MeterProvider, error) {
	exporter, err := MetricExporter(ctx, dataset, options...)
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

	opts := []metric.Option{
		metric.WithReader(metric.NewPeriodicReader(
			exporter,
			metric.WithInterval(time.Second*5), // FIXME(lukasmalkmus): Just for testing!
			metric.WithTimeout(time.Second*5),  // FIXME(lukasmalkmus): Just for testing!
		)),
		metric.WithResource(rs),
	}

	return metric.NewMeterProvider(opts...), nil
}

// InitMetrics initializes OpenTelemetry metrics with the given service name,
// version and options. If initialization succeeds, the returned cleanup
// function must be called to shut down the meter provider and flush any
// remaining datapoints. The error returned by the cleanup function must be
// checked, as well.
func InitMetrics(ctx context.Context, dataset, serviceName, serviceVersion string, options ...Option) (func() error, error) {
	meterProvider, err := MeterProvider(ctx, dataset, serviceName, serviceVersion, options...)
	if err != nil {
		return nil, err
	}

	otel.SetMeterProvider(meterProvider)

	closeFunc := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		return meterProvider.Shutdown(ctx)
	}

	return closeFunc, nil
}
