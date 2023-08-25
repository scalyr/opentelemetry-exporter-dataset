// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package opensearchexporter

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configauth"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/exporter/exporterhelper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/opensearchexporter/internal/metadata"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
	require.NoError(t, err)

	sampleEndpoint := "https://opensearch.example.com:9200"
	sampleCfg := withDefaultConfig(func(config *Config) {
		config.Endpoint = sampleEndpoint
	})
	maxIdleConns := 100
	idleConnTimeout := 90 * time.Second

	tests := []struct {
		id                   component.ID
		expected             component.Config
		configValidateAssert assert.ErrorAssertionFunc
	}{
		{
			id:                   component.NewIDWithName(metadata.Type, ""),
			expected:             sampleCfg,
			configValidateAssert: assert.NoError,
		},
		{
			id: component.NewIDWithName(metadata.Type, "trace"),
			expected: &Config{
				Dataset:   "ngnix",
				Namespace: "eu",
				HTTPClientSettings: confighttp.HTTPClientSettings{
					Endpoint: sampleEndpoint,
					Timeout:  2 * time.Minute,
					Headers: map[string]configopaque.String{
						"myheader": "test",
					},
					MaxIdleConns:    &maxIdleConns,
					IdleConnTimeout: &idleConnTimeout,
					Auth:            &configauth.Authentication{AuthenticatorID: component.NewID("sample_basic_auth")},
				},
				RetrySettings: exporterhelper.RetrySettings{
					Enabled:             true,
					InitialInterval:     100 * time.Millisecond,
					MaxInterval:         30 * time.Second,
					MaxElapsedTime:      5 * time.Minute,
					Multiplier:          1.5,
					RandomizationFactor: 0.5,
				},
			},
			configValidateAssert: assert.NoError,
		},
		{
			id: component.NewIDWithName(metadata.Type, "empty_dataset"),
			expected: withDefaultConfig(func(config *Config) {
				config.Endpoint = sampleEndpoint
				config.Dataset = ""
				config.Namespace = "eu"
			}),
			configValidateAssert: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, errDatasetNoValue.Error())
			},
		},
		{
			id: component.NewIDWithName(metadata.Type, "empty_namespace"),
			expected: withDefaultConfig(func(config *Config) {
				config.Endpoint = sampleEndpoint
				config.Dataset = "ngnix"
				config.Namespace = ""
			}),
			configValidateAssert: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, errNamespaceNoValue.Error())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.id.String(), func(t *testing.T) {
			factory := NewFactory()
			cfg := factory.CreateDefaultConfig()

			sub, err := cm.Sub(tt.id.String())
			require.NoError(t, err)
			require.NoError(t, component.UnmarshalConfig(sub, cfg))

			vv := component.ValidateConfig(cfg)
			tt.configValidateAssert(t, vv)
			assert.Equal(t, tt.expected, cfg)
		})
	}
}

// withDefaultConfig create a new default configuration
// and applies provided functions to it.
func withDefaultConfig(fns ...func(*Config)) *Config {
	cfg := newDefaultConfig().(*Config)
	for _, fn := range fns {
		fn(cfg)
	}
	return cfg
}
