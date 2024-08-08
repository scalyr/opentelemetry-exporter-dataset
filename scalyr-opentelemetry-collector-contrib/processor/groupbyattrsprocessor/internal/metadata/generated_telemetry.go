// Code generated by mdatagen. DO NOT EDIT.

package metadata

import (
	"errors"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configtelemetry"
)

func Meter(settings component.TelemetrySettings) metric.Meter {
	return settings.MeterProvider.Meter("otelcol/groupbyattrs")
}

func Tracer(settings component.TelemetrySettings) trace.Tracer {
	return settings.TracerProvider.Tracer("otelcol/groupbyattrs")
}

// TelemetryBuilder provides an interface for components to report telemetry
// as defined in metadata and user config.
type TelemetryBuilder struct {
	meter                                     metric.Meter
	ProcessorGroupbyattrsLogGroups            metric.Int64Histogram
	ProcessorGroupbyattrsMetricGroups         metric.Int64Histogram
	ProcessorGroupbyattrsNumGroupedLogs       metric.Int64Counter
	ProcessorGroupbyattrsNumGroupedMetrics    metric.Int64Counter
	ProcessorGroupbyattrsNumGroupedSpans      metric.Int64Counter
	ProcessorGroupbyattrsNumNonGroupedLogs    metric.Int64Counter
	ProcessorGroupbyattrsNumNonGroupedMetrics metric.Int64Counter
	ProcessorGroupbyattrsNumNonGroupedSpans   metric.Int64Counter
	ProcessorGroupbyattrsSpanGroups           metric.Int64Histogram
	level                                     configtelemetry.Level
}

// telemetryBuilderOption applies changes to default builder.
type telemetryBuilderOption func(*TelemetryBuilder)

// WithLevel sets the current telemetry level for the component.
func WithLevel(lvl configtelemetry.Level) telemetryBuilderOption {
	return func(builder *TelemetryBuilder) {
		builder.level = lvl
	}
}

// NewTelemetryBuilder provides a struct with methods to update all internal telemetry
// for a component
func NewTelemetryBuilder(settings component.TelemetrySettings, options ...telemetryBuilderOption) (*TelemetryBuilder, error) {
	builder := TelemetryBuilder{level: configtelemetry.LevelBasic}
	for _, op := range options {
		op(&builder)
	}
	var err, errs error
	if builder.level >= configtelemetry.LevelBasic {
		builder.meter = Meter(settings)
	} else {
		builder.meter = noop.Meter{}
	}
	builder.ProcessorGroupbyattrsLogGroups, err = builder.meter.Int64Histogram(
		"otelcol_processor_groupbyattrs_log_groups",
		metric.WithDescription("Distribution of groups extracted for logs"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.ProcessorGroupbyattrsMetricGroups, err = builder.meter.Int64Histogram(
		"otelcol_processor_groupbyattrs_metric_groups",
		metric.WithDescription("Distribution of groups extracted for metrics"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.ProcessorGroupbyattrsNumGroupedLogs, err = builder.meter.Int64Counter(
		"otelcol_processor_groupbyattrs_num_grouped_logs",
		metric.WithDescription("Number of logs that had attributes grouped"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.ProcessorGroupbyattrsNumGroupedMetrics, err = builder.meter.Int64Counter(
		"otelcol_processor_groupbyattrs_num_grouped_metrics",
		metric.WithDescription("Number of metrics that had attributes grouped"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.ProcessorGroupbyattrsNumGroupedSpans, err = builder.meter.Int64Counter(
		"otelcol_processor_groupbyattrs_num_grouped_spans",
		metric.WithDescription("Number of spans that had attributes grouped"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.ProcessorGroupbyattrsNumNonGroupedLogs, err = builder.meter.Int64Counter(
		"otelcol_processor_groupbyattrs_num_non_grouped_logs",
		metric.WithDescription("Number of logs that did not have attributes grouped"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.ProcessorGroupbyattrsNumNonGroupedMetrics, err = builder.meter.Int64Counter(
		"otelcol_processor_groupbyattrs_num_non_grouped_metrics",
		metric.WithDescription("Number of metrics that did not have attributes grouped"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.ProcessorGroupbyattrsNumNonGroupedSpans, err = builder.meter.Int64Counter(
		"otelcol_processor_groupbyattrs_num_non_grouped_spans",
		metric.WithDescription("Number of spans that did not have attributes grouped"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.ProcessorGroupbyattrsSpanGroups, err = builder.meter.Int64Histogram(
		"otelcol_processor_groupbyattrs_span_groups",
		metric.WithDescription("Distribution of groups extracted for spans"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	return &builder, errs
}
