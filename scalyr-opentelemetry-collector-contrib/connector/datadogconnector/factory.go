// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:generate mdatagen metadata.yaml

package datadogconnector // import "github.com/open-telemetry/opentelemetry-collector-contrib/connector/datadogconnector"

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/consumer"

	"github.com/open-telemetry/opentelemetry-collector-contrib/connector/datadogconnector/internal/metadata"
	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/datadog"
)

// NewFactory creates a factory for tailtracer connector.
func NewFactory() connector.Factory {
	//  OTel connector factory to make a factory for connectors
	return connector.NewFactory(
		metadata.Type,
		createDefaultConfig,
		connector.WithTracesToMetrics(createTracesToMetricsConnector, metadata.TracesToMetricsStability),
		connector.WithTracesToTraces(createTracesToTracesConnector, metadata.TracesToTracesStability))
}

func createDefaultConfig() component.Config {
	return &Config{
		Traces: TracesConfig{
			IgnoreResources: []string{},
			TraceBuffer:     1000,
		},
	}
}

// defines the consumer type of the connector
// we want to consume traces and export metrics therefore define nextConsumer as metrics, consumer is the next component in the pipeline
func createTracesToMetricsConnector(_ context.Context, params connector.CreateSettings, cfg component.Config, nextConsumer consumer.Metrics) (connector.Traces, error) {
	initializeMetricsClient(params)
	c, err := newTraceToMetricConnector(params.TelemetrySettings, cfg, nextConsumer)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func createTracesToTracesConnector(_ context.Context, params connector.CreateSettings, _ component.Config, nextConsumer consumer.Traces) (connector.Traces, error) {
	initializeMetricsClient(params)
	return newTraceToTraceConnector(params.Logger, nextConsumer), nil
}

func initializeMetricsClient(params connector.CreateSettings) {
	datadog.InitializeMetricClient(params.MeterProvider)
}
