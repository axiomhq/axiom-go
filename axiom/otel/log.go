package otel

import (
	"context"
	"fmt"
	"time"

	logglobal "go.opentelemetry.io/otel/log/global"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/sdk/resource"

	sdklog "go.opentelemetry.io/otel/sdk/log"

	// Keep in sync with
	// https://github.com/open-telemetry/opentelemetry-go/blob/main/sdk/resource/builtin.go#L16.
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
)

// LogExporter configures and returns a new exporter for OpenTelemetry logs.
func LogExporter(ctx context.Context, dataset string, options ...Option) (sdklog.Exporter, error) {
	config := defaultExporterConfig("/v1/logs")

	for _, option := range options {
		if option == nil {
			continue
		} else if err := option(&config); err != nil {
			return nil, err
		}
	}

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

	opts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(u.Host),
	}
	if u.Path != "" {
		opts = append(opts, otlploghttp.WithURLPath(u.Path))
	}
	if u.Scheme == "http" {
		opts = append(opts, otlploghttp.WithInsecure())
	}
	if config.Timeout > 0 {
		opts = append(opts, otlploghttp.WithTimeout(config.Timeout))
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
		opts = append(opts, otlploghttp.WithHeaders(headers))
	}

	return otlploghttp.New(ctx, opts...)
}

// LoggerProvider configures and returns a new OpenTelemetry LoggerProvider.
func LoggerProvider(ctx context.Context, dataset, serviceName, serviceVersion string, options ...Option) (*sdklog.LoggerProvider, error) {
	exporter, err := LogExporter(ctx, dataset, options...)
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

	opts := []sdklog.LoggerProviderOption{
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
		sdklog.WithResource(rs),
	}

	return sdklog.NewLoggerProvider(opts...), nil
}

// InitLogging initializes OpenTelemetry logging with the given service name,
// version and options. If initialization succeeds, the returned cleanup
// function must be called to shut down the logger provider and flush any
// remaining log records. The error returned by the cleanup function must be
// checked, as well.
func InitLogging(ctx context.Context, dataset, serviceName, serviceVersion string, options ...Option) (func() error, error) {
	loggerProvider, err := LoggerProvider(ctx, dataset, serviceName, serviceVersion, options...)
	if err != nil {
		return nil, err
	}

	logglobal.SetLoggerProvider(loggerProvider)

	closeFunc := func() error {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second*15)
		defer cancel()

		return loggerProvider.Shutdown(ctx)
	}

	return closeFunc, nil
}
