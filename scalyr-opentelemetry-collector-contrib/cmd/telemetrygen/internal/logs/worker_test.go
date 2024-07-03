// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package logs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/cmd/telemetrygen/internal/common"
)

const (
	telemetryAttrKeyOne   = "k1"
	telemetryAttrKeyTwo   = "k2"
	telemetryAttrValueOne = "v1"
	telemetryAttrValueTwo = "v2"
)

type mockExporter struct {
	logs []plog.Logs
}

func (m *mockExporter) export(logs plog.Logs) error {
	m.logs = append(m.logs, logs)
	return nil
}

func TestFixedNumberOfLogs(t *testing.T) {
	cfg := &Config{
		Config: common.Config{
			WorkerCount: 1,
		},
		NumLogs:        5,
		SeverityText:   "Info",
		SeverityNumber: 9,
	}

	exp := &mockExporter{}

	// test
	logger, _ := zap.NewDevelopment()
	require.NoError(t, Run(cfg, exp, logger))

	time.Sleep(1 * time.Second)

	// verify
	require.Len(t, exp.logs, 5)
}

func TestRateOfLogs(t *testing.T) {
	cfg := &Config{
		Config: common.Config{
			Rate:          10,
			TotalDuration: time.Second / 2,
			WorkerCount:   1,
		},
		SeverityText:   "Info",
		SeverityNumber: 9,
	}
	exp := &mockExporter{}

	// test
	require.NoError(t, Run(cfg, exp, zap.NewNop()))

	// verify
	// the minimum acceptable number of logs for the rate of 10/sec for half a second
	assert.True(t, len(exp.logs) >= 5, "there should have been 5 or more logs, had %d", len(exp.logs))
	// the maximum acceptable number of logs for the rate of 10/sec for half a second
	assert.True(t, len(exp.logs) <= 20, "there should have been less than 20 logs, had %d", len(exp.logs))
}

func TestUnthrottled(t *testing.T) {
	cfg := &Config{
		Config: common.Config{
			TotalDuration: 1 * time.Second,
			WorkerCount:   1,
		},
		SeverityText:   "Info",
		SeverityNumber: 9,
	}
	exp := &mockExporter{}

	// test
	logger, _ := zap.NewDevelopment()
	require.NoError(t, Run(cfg, exp, logger))

	assert.True(t, len(exp.logs) > 100, "there should have been more than 100 logs, had %d", len(exp.logs))
}

func TestCustomBody(t *testing.T) {
	cfg := &Config{
		Body:    "custom body",
		NumLogs: 1,
		Config: common.Config{
			WorkerCount: 1,
		},
		SeverityText:   "Info",
		SeverityNumber: 9,
	}
	exp := &mockExporter{}

	// test
	logger, _ := zap.NewDevelopment()
	require.NoError(t, Run(cfg, exp, logger))

	assert.Equal(t, "custom body", exp.logs[0].ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0).Body().AsString())
}

func TestLogsWithNoTelemetryAttributes(t *testing.T) {
	cfg := configWithNoAttributes(2, "custom body")

	exp := &mockExporter{}

	// test
	logger, _ := zap.NewDevelopment()
	require.NoError(t, Run(cfg, exp, logger))

	time.Sleep(1 * time.Second)

	// verify
	require.Len(t, exp.logs, 2)
	for _, log := range exp.logs {
		rlogs := log.ResourceLogs()
		for i := 0; i < rlogs.Len(); i++ {
			attrs := rlogs.At(i).ScopeLogs().At(0).LogRecords().At(0).Attributes()
			assert.Equal(t, 1, attrs.Len(), "shouldn't have more than 1 attribute")
		}
	}
}

func TestLogsWithOneTelemetryAttributes(t *testing.T) {
	qty := 1
	cfg := configWithOneAttribute(qty, "custom body")

	exp := &mockExporter{}

	// test
	logger, _ := zap.NewDevelopment()
	require.NoError(t, Run(cfg, exp, logger))

	time.Sleep(1 * time.Second)

	// verify
	require.Len(t, exp.logs, qty)
	for _, log := range exp.logs {
		rlogs := log.ResourceLogs()
		for i := 0; i < rlogs.Len(); i++ {
			attrs := rlogs.At(i).ScopeLogs().At(0).LogRecords().At(0).Attributes()
			assert.Equal(t, 2, attrs.Len(), "shouldn't have less than 2 attributes")

			val, ok := attrs.Get(telemetryAttrKeyOne)
			assert.Truef(t, ok, "there should be an attribute with key %s", telemetryAttrKeyOne)
			if ok {
				assert.EqualValues(t, val.AsString(), telemetryAttrValueOne)
			}

		}
	}
}

func TestLogsWithMultipleTelemetryAttributes(t *testing.T) {
	qty := 1
	cfg := configWithMultipleAttributes(qty, "custom body")

	exp := &mockExporter{}

	// test
	logger, _ := zap.NewDevelopment()
	require.NoError(t, Run(cfg, exp, logger))

	time.Sleep(1 * time.Second)

	// verify
	require.Len(t, exp.logs, qty)
	for _, log := range exp.logs {
		rlogs := log.ResourceLogs()
		for i := 0; i < rlogs.Len(); i++ {
			attrs := rlogs.At(i).ScopeLogs().At(0).LogRecords().At(0).Attributes()
			assert.Equal(t, 3, attrs.Len(), "shouldn't have less than 3 attributes")
		}
	}
}

func TestLogsWithTraceIDAndSpanID(t *testing.T) {
	qty := 1
	cfg := configWithOneAttribute(qty, "custom body")
	cfg.TraceID = "ae87dadd90e9935a4bc9660628efd569"
	cfg.SpanID = "5828fa4960140870"

	exp := &mockExporter{}

	// test
	logger, _ := zap.NewDevelopment()
	require.NoError(t, Run(cfg, exp, logger))

	// verify
	require.Len(t, exp.logs, qty)
	for _, log := range exp.logs {
		rlogs := log.ResourceLogs()
		for i := 0; i < rlogs.Len(); i++ {
			log := rlogs.At(i).ScopeLogs().At(0).LogRecords().At(0)
			assert.Equal(t, "ae87dadd90e9935a4bc9660628efd569", log.TraceID().String())
			assert.Equal(t, "5828fa4960140870", log.SpanID().String())
		}
	}
}

func configWithNoAttributes(qty int, body string) *Config {
	return &Config{
		Body:    body,
		NumLogs: qty,
		Config: common.Config{
			WorkerCount:         1,
			TelemetryAttributes: nil,
		},
		SeverityText:   "Info",
		SeverityNumber: 9,
	}
}

func configWithOneAttribute(qty int, body string) *Config {
	return &Config{
		Body:    body,
		NumLogs: qty,
		Config: common.Config{
			WorkerCount:         1,
			TelemetryAttributes: common.KeyValue{telemetryAttrKeyOne: telemetryAttrValueOne},
		},
		SeverityText:   "Info",
		SeverityNumber: 9,
	}
}

func configWithMultipleAttributes(qty int, body string) *Config {
	kvs := common.KeyValue{telemetryAttrKeyOne: telemetryAttrValueOne, telemetryAttrKeyTwo: telemetryAttrValueTwo}
	return &Config{
		Body:    body,
		NumLogs: qty,
		Config: common.Config{
			WorkerCount:         1,
			TelemetryAttributes: kvs,
		},
		SeverityText:   "Info",
		SeverityNumber: 9,
	}
}
