// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package logicmonitorexporter

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/exporter/exporterhelper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/logicmonitorexporter/internal/metadata"
)

func TestConfigValidation(t *testing.T) {
	testcases := []struct {
		name         string
		cfg          *Config
		wantErr      bool
		errorMessage string
	}{
		{
			name: "empty endpoint",
			cfg: &Config{
				HTTPClientSettings: confighttp.HTTPClientSettings{
					Endpoint: "",
				},
			},
			wantErr:      true,
			errorMessage: "Endpoint should not be empty",
		},
		{
			name: "missing http scheme",
			cfg: &Config{
				HTTPClientSettings: confighttp.HTTPClientSettings{
					Endpoint: "test.com/dummy",
				},
			},
			wantErr:      true,
			errorMessage: "Endpoint must be valid",
		},
		{
			name: "invalid endpoint format",
			cfg: &Config{
				HTTPClientSettings: confighttp.HTTPClientSettings{
					Endpoint: "invalid.com@#$%",
				},
			},
			wantErr:      true,
			errorMessage: "Endpoint must be valid",
		},
		{
			name: "valid config",
			cfg: &Config{
				HTTPClientSettings: confighttp.HTTPClientSettings{
					Endpoint: "http://validurl.com/rest",
				},
			},
			wantErr:      false,
			errorMessage: "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("config validation failed: error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if tc.wantErr {
				assert.Error(t, err)
				if len(tc.errorMessage) != 0 {
					assert.Equal(t, errors.New(tc.errorMessage), err, "Error messages must match")
				}
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
	require.NoError(t, err)

	tests := []struct {
		id       component.ID
		expected component.Config
	}{
		{
			id: component.NewIDWithName(metadata.Type, "apitoken"),
			expected: &Config{
				RetrySettings: exporterhelper.NewDefaultRetrySettings(),
				QueueSettings: exporterhelper.NewDefaultQueueSettings(),
				HTTPClientSettings: confighttp.HTTPClientSettings{
					Endpoint: "https://company.logicmonitor.com/rest",
				},
				APIToken: APIToken{
					AccessID:  "accessid",
					AccessKey: "accesskey",
				},
			},
		},
		{
			id: component.NewIDWithName(metadata.Type, "bearertoken"),
			expected: &Config{
				RetrySettings: exporterhelper.NewDefaultRetrySettings(),
				QueueSettings: exporterhelper.NewDefaultQueueSettings(),
				HTTPClientSettings: confighttp.HTTPClientSettings{
					Endpoint: "https://company.logicmonitor.com/rest",
					Headers: map[string]configopaque.String{
						"Authorization": "Bearer <token>",
					},
				},
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

			assert.NoError(t, component.ValidateConfig(cfg))
			assert.Equal(t, tt.expected, cfg)
		})
	}
}
