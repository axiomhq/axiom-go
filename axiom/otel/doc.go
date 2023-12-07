// Package otel provides helpers for using [OpenTelemetry] with Axiom.
//
// Usage:
//
//	import "github.com/axiomhq/axiom-go/axiom/otel"
//
// Different levels of helpers are available, from just setting up
// instrumentation to getting access to lower level components to costumize
// instrumentation or integrate with existing OpenTelemetry setups:
//
//   - [InitMetrics]/[InitTracing]: Initializes OpenTelemetry and sets the
//     global meter/tracer prodiver so the official OpenTelemetry Go SDK can be
//     used to get a meter/tracer and instrument code. Sane defaults for the
//     providers are applied.
//   - [MeterProvider]/[TracerProvider]: Configures and returns a new
//     OpenTelemetry meter/tracer provider but does not set it as the global
//     meter/tracer provider.
//   - [MetricExporter]/[TraceExporter]: Configures and returns a new
//     OpenTelemetry metric/trace exporter. This sets up the exporter that sends
//     metrics/traces to Axiom but allows for a more advanced setup of the
//     meter/tracer provider.
//
// If you wish for traces to propagate beyond the current process, you need to
// set the global propagator to the OpenTelemetry trace context propagator. This
// can be done by calling:
//
//	import (
//	    "go.opentelemetry.io/otel"
//	    "go.opentelemetry.io/otel/propagation"
//	)
//	// ...
//	otel.SetTextMapPropagator(propagation.TraceContext{})
//	// or
//	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
//
// Refer to https://opentelemetry.io/docs/instrumentation/go/manual/#propagators-and-context
// for more information.
//
// [OpenTelemetry]: https://opentelemetry.io
package otel
