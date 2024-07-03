// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package elasticsearchexporter

import (
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/exporter/exporterhelper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/elasticsearchexporter/internal/metadata"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	defaultCfg := createDefaultConfig()
	defaultCfg.(*Config).Endpoints = []string{"https://elastic.example.com:9200"}

	defaultLogstashFormatCfg := createDefaultConfig()
	defaultLogstashFormatCfg.(*Config).Endpoints = []string{"http://localhost:9200"}
	defaultLogstashFormatCfg.(*Config).LogstashFormat.Enabled = true

	defaultRawCfg := createDefaultConfig()
	defaultRawCfg.(*Config).Endpoints = []string{"http://localhost:9200"}
	defaultRawCfg.(*Config).Mapping.Mode = "raw"

	tests := []struct {
		configFile string
		id         component.ID
		expected   component.Config
	}{
		{
			id:         component.NewIDWithName(metadata.Type, ""),
			configFile: "config.yaml",
			expected:   defaultCfg,
		},
		{
			id:         component.NewIDWithName(metadata.Type, "trace"),
			configFile: "config.yaml",
			expected: &Config{
				QueueSettings: exporterhelper.QueueSettings{
					Enabled:      false,
					NumConsumers: exporterhelper.NewDefaultQueueSettings().NumConsumers,
					QueueSize:    exporterhelper.NewDefaultQueueSettings().QueueSize,
				},
				Endpoints:   []string{"https://elastic.example.com:9200"},
				Index:       "",
				LogsIndex:   "logs-generic-default",
				TracesIndex: "trace_index",
				Pipeline:    "mypipeline",
				ClientConfig: ClientConfig{
					Authentication: AuthenticationSettings{
						User:     "elastic",
						Password: "search",
						APIKey:   "AvFsEiPs==",
					},
					Timeout: 2 * time.Minute,
					Headers: map[string]string{
						"myheader": "test",
					},
				},
				Discovery: DiscoverySettings{
					OnStart: true,
				},
				Flush: FlushSettings{
					Bytes: 10485760,
				},
				Retry: RetrySettings{
					Enabled:         true,
					MaxRequests:     5,
					InitialInterval: 100 * time.Millisecond,
					MaxInterval:     1 * time.Minute,
					RetryOnStatus:   []int{http.StatusTooManyRequests, http.StatusInternalServerError},
				},
				Mapping: MappingsSettings{
					Mode:  "none",
					Dedup: true,
					Dedot: true,
				},
				LogstashFormat: LogstashFormatSettings{
					Enabled:         false,
					PrefixSeparator: "-",
					DateFormat:      "%Y.%m.%d",
				},
			},
		},
		{
			id:         component.NewIDWithName(metadata.Type, "log"),
			configFile: "config.yaml",
			expected: &Config{
				QueueSettings: exporterhelper.QueueSettings{
					Enabled:      true,
					NumConsumers: exporterhelper.NewDefaultQueueSettings().NumConsumers,
					QueueSize:    exporterhelper.NewDefaultQueueSettings().QueueSize,
				},
				Endpoints:   []string{"http://localhost:9200"},
				Index:       "",
				LogsIndex:   "my_log_index",
				TracesIndex: "traces-generic-default",
				Pipeline:    "mypipeline",
				ClientConfig: ClientConfig{
					Authentication: AuthenticationSettings{
						User:     "elastic",
						Password: "search",
						APIKey:   "AvFsEiPs==",
					},
					Timeout: 2 * time.Minute,
					Headers: map[string]string{
						"myheader": "test",
					},
				},
				Discovery: DiscoverySettings{
					OnStart: true,
				},
				Flush: FlushSettings{
					Bytes: 10485760,
				},
				Retry: RetrySettings{
					Enabled:         true,
					MaxRequests:     5,
					InitialInterval: 100 * time.Millisecond,
					MaxInterval:     1 * time.Minute,
					RetryOnStatus:   []int{http.StatusTooManyRequests, http.StatusInternalServerError},
				},
				Mapping: MappingsSettings{
					Mode:  "none",
					Dedup: true,
					Dedot: true,
				},
				LogstashFormat: LogstashFormatSettings{
					Enabled:         false,
					PrefixSeparator: "-",
					DateFormat:      "%Y.%m.%d",
				},
			},
		},
		{
			id:         component.NewIDWithName(metadata.Type, "logstash_format"),
			configFile: "config.yaml",
			expected:   defaultLogstashFormatCfg,
		},
		{
			id:         component.NewIDWithName(metadata.Type, "raw"),
			configFile: "config.yaml",
			expected:   defaultRawCfg,
		},
		{
			id:         component.NewIDWithName(metadata.Type, "cloudid"),
			configFile: "config.yaml",
			expected: withDefaultConfig(func(cfg *Config) {
				cfg.CloudID = "foo:YmFyLmNsb3VkLmVzLmlvJGFiYzEyMyRkZWY0NTY="
			}),
		},
		{
			id:         component.NewIDWithName(metadata.Type, "deprecated_index"),
			configFile: "config.yaml",
			expected: withDefaultConfig(func(cfg *Config) {
				cfg.Endpoints = []string{"https://elastic.example.com:9200"}
				cfg.Index = "my_log_index"
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.id.String(), func(t *testing.T) {
			factory := NewFactory()
			cfg := factory.CreateDefaultConfig()

			cm, err := confmaptest.LoadConf(filepath.Join("testdata", tt.configFile))
			require.NoError(t, err)

			sub, err := cm.Sub(tt.id.String())
			require.NoError(t, err)
			require.NoError(t, sub.Unmarshal(cfg))

			assert.NoError(t, component.ValidateConfig(cfg))
			assert.Equal(t, tt.expected, cfg)
		})
	}
}

// TestConfig_Validate tests the error cases of Config.Validate.
//
// Successful validation should be covered by TestConfig above.
func TestConfig_Validate(t *testing.T) {
	tests := map[string]struct {
		config *Config
		err    string
	}{
		"no endpoints": {
			config: withDefaultConfig(),
			err:    "endpoints or cloudid must be specified",
		},
		"empty endpoint": {
			config: withDefaultConfig(func(cfg *Config) {
				cfg.Endpoints = []string{""}
			}),
			err: "endpoints must not include empty entries",
		},
		"invalid endpoint": {
			config: withDefaultConfig(func(cfg *Config) {
				cfg.Endpoints = []string{"*:!"}
			}),
			err: `invalid endpoint "*:!": parse "*:!": first path segment in URL cannot contain colon`,
		},
		"invalid cloudid": {
			config: withDefaultConfig(func(cfg *Config) {
				cfg.CloudID = "invalid"
			}),
			err: `invalid CloudID "invalid"`,
		},
		"invalid decoded cloudid": {
			config: withDefaultConfig(func(cfg *Config) {
				cfg.CloudID = "foo:YWJj"
			}),
			err: `invalid decoded CloudID "abc"`,
		},
		"endpoint and cloudid both set": {
			config: withDefaultConfig(func(cfg *Config) {
				cfg.Endpoints = []string{"http://test:9200"}
				cfg.CloudID = "foo:YmFyLmNsb3VkLmVzLmlvJGFiYzEyMyRkZWY0NTY="
			}),
			err: "only one of endpoints or cloudid may be specified",
		},
		"invalid mapping mode": {
			config: withDefaultConfig(func(cfg *Config) {
				cfg.Endpoints = []string{"http://test:9200"}
				cfg.Mapping.Mode = "invalid"
			}),
			err: `unknown mapping mode "invalid"`,
		},
		"invalid scheme": {
			config: withDefaultConfig(func(cfg *Config) {
				cfg.Endpoints = []string{"without_scheme"}
			}),
			err: `invalid endpoint "without_scheme": invalid scheme "", expected "http" or "https"`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := tt.config.Validate()
			assert.EqualError(t, err, tt.err)
		})
	}
}

func TestConfig_Validate_Environment(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		t.Setenv("ELASTICSEARCH_URL", "http://test:9200")
		config := withDefaultConfig()
		err := config.Validate()
		require.NoError(t, err)
	})
	t.Run("invalid", func(t *testing.T) {
		t.Setenv("ELASTICSEARCH_URL", "http://valid:9200, *:!")
		config := withDefaultConfig()
		err := config.Validate()
		assert.EqualError(t, err, `invalid endpoint "*:!": parse "*:!": first path segment in URL cannot contain colon`)
	})
}

func withDefaultConfig(fns ...func(*Config)) *Config {
	cfg := createDefaultConfig().(*Config)
	for _, fn := range fns {
		fn(cfg)
	}
	return cfg
}
