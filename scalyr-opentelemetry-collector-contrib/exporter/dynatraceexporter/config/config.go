// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package config // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/dynatraceexporter/config"

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dynatrace-oss/dynatrace-metric-utils-go/metric/apiconstants"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/exporter/exporterhelper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/resourcetotelemetry"
)

// Config defines configuration for the Dynatrace exporter.
type Config struct {
	confighttp.ClientConfig `mapstructure:",squash"`

	exporterhelper.QueueSettings `mapstructure:"sending_queue"`
	configretry.BackOffConfig    `mapstructure:"retry_on_failure"`
	ResourceToTelemetrySettings  resourcetotelemetry.Settings `mapstructure:"resource_to_telemetry_conversion"`

	// Dynatrace API token with metrics ingest permission
	APIToken string `mapstructure:"api_token"`

	// DefaultDimensions will be added to all exported metrics
	DefaultDimensions map[string]string `mapstructure:"default_dimensions"`

	// String to prefix all metric names
	Prefix string `mapstructure:"prefix"`

	// Tags will be added to all exported metrics
	// Deprecated: Please use DefaultDimensions instead
	Tags []string `mapstructure:"tags"`
}

func (c *Config) Validate() error {
	if err := c.QueueSettings.Validate(); err != nil {
		return fmt.Errorf("queue settings has invalid configuration: %w", err)
	}

	if c.ClientConfig.Headers == nil {
		c.ClientConfig.Headers = make(map[string]configopaque.String)
	}
	c.APIToken = strings.TrimSpace(c.APIToken)

	if c.Endpoint == "" {
		c.Endpoint = apiconstants.GetDefaultOneAgentEndpoint()
	} else {
		if c.APIToken == "" {
			return errors.New("api_token is required if Endpoint is provided")
		}

		c.ClientConfig.Headers["Authorization"] = configopaque.String(fmt.Sprintf("Api-Token %s", c.APIToken))
	}

	if !(strings.HasPrefix(c.Endpoint, "http://") || strings.HasPrefix(c.Endpoint, "https://")) {
		return errors.New("endpoint must start with https:// or http://")
	}

	c.ClientConfig.Headers["Content-Type"] = "text/plain; charset=UTF-8"
	c.ClientConfig.Headers["User-Agent"] = "opentelemetry-collector"

	return nil
}
