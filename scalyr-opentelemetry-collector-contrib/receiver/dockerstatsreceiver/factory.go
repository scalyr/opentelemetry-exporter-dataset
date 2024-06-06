// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package dockerstatsreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver"

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/featuregate"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/scraperhelper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver/internal/metadata"
)

var _ = featuregate.GlobalRegistry().MustRegister(
	"receiver.dockerstats.useScraperV2",
	featuregate.StageStable,
	featuregate.WithRegisterDescription("When enabled, the receiver will use the function ScrapeV2 to collect metrics. This allows each metric to be turned off/on via config. The new metrics are slightly different to the legacy implementation."),
	featuregate.WithRegisterReferenceURL("https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/9794"),
	featuregate.WithRegisterToVersion("0.74.0"),
)

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		metadata.Type,
		createDefaultConfig,
		receiver.WithMetrics(createMetricsReceiver, metadata.MetricsStability))
}

func createDefaultConfig() component.Config {
	scs := scraperhelper.NewDefaultScraperControllerSettings(metadata.Type)
	scs.CollectionInterval = 10 * time.Second
	scs.Timeout = 5 * time.Second
	return &Config{
		ScraperControllerSettings: scs,
		Endpoint:                  "unix:///var/run/docker.sock",
		DockerAPIVersion:          defaultDockerAPIVersion,
		MetricsBuilderConfig:      metadata.DefaultMetricsBuilderConfig(),
	}
}

func createMetricsReceiver(
	_ context.Context,
	params receiver.CreateSettings,
	config component.Config,
	consumer consumer.Metrics,
) (receiver.Metrics, error) {
	dockerConfig := config.(*Config)
	dsr := newMetricsReceiver(params, dockerConfig)

	scrp, err := scraperhelper.NewScraper(metadata.Type.String(), dsr.scrapeV2, scraperhelper.WithStart(dsr.start))
	if err != nil {
		return nil, err
	}

	return scraperhelper.NewScraperControllerReceiver(&dsr.config.ScraperControllerSettings, params, consumer, scraperhelper.AddScraper(scrp))
}
