// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0
package datadog

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/DataDog/datadog-agent/pkg/trace/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/metric/metricdata/metricdatatest"
)

func setupMetricClient() (*metric.ManualReader, metrics.StatsClient, *metric.MeterProvider) {
	initializeOnce = sync.Once{}
	reader := metric.NewManualReader()
	meterProvider := metric.NewMeterProvider(metric.WithReader(reader))
	metricClient := InitializeMetricClient(meterProvider)
	return reader, metricClient, meterProvider
}

func TestNewMetricClient(t *testing.T) {
	_, metricClient, _ := setupMetricClient()
	assert.Equal(t, metrics.Client, metricClient)
}

func TestNewMetricClientComplex(t *testing.T) {
	_, _, meterProvider := setupMetricClient()
	metricClient2 := InitializeMetricClient(meterProvider)
	assert.Equal(t, metrics.Client, metricClient2)
}

func TestGauge(t *testing.T) {
	reader, metricClient, _ := setupMetricClient()

	err := metricClient.Gauge("test_gauge", 1, []string{"otlp:true", "service:otelcol"}, 1)
	assert.NoError(t, err)
	rm := metricdata.ResourceMetrics{}
	assert.NoError(t, reader.Collect(context.Background(), &rm))
	require.Len(t, rm.ScopeMetrics, 1)
	sm := rm.ScopeMetrics[0]
	require.Len(t, sm.Metrics, 1)
	got := sm.Metrics[0]
	want := metricdata.Metrics{
		Name: "test_gauge",
		Data: metricdata.Gauge[float64]{
			DataPoints: []metricdata.DataPoint[float64]{
				{Value: 1, Attributes: attribute.NewSet(attribute.String("otlp", "true"), attribute.String("service", "otelcol"))},
			},
		},
	}
	metricdatatest.AssertEqual(t, want, got, metricdatatest.IgnoreTimestamp())
}

func TestGaugeMultiple(t *testing.T) {
	reader, metricClient, _ := setupMetricClient()

	err := metricClient.Gauge("test_gauge", 1, []string{"otlp:true"}, 1)
	assert.NoError(t, err)
	err = metricClient.Gauge("test_gauge", 2, []string{"otlp:true"}, 1)
	assert.NoError(t, err)

	rm := metricdata.ResourceMetrics{}
	assert.NoError(t, reader.Collect(context.Background(), &rm))
	require.Len(t, rm.ScopeMetrics, 1)
	sm := rm.ScopeMetrics[0]
	require.Len(t, sm.Metrics, 1)
	got := sm.Metrics[0]
	want := metricdata.Metrics{
		Name: "test_gauge",
		Data: metricdata.Gauge[float64]{
			DataPoints: []metricdata.DataPoint[float64]{
				{Value: 2, Attributes: attribute.NewSet(attribute.String("otlp", "true"))},
			},
		},
	}
	metricdatatest.AssertEqual(t, want, got, metricdatatest.IgnoreTimestamp())
}

