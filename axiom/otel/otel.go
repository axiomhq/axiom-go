package otel

import (
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	// Keep in sync with https://github.com/open-telemetry/opentelemetry-go/blob/main/sdk/resource/builtin.go#L25.
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"

	"github.com/axiomhq/axiom-go/internal/version"
)

var userAgent string

func init() {
	userAgent = "axiom-go"
	if v := version.Get(); v != "" {
		userAgent += fmt.Sprintf("/%s", v)
	}
}

// UserAgentAttribute returns a new OpenTelemetry axiom-go user agent attribute.
func UserAgentAttribute() attribute.KeyValue {
	return semconv.UserAgentOriginal(userAgent)
}
