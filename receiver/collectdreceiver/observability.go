// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package collectdreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/collectdreceiver"

import (
	"context"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

func init() {
	_ = view.Register(
		viewInvalidRequests,
		viewRequestsReceived,
		viewMetricsReceived,
		viewEventsReceived,
		viewBlankDefaultAttrs,
	)
}

var (
	mErrors            = stats.Int64("otelcol/collectd/errors", "Errors encountered during processing of collectd requests", "1")
	mRequestsReceived  = stats.Int64("otelcol/collectd/requests_received", "Number of total requests received", "1")
	mMetricsReceived   = stats.Int64("otelcol/collectd/metrics_received", "Number of metrics received", "1")
	mEventsReceived    = stats.Int64("otelcol/collectd/events_received", "Number of events received", "1")
	mBlankDefaultAttrs = stats.Int64("otelcol/collectd/blank_default_attrs", "Number of blank default attributes received", "1")
)

var viewInvalidRequests = &view.View{
	Name:        mErrors.Name(),
	Description: mErrors.Description(),
	Measure:     mErrors,
	Aggregation: view.Sum(),
}

var viewRequestsReceived = &view.View{
	Name:        mRequestsReceived.Name(),
	Description: mRequestsReceived.Description(),
	Measure:     mRequestsReceived,
	Aggregation: view.Sum(),
}

var viewMetricsReceived = &view.View{
	Name:        mMetricsReceived.Name(),
	Description: mMetricsReceived.Description(),
	Measure:     mMetricsReceived,
	Aggregation: view.Sum(),
}

var viewEventsReceived = &view.View{
	Name:        mEventsReceived.Name(),
	Description: mEventsReceived.Description(),
	Measure:     mEventsReceived,
	Aggregation: view.Sum(),
}

var viewBlankDefaultAttrs = &view.View{
	Name:        mBlankDefaultAttrs.Name(),
	Description: mBlankDefaultAttrs.Description(),
	Measure:     mBlankDefaultAttrs,
	Aggregation: view.Sum(),
}

func recordRequestErrors() {
	stats.Record(context.Background(), mErrors.M(int64(1)))
}

func recordRequestReceived() {
	stats.Record(context.Background(), mRequestsReceived.M(int64(1)))
}

func recordMetricsReceived() {
	stats.Record(context.Background(), mMetricsReceived.M(int64(1)))
}

func recordEventsReceived() {
	stats.Record(context.Background(), mEventsReceived.M(int64(1)))
}

func recordDefaultBlankAttrs() {
	stats.Record(context.Background(), mBlankDefaultAttrs.M(int64(1)))
}
