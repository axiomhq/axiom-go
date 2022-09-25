// Package otel provides helpers for using OpenTelemetry with Axiom.
//
// Different levels of helpers are available, from just setting up tracing to
// getting access to lower level components to costumize tracing or integrate
// with existing OpenTelemetry setups:
//
//   - InitTracing: Initializes OpenTelemetry and sets the global tracer
//     prodiver so the official OpenTelemetry Go SDK can be used to get a tracer
//     and instrument code. Sane defaults for the tracer provider are applied.
//   - TracerProvider: Configures and returns a new OpenTelemetry tracer
//     provider but does not set it as the global tracer provider.
//   - TraceExporter: Configures and returns a new OpenTelemetry trace exporter.
//     This sets up the exporter that sends traces to Axiom but allows for a
//     more advanced setup of the tracer provider.
package otel
