package otel

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// WithCapability returns a copy of ctx with the given capability name set as
// OTel baggage. All spans created within the returned context will have the
// "gen_ai.capability.name" attribute set automatically.
func WithCapability(ctx context.Context, name string) context.Context {
	return withBaggageMember(ctx, "capability", name)
}

// WithStep returns a copy of ctx with the given step name set as OTel baggage.
// All spans created within the returned context will have the
// "gen_ai.step.name" attribute set automatically.
func WithStep(ctx context.Context, name string) context.Context {
	return withBaggageMember(ctx, "step", name)
}

func withBaggageMember(ctx context.Context, key, value string) context.Context {
	m, _ := baggage.NewMemberRaw(key, value)
	bag, _ := baggage.FromContext(ctx).SetMember(m)
	return baggage.ContextWithBaggage(ctx, bag)
}

// baggageSpanProcessor is a [sdktrace.SpanProcessor] that reads OTel baggage
// from the context and sets span attributes accordingly.
type baggageSpanProcessor struct{}

func (p *baggageSpanProcessor) OnStart(ctx context.Context, s sdktrace.ReadWriteSpan) {
	bag := baggage.FromContext(ctx)
	if v := bag.Member("capability").Value(); v != "" {
		s.SetAttributes(attribute.String("gen_ai.capability.name", v))
	}
	if v := bag.Member("step").Value(); v != "" {
		s.SetAttributes(attribute.String("gen_ai.step.name", v))
	}
}

func (*baggageSpanProcessor) OnEnd(sdktrace.ReadOnlySpan)      {}
func (*baggageSpanProcessor) Shutdown(context.Context) error   { return nil }
func (*baggageSpanProcessor) ForceFlush(context.Context) error { return nil }
