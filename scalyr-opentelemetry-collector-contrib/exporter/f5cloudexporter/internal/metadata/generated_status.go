// Code generated by mdatagen. DO NOT EDIT.

package metadata

import (
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

var (
	Type = component.MustNewType("f5cloud")
)

const (
	TracesStability  = component.StabilityLevelDeprecated
	MetricsStability = component.StabilityLevelDeprecated
	LogsStability    = component.StabilityLevelDeprecated
)

func Meter(settings component.TelemetrySettings) metric.Meter {
	return settings.MeterProvider.Meter("otelcol/f5cloud")
}

func Tracer(settings component.TelemetrySettings) trace.Tracer {
	return settings.TracerProvider.Tracer("otelcol/f5cloud")
}