func TestCount(t *testing.T) {
	reader, metricClient, _ := setupMetricClient()

	err := metricClient.Count("test_count", 1, []string{"otlp:true", "service:otelcol"}, 1)
	assert.NoError(t, err)
	rm := metricdata.ResourceMetrics{}
	assert.NoError(t, reader.Collect(context.Background(), &rm))
	require.Len(t, rm.ScopeMetrics, 1)
	sm := rm.ScopeMetrics[0]
	require.Len(t, sm.Metrics, 1)
	got := sm.Metrics[0]
	want := metricdata.Metrics{
		Name: "test_count",
		Data: metricdata.Sum[int64]{
			Temporality: metricdata.CumulativeTemporality,
			IsMonotonic: true,
			DataPoints: []metricdata.DataPoint[int64]{
				{Value: 1, Attributes: attribute.NewSet(attribute.String("otlp", "true"), attribute.String("service", "otelcol"))},
			},
		},
	}
	metricdatatest.AssertEqual(t, want, got, metricdatatest.IgnoreTimestamp())

	err = metricClient.Count("test_count", 2, []string{"otlp:true", "service:otelcol"}, 1)
	assert.NoError(t, err)
	err = metricClient.Count("test_count", 3, []string{"otlp:true", "service:otelcol"}, 1)
	assert.NoError(t, err)
	err = metricClient.Count("test_count2", 3, []string{"otlp:true", "service:otelcol"}, 1)
	assert.NoError(t, err)
	assert.NoError(t, reader.Collect(context.Background(), &rm))
	require.Len(t, rm.ScopeMetrics, 1)
	sm = rm.ScopeMetrics[0]
	require.Len(t, sm.Metrics, 2)
	got = sm.Metrics[0]
	want = metricdata.Metrics{
		Name: "test_count",
		Data: metricdata.Sum[int64]{
			Temporality: metricdata.CumulativeTemporality,
			IsMonotonic: true,
			DataPoints: []metricdata.DataPoint[int64]{
				{Value: 6, Attributes: attribute.NewSet(attribute.String("otlp", "true"), attribute.String("service", "otelcol"))},
			},
		},
	}
	metricdatatest.AssertEqual(t, want, got, metricdatatest.IgnoreTimestamp())

	got = sm.Metrics[1]
	want = metricdata.Metrics{
		Name: "test_count2",
		Data: metricdata.Sum[int64]{
			Temporality: metricdata.CumulativeTemporality,
			IsMonotonic: true,
			DataPoints: []metricdata.DataPoint[int64]{
				{Value: 3, Attributes: attribute.NewSet(attribute.String("otlp", "true"), attribute.String("service", "otelcol"))},
			},
		},
	}
	metricdatatest.AssertEqual(t, want, got, metricdatatest.IgnoreTimestamp())
}

func TestHistogram(t *testing.T) {
	reader, metricClient, _ := setupMetricClient()

	err := metricClient.Histogram("test_histogram", 1, []string{"otlp:true", "service:otelcol"}, 1)
	assert.NoError(t, err)
	rm := metricdata.ResourceMetrics{}
	assert.NoError(t, reader.Collect(context.Background(), &rm))
	require.Len(t, rm.ScopeMetrics, 1)
	sm := rm.ScopeMetrics[0]
	require.Len(t, sm.Metrics, 1)
	got := sm.Metrics[0]
	want := metricdata.Metrics{
		Name: "test_histogram",
		Data: metricdata.Histogram[float64]{
			Temporality: metricdata.CumulativeTemporality,
			DataPoints: []metricdata.HistogramDataPoint[float64]{{
				Attributes:   attribute.NewSet(attribute.String("otlp", "true"), attribute.String("service", "otelcol")),
				Bounds:       []float64{0, 5, 10, 25, 50, 75, 100, 250, 500, 750, 1000, 2500, 5000, 7500, 10000},
				BucketCounts: []uint64{0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				Count:        1,
				Min:          metricdata.NewExtrema(1.0),
				Max:          metricdata.NewExtrema(1.0),
				Sum:          1,
			}},
		},
	}
	metricdatatest.AssertEqual(t, want, got, metricdatatest.IgnoreTimestamp())
}

func TestTiming(t *testing.T) {
	reader, metricClient, _ := setupMetricClient()

	err := metricClient.Timing("test_timing", time.Duration(1000000000), []string{"otlp:true", "service:otelcol"}, 1)
	assert.NoError(t, err)
	rm := metricdata.ResourceMetrics{}
	assert.NoError(t, reader.Collect(context.Background(), &rm))
	require.Len(t, rm.ScopeMetrics, 1)
	sm := rm.ScopeMetrics[0]
	require.Len(t, sm.Metrics, 1)
	got := sm.Metrics[0]
	want := metricdata.Metrics{
		Name: "test_timing",
		Data: metricdata.Histogram[float64]{
			Temporality: metricdata.CumulativeTemporality,
			DataPoints: []metricdata.HistogramDataPoint[float64]{{
				Attributes:   attribute.NewSet(attribute.String("otlp", "true"), attribute.String("service", "otelcol")),
				Bounds:       []float64{0, 5, 10, 25, 50, 75, 100, 250, 500, 750, 1000, 2500, 5000, 7500, 10000},
				BucketCounts: []uint64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0},
				Count:        1,
				Min:          metricdata.NewExtrema(1000.0),
				Max:          metricdata.NewExtrema(1000.0),
				Sum:          1000,
			}},
		},
	}
	metricdatatest.AssertEqual(t, want, got, metricdatatest.IgnoreTimestamp())
}
