// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package k8seventsreceiver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

func TestK8sEventToLogData(t *testing.T) {
	k8sEvent := getEvent()

	ld := k8sEventToLogData(zap.NewNop(), k8sEvent)
	rl := ld.ResourceLogs().At(0)
	resourceAttrs := rl.Resource().Attributes()
	lr := rl.ScopeLogs().At(0)
	attrs := lr.LogRecords().At(0).Attributes()
	assert.Equal(t, ld.ResourceLogs().Len(), 1)
	assert.Equal(t, resourceAttrs.Len(), 7)
	assert.Equal(t, attrs.Len(), 7)

	// Count attribute will not be present in the LogData
	k8sEvent.Count = 0
	ld = k8sEventToLogData(zap.NewNop(), k8sEvent)
	assert.Equal(t, ld.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0).Attributes().Len(), 6)
}

func TestK8sEventToLogDataWithApiAndResourceVersion(t *testing.T) {
	k8sEvent := getEvent()

	ld := k8sEventToLogData(zap.NewNop(), k8sEvent)
	attrs := ld.ResourceLogs().At(0).Resource().Attributes()
	attr, ok := attrs.Get("k8s.object.api_version")
	assert.Equal(t, true, ok)
	assert.Equal(t, "v1", attr.AsString())

	attr, ok = attrs.Get("k8s.object.resource_version")
	assert.Equal(t, true, ok)
	assert.Equal(t, "", attr.AsString())

	// add ResourceVersion
	k8sEvent.InvolvedObject.ResourceVersion = "7387066320"
	ld = k8sEventToLogData(zap.NewNop(), k8sEvent)
	attrs = ld.ResourceLogs().At(0).Resource().Attributes()
	attr, ok = attrs.Get("k8s.object.resource_version")
	assert.Equal(t, true, ok)
	assert.Equal(t, "7387066320", attr.AsString())
}

func TestUnknownSeverity(t *testing.T) {
	k8sEvent := getEvent()
	k8sEvent.Type = "Unknown"

	ld := k8sEventToLogData(zap.NewNop(), k8sEvent)
	rl := ld.ResourceLogs().At(0)
	logEntry := rl.ScopeLogs().At(0).LogRecords().At(0)

	assert.Equal(t, logEntry.SeverityNumber(), plog.SeverityNumberUnspecified)
	assert.Equal(t, logEntry.SeverityText(), "")
}
