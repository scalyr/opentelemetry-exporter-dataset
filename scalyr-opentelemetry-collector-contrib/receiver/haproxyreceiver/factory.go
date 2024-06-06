// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package haproxyreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/haproxyreceiver"

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/scraperhelper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/haproxyreceiver/internal/metadata"
)

// NewFactory creates a new HAProxy receiver factory.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		metadata.Type,
		newDefaultConfig,
		receiver.WithMetrics(newReceiver, metadata.MetricsStability))
}

func newDefaultConfig() component.Config {
	return &Config{
		ScraperControllerSettings: scraperhelper.NewDefaultScraperControllerSettings(metadata.Type),
		MetricsBuilderConfig:      metadata.DefaultMetricsBuilderConfig(),
	}
}

func newReceiver(
	_ context.Context,
	settings receiver.CreateSettings,
	cfg component.Config,
	consumer consumer.Metrics,
) (receiver.Metrics, error) {
	haProxyCfg := cfg.(*Config)
	mp := newScraper(haProxyCfg, settings)
	s, err := scraperhelper.NewScraper(metadata.Type.String(), mp.scrape, scraperhelper.WithStart(mp.start))
	if err != nil {
		return nil, err
	}

	return scraperhelper.NewScraperControllerReceiver(
		&haProxyCfg.ScraperControllerSettings,
		settings,
		consumer,
		scraperhelper.AddScraper(s),
	)
}
