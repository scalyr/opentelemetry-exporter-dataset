// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:generate mdatagen metadata.yaml

package coralogixexporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/coralogixexporter"

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configcompression"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/consumer"
	exp "go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/coralogixexporter/internal/metadata"
)

// NewFactory by Coralogix
func NewFactory() exp.Factory {
	return exp.NewFactory(
		metadata.Type,
		createDefaultConfig,
		exp.WithTraces(createTraceExporter, metadata.TracesStability),
		exp.WithMetrics(createMetricsExporter, metadata.MetricsStability),
		exp.WithLogs(createLogsExporter, metadata.LogsStability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		QueueSettings:   exporterhelper.NewDefaultQueueSettings(),
		RetrySettings:   exporterhelper.NewDefaultRetrySettings(),
		TimeoutSettings: exporterhelper.NewDefaultTimeoutSettings(),
		DomainSettings: configgrpc.GRPCClientSettings{
			Compression: configcompression.Gzip,
		},
		GRPCClientSettings: configgrpc.GRPCClientSettings{
			Endpoint: "https://",
		},
		// Traces GRPC client
		Traces: configgrpc.GRPCClientSettings{
			Endpoint:    "https://",
			Compression: configcompression.Gzip,
		},
		Metrics: configgrpc.GRPCClientSettings{
			Endpoint: "https://",
			// Default to gzip compression
			Compression:     configcompression.Gzip,
			WriteBufferSize: 512 * 1024,
		},
		Logs: configgrpc.GRPCClientSettings{
			Endpoint:    "https://",
			Compression: configcompression.Gzip,
		},
		PrivateKey: "",
		AppName:    "",
	}
}

func createTraceExporter(ctx context.Context, set exp.CreateSettings, config component.Config) (exp.Traces, error) {
	cfg := config.(*Config)

	exporter, err := newTracesExporter(cfg, set)
	if err != nil {
		return nil, err
	}

	return exporterhelper.NewTracesExporter(
		ctx,
		set,
		config,
		exporter.pushTraces,
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: true}),
		exporterhelper.WithTimeout(cfg.TimeoutSettings),
		exporterhelper.WithRetry(cfg.RetrySettings),
		exporterhelper.WithQueue(cfg.QueueSettings),
		exporterhelper.WithStart(exporter.start),
		exporterhelper.WithShutdown(exporter.shutdown),
	)
}

func createMetricsExporter(
	ctx context.Context,
	set exp.CreateSettings,
	cfg component.Config,
) (exp.Metrics, error) {
	oce, err := newMetricsExporter(cfg, set)
	if err != nil {
		return nil, err
	}
	oCfg := cfg.(*Config)
	return exporterhelper.NewMetricsExporter(
		ctx,
		set,
		cfg,
		oce.pushMetrics,
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: true}),
		exporterhelper.WithTimeout(oCfg.TimeoutSettings),
		exporterhelper.WithRetry(oCfg.RetrySettings),
		exporterhelper.WithQueue(oCfg.QueueSettings),
		exporterhelper.WithStart(oce.start),
		exporterhelper.WithShutdown(oce.shutdown),
	)
}

func createLogsExporter(
	ctx context.Context,
	set exp.CreateSettings,
	cfg component.Config,
) (exp.Logs, error) {
	oce, err := newLogsExporter(cfg, set)
	if err != nil {
		return nil, err
	}
	oCfg := cfg.(*Config)
	return exporterhelper.NewLogsExporter(
		ctx,
		set,
		cfg,
		oce.pushLogs,
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: true}),
		exporterhelper.WithTimeout(oCfg.TimeoutSettings),
		exporterhelper.WithRetry(oCfg.RetrySettings),
		exporterhelper.WithQueue(oCfg.QueueSettings),
		exporterhelper.WithStart(oce.start),
		exporterhelper.WithShutdown(oce.shutdown),
	)
}
