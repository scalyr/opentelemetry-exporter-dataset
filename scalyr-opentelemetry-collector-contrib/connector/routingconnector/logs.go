// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package routingconnector // import "github.com/open-telemetry/opentelemetry-collector-contrib/connector/routingconnector"

import (
	"context"
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottlresource"
)

type logsConnector struct {
	component.StartFunc
	component.ShutdownFunc

	logger *zap.Logger
	config *Config
	router *router[consumer.Logs]
}

func newLogsConnector(
	set connector.CreateSettings,
	config component.Config,
	logs consumer.Logs,
) (*logsConnector, error) {
	cfg := config.(*Config)

	lr, ok := logs.(connector.LogsRouter)
	if !ok {
		return nil, errUnexpectedConsumer
	}

	r, err := newRouter(
		cfg.Table,
		cfg.DefaultPipelines,
		lr.Consumer,
		set.TelemetrySettings)

	if err != nil {
		return nil, err
	}

	return &logsConnector{
		logger: set.TelemetrySettings.Logger,
		config: cfg,
		router: r,
	}, nil
}

func (c *logsConnector) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

func (c *logsConnector) ConsumeLogs(ctx context.Context, ld plog.Logs) error {
	// routingEntry is used to group plog.ResourceLogs that are routed to
	// the same set of exporters.
	// This way we're not ending up with all the logs split up which would cause
	// higher CPU usage.
	groups := make(map[consumer.Logs]plog.Logs)
	var errs error

	for i := 0; i < ld.ResourceLogs().Len(); i++ {
		rlogs := ld.ResourceLogs().At(i)
		rtx := ottlresource.NewTransformContext(rlogs.Resource())

		noRoutesMatch := true
		for _, route := range c.router.routes {
			_, isMatch, err := route.statement.Execute(ctx, rtx)
			if err != nil {
				if c.config.ErrorMode == ottl.PropagateError {
					return err
				}
				c.group(groups, c.router.defaultConsumer, rlogs)
				continue
			}
			if isMatch {
				noRoutesMatch = false
				c.group(groups, route.consumer, rlogs)
			}

		}

		if noRoutesMatch {
			// no route conditions are matched, add resource logs to default exporters group
			c.group(groups, c.router.defaultConsumer, rlogs)
		}
	}
	for consumer, group := range groups {
		errs = errors.Join(errs, consumer.ConsumeLogs(ctx, group))
	}
	return errs
}

func (c *logsConnector) group(
	groups map[consumer.Logs]plog.Logs,
	consumer consumer.Logs,
	logs plog.ResourceLogs,
) {
	if consumer == nil {
		return
	}
	group, ok := groups[consumer]
	if !ok {
		group = plog.NewLogs()
	}
	logs.CopyTo(group.ResourceLogs().AppendEmpty())
	groups[consumer] = group
}
