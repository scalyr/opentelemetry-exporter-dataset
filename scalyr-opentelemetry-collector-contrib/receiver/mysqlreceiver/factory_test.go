// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package mysqlreceiver

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confignet"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

func TestType(t *testing.T) {
	factory := NewFactory()
	ft := factory.Type()
	require.EqualValues(t, "mysql", ft)
}

func TestValidConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.Username = "otel"
	cfg.Password = "otel"
	cfg.Endpoint = "localhost:3306"
	require.NoError(t, component.ValidateConfig(cfg))
}

func TestCreateMetricsReceiver(t *testing.T) {
	factory := NewFactory()
	metricsReceiver, err := factory.CreateMetricsReceiver(
		context.Background(),
		receivertest.NewNopCreateSettings(),
		&Config{
			ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
				CollectionInterval: 10 * time.Second,
				InitialDelay:       time.Second,
			},
			Username: "otel",
			Password: "otel",
			NetAddr: confignet.NetAddr{
				Endpoint: "localhost:3306",
			},
		},
		consumertest.NewNop(),
	)
	require.NoError(t, err)
	require.NotNil(t, metricsReceiver)
}
