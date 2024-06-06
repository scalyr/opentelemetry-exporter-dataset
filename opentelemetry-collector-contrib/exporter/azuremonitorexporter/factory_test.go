// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package azuremonitorexporter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/exporter/exportertest"
)

// An inappropriate config
type badConfig struct {
}

func TestCreateTracesExporterUsingSpecificTransportChannel(t *testing.T) {
	// mock transport channel creation
	f := factory{tChannel: &mockTransportChannel{}}
	ctx := context.Background()
	params := exportertest.NewNopCreateSettings()
	config := createDefaultConfig().(*Config)
	config.ConnectionString = "InstrumentationKey=test-key;IngestionEndpoint=https://test-endpoint/"
	exporter, err := f.createTracesExporter(ctx, params, config)
	assert.NotNil(t, exporter)
	assert.NoError(t, err)
}

func TestCreateTracesExporterUsingDefaultTransportChannel(t *testing.T) {
	// We get the default transport channel creation, if we don't specify one during f creation
	f := factory{}
	assert.Nil(t, f.tChannel)
	ctx := context.Background()
	config := createDefaultConfig().(*Config)
	config.ConnectionString = "InstrumentationKey=test-key;IngestionEndpoint=https://test-endpoint/"
	exporter, err := f.createTracesExporter(ctx, exportertest.NewNopCreateSettings(), config)
	assert.NotNil(t, exporter)
	assert.NoError(t, err)
	assert.NotNil(t, f.tChannel)
}

func TestCreateTracesExporterUsingBadConfig(t *testing.T) {
	// We get the default transport channel creation, if we don't specify one during factory creation
	f := factory{}
	assert.Nil(t, f.tChannel)
	ctx := context.Background()
	params := exportertest.NewNopCreateSettings()

	badConfig := &badConfig{}

	exporter, err := f.createTracesExporter(ctx, params, badConfig)
	assert.Nil(t, exporter)
	assert.Error(t, err)
}
