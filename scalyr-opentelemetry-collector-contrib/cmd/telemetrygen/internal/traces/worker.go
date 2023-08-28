// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package traces // import "github.com/open-telemetry/opentelemetry-collector-contrib/cmd/telemetrygen/internal/traces"

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type worker struct {
	running          *atomic.Bool    // pointer to shared flag that indicates it's time to stop the test
	numTraces        int             // how many traces the worker has to generate (only when duration==0)
	propagateContext bool            // whether the worker needs to propagate the trace context via HTTP headers
	statusCode       codes.Code      // the status code set for the child and parent spans
	totalDuration    time.Duration   // how long to run the test for (overrides `numTraces`)
	limitPerSecond   rate.Limit      // how many spans per second to generate
	wg               *sync.WaitGroup // notify when done
	logger           *zap.Logger
	loadSize         int
}

const (
	fakeIP string = "1.2.3.4"

	fakeSpanDuration = 123 * time.Microsecond

	charactersPerMB = 1024 * 1024 // One character takes up one byte of space, so this number comes from the number of bytes in a megabyte
)

func (w worker) simulateTraces() {
	tracer := otel.Tracer("telemetrygen")
	limiter := rate.NewLimiter(w.limitPerSecond, 1)
	var i int
	for w.running.Load() {
		ctx, sp := tracer.Start(context.Background(), "lets-go", trace.WithAttributes(
			semconv.NetPeerIPKey.String(fakeIP),
			semconv.PeerServiceKey.String("telemetrygen-server"),
		),
			trace.WithSpanKind(trace.SpanKindClient),
		)
		for j := 0; j < w.loadSize; j++ {
			sp.SetAttributes(attribute.String(fmt.Sprintf("load-%v", j), string(make([]byte, charactersPerMB))))
		}

		childCtx := ctx
		if w.propagateContext {
			header := propagation.HeaderCarrier{}
			// simulates going remote
			otel.GetTextMapPropagator().Inject(childCtx, header)

			// simulates getting a request from a client
			childCtx = otel.GetTextMapPropagator().Extract(childCtx, header)
		}

		_, child := tracer.Start(childCtx, "okey-dokey", trace.WithAttributes(
			semconv.NetPeerIPKey.String(fakeIP),
			semconv.PeerServiceKey.String("telemetrygen-client"),
		),
			trace.WithSpanKind(trace.SpanKindServer),
		)

		if err := limiter.Wait(context.Background()); err != nil {
			w.logger.Fatal("limiter waited failed, retry", zap.Error(err))
		}

		opt := trace.WithTimestamp(time.Now().Add(fakeSpanDuration))
		child.SetStatus(w.statusCode, "")
		child.End(opt)
		sp.SetStatus(w.statusCode, "")
		sp.End(opt)

		i++
		if w.numTraces != 0 {
			if i >= w.numTraces {
				break
			}
		}
	}
	w.logger.Info("traces generated", zap.Int("traces", i))
	w.wg.Done()
}
