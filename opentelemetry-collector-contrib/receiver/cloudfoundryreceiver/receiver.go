// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package cloudfoundryreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/cloudfoundryreceiver"

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"code.cloudfoundry.org/go-loggregator"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/receiverhelper"
)

const (
	transport              = "http"
	dataFormat             = "cloudfoundry"
	instrumentationLibName = "otelcol/cloudfoundry"
)

var _ receiver.Metrics = (*cloudFoundryReceiver)(nil)

// newCloudFoundryReceiver implements the receiver.Metrics for Cloud Foundry protocol.
type cloudFoundryReceiver struct {
	settings          component.TelemetrySettings
	cancel            context.CancelFunc
	config            Config
	nextConsumer      consumer.Metrics
	obsrecv           *receiverhelper.ObsReport
	goroutines        sync.WaitGroup
	receiverStartTime time.Time
}

// newCloudFoundryReceiver creates the Cloud Foundry receiver with the given parameters.
func newCloudFoundryReceiver(
	settings receiver.CreateSettings,
	config Config,
	nextConsumer consumer.Metrics) (receiver.Metrics, error) {

	obsrecv, err := receiverhelper.NewObsReport(receiverhelper.ObsReportSettings{
		ReceiverID:             settings.ID,
		Transport:              transport,
		ReceiverCreateSettings: settings,
	})
	if err != nil {
		return nil, err
	}

	return &cloudFoundryReceiver{
		settings:          settings.TelemetrySettings,
		config:            config,
		nextConsumer:      nextConsumer,
		obsrecv:           obsrecv,
		receiverStartTime: time.Now(),
	}, nil
}

func (cfr *cloudFoundryReceiver) Start(ctx context.Context, host component.Host) error {
	tokenProvider, tokenErr := newUAATokenProvider(cfr.settings.Logger, cfr.config.UAA.LimitedClientConfig, cfr.config.UAA.Username, string(cfr.config.UAA.Password))
	if tokenErr != nil {
		return fmt.Errorf("create cloud foundry UAA token provider: %w", tokenErr)
	}

	streamFactory, streamErr := newEnvelopeStreamFactory(
		ctx,
		cfr.settings,
		tokenProvider,
		cfr.config.RLPGateway.ClientConfig,
		host,
	)
	if streamErr != nil {
		return fmt.Errorf("creating cloud foundry RLP envelope stream factory: %w", streamErr)
	}

	innerCtx, cancel := context.WithCancel(ctx)
	cfr.cancel = cancel

	cfr.goroutines.Add(1)

	go func() {
		defer cfr.goroutines.Done()
		cfr.settings.Logger.Debug("cloud foundry receiver starting")

		_, tokenErr = tokenProvider.ProvideToken()
		if tokenErr != nil {
			cfr.settings.ReportStatus(component.NewFatalErrorEvent(fmt.Errorf("cloud foundry receiver failed to fetch initial token from UAA: %w", tokenErr)))
			return
		}

		envelopeStream, err := streamFactory.CreateStream(innerCtx, cfr.config.RLPGateway.ShardID)
		if err != nil {
			cfr.settings.ReportStatus(component.NewFatalErrorEvent(fmt.Errorf("creating RLP gateway envelope stream: %w", err)))
			return
		}

		cfr.streamMetrics(innerCtx, envelopeStream)
		cfr.settings.Logger.Debug("cloudfoundry metrics streamer stopped")
	}()

	return nil
}

func (cfr *cloudFoundryReceiver) Shutdown(_ context.Context) error {
	if cfr.cancel == nil {
		return nil
	}
	cfr.cancel()
	cfr.goroutines.Wait()
	return nil
}

func (cfr *cloudFoundryReceiver) streamMetrics(
	ctx context.Context,
	stream loggregator.EnvelopeStream) {

	for {
		// Blocks until non-empty result or context is cancelled (returns nil in that case)
		envelopes := stream()
		if envelopes == nil {
			// If context has not been cancelled, then nil means the shutdown was due to an error within stream
			if ctx.Err() == nil {
				cfr.settings.ReportStatus(component.NewFatalErrorEvent(errors.New("RLP gateway streamer shut down due to an error")))
			}

			break
		}

		metrics := pmetric.NewMetrics()
		libraryMetrics := createLibraryMetricsSlice(metrics)

		for _, envelope := range envelopes {
			if envelope != nil {
				// There is no concept of startTime in CF loggregator, and we do not know the uptime of the component
				// from which the metric originates, so just provide receiver start time as metric start time
				convertEnvelopeToMetrics(envelope, libraryMetrics, cfr.receiverStartTime)
			}
		}

		if libraryMetrics.Len() > 0 {
			obsCtx := cfr.obsrecv.StartMetricsOp(ctx)
			err := cfr.nextConsumer.ConsumeMetrics(ctx, metrics)
			cfr.obsrecv.EndMetricsOp(obsCtx, dataFormat, metrics.DataPointCount(), err)
		}
	}
}

func createLibraryMetricsSlice(metrics pmetric.Metrics) pmetric.MetricSlice {
	resourceMetrics := metrics.ResourceMetrics()
	resourceMetric := resourceMetrics.AppendEmpty()
	resourceMetric.Resource().Attributes()
	libraryMetricsSlice := resourceMetric.ScopeMetrics()
	libraryMetrics := libraryMetricsSlice.AppendEmpty()
	libraryMetrics.Scope().SetName(instrumentationLibName)
	return libraryMetrics.Metrics()
}
