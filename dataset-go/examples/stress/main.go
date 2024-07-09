/*
 * Copyright 2023 SentinelOne, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/scalyr/dataset-go/pkg/version"

	"github.com/scalyr/dataset-go/pkg/buffer_config"

	"github.com/scalyr/dataset-go/pkg/api/add_events"
	"github.com/scalyr/dataset-go/pkg/client"
	"github.com/scalyr/dataset-go/pkg/config"
	"go.uber.org/zap"
)

const (
	MaxLifeTimeMultiplier    = 50
	PurgeOlderThanMultiplier = 150
)

func main() {
	eventsCount := flag.Int("events", 1e5, "number of events")
	bucketsCount := flag.Int("buckets", 1e5, "number of buckets")
	parallel := flag.Int("parallel", 20, "number of parallel outgoing connections")
	sleep := flag.Duration("sleep", 10*time.Millisecond, "sleep between sending two events")
	logFile := flag.String("log", fmt.Sprintf("log-%s-%d.log", version.Version, time.Now().UnixMilli()), "log file for stats")
	logEvery := flag.Duration("log-every", time.Second, "how often log statistics")
	enablePProf := flag.Bool("pprof", false, "enable pprof")

	flag.Parse()

	logger := zap.Must(zap.NewDevelopment())

	// log input parameters
	logger.Info("Running stress test - input:",
		zap.Int("events", *eventsCount),
		zap.Int("buckets", *bucketsCount),
		zap.Int("parallel", *parallel),
		zap.Duration("sleep", *sleep),
		zap.String("log", *logFile),
		zap.Duration("log-every", *logEvery),
		zap.Bool("pprof", *enablePProf),
		zap.String("version", version.Version),
	)

	if *bucketsCount == -1 {
		*bucketsCount = PurgeOlderThanMultiplier
	}

	logger.Info("Running stress test - adjusted:",
		zap.Int("events", *eventsCount),
		zap.Int("buckets", *bucketsCount),
		zap.Int("parallel", *parallel),
		zap.Duration("sleep", *sleep),
		zap.String("log", *logFile),
		zap.Duration("log-every", *logEvery),
		zap.Bool("pprof", *enablePProf),
		zap.String("version", version.Version),
	)

	if *enablePProf {
		runtime.SetBlockProfileRate(1)
		runtime.SetMutexProfileFraction(1)
		go func() {
			http.ListenAndServe("localhost:8080", nil)
		}()
	}

	apiCalls := atomic.Uint64{}

	// start dummy server that accepts everything
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		apiCalls.Add(1)
		payload, err := json.Marshal(map[string]interface{}{
			"status":       "success",
			"bytesCharged": 42,
		})
		check(err)
		_, err = w.Write(payload)
		check(err)
	}))
	defer server.Close()

	cfg := config.NewDefaultDataSetConfig()
	bufferCfg, err := cfg.BufferSettings.WithOptions(
		buffer_config.WithGroupBy([]string{"body.str"}),
		buffer_config.WithMaxLifetime(MaxLifeTimeMultiplier**sleep),
		buffer_config.WithPurgeOlderThan(PurgeOlderThanMultiplier**sleep),
		buffer_config.WithMaxParallelOutgoing(*parallel),
	)
	check(err)
	cfgUpdated, err := cfg.WithOptions(
		config.WithBufferSettings(*bufferCfg),
		config.WithEndpoint(server.URL),
	)
	check(err)
	cfgUpdated.Tokens.WriteLog = "foo"

	dataSetClient, err := client.NewClient(
		cfgUpdated,
		&http.Client{},
		logger,
		nil,
		nil,
	)
	check(err)

	go logStats(dataSetClient, &apiCalls, *logFile, *logEvery)

	// start sending events
	logger.Info(
		"STRESS - Start adding events",
	)
	for i := 0; i < *eventsCount; i++ {
		batch := make([]*add_events.EventBundle, 0)
		key := fmt.Sprintf("%d", i%*bucketsCount)

		logger.Debug(
			"STRESS - Creating event",
			zap.Int("i", i),
			zap.String("key", key),
		)

		attrs := make(map[string]interface{})
		attrs["body.str"] = key
		attrs["attributes.p1"] = strings.Repeat("A", rand.Intn(2000))

		event := &add_events.Event{
			Thread: "5",
			Sev:    3,
			Ts:     fmt.Sprintf("%d", time.Now().Nanosecond()),
			Attrs:  attrs,
		}

		thread := &add_events.Thread{
			Id:   "5",
			Name: "fred",
		}
		log := &add_events.Log{
			Id: "LO",
			Attrs: map[string]interface{}{
				"key": strings.Repeat("A", rand.Intn(200)),
			},
		}
		eventBundle := &add_events.EventBundle{Event: event, Thread: thread, Log: log}

		batch = append(batch, eventBundle)
		go func(batch []*add_events.EventBundle) {
			err := dataSetClient.AddEvents(batch)
			check(err)
		}(batch)

		if i%*bucketsCount == 0 {
			time.Sleep(PurgeOlderThanMultiplier * *sleep)
		}
		time.Sleep(*sleep)
	}

	// wait until everything is processed
	logger.Info(
		"STRESS - Wait for everything to finish",
	)
	for {
		processed := uint64(0)
		stats := dataSetClient.Statistics()
		if stats != nil {
			processed = stats.Events.Processed()
		}
		logger.Info("Processed events",
			zap.Uint64("processed", processed),
			zap.Int("expecting", *eventsCount),
		)
		if processed >= uint64(*eventsCount) {
			break
		}
		time.Sleep(time.Second)
	}

	// wait for extra 1 minute to see how the memory will behave
	logger.Info(
		"STRESS - Extra sleep at the end",
	)
	extraSleepFor := 60
	for i := 0; i <= extraSleepFor; i++ {
		time.Sleep(time.Second)
		logger.Info("Extra sleep",
			zap.Int("now", i),
			zap.Int("limit", extraSleepFor),
		)
	}

	shutdownError := dataSetClient.Shutdown()
	check(shutdownError)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func logStats(client *client.DataSetClient, apiCalls *atomic.Uint64, logFile string, logEvery time.Duration) {
	f, err := os.Create(logFile)
	check(err)

	_, err = f.WriteString("i\tTime\tEvEnqueued\tEvProcessed\tEvBroken\tEvDropped\tBufEnqueued\tBufProcessed\tBufBroken\tBufDropped\tSesOpened\tSesClosed\tCalls\tHeapAlloc\tHeapSys\tMallocs\tFrees\tHeapObjects\tVersion\n")
	check(err)

	for i := 0; ; i++ {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		evEnqueued := uint64(0)
		evProcessed := uint64(0)
		evDropped := uint64(0)
		evBroken := uint64(0)
		bufEnqueued := uint64(0)
		bufProcessed := uint64(0)
		bufDropped := uint64(0)
		bufBroken := uint64(0)
		sesOpened := uint64(0)
		sesClosed := uint64(0)
		clientStats := client.Statistics()
		if clientStats != nil {
			evEnqueued = clientStats.Events.Enqueued()
			evProcessed = clientStats.Events.Processed()
			evDropped = clientStats.Events.Dropped()
			evBroken = clientStats.Events.Broken()

			bufEnqueued = clientStats.Buffers.Enqueued()
			bufProcessed = clientStats.Buffers.Processed()
			bufDropped = clientStats.Buffers.Dropped()
			bufBroken = clientStats.Buffers.Broken()

			sesOpened = clientStats.Sessions.SessionsOpened()
			sesClosed = clientStats.Sessions.SessionsClosed()
		}

		_, err := f.WriteString(
			fmt.Sprintf(
				"%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%s\n",
				i,
				time.Now().Unix(),
				evEnqueued,
				evProcessed,
				evDropped,
				evBroken,
				bufEnqueued,
				bufProcessed,
				bufDropped,
				bufBroken,
				sesOpened,
				sesClosed,
				apiCalls.Load(),
				memStats.HeapAlloc,
				memStats.HeapSys,
				memStats.Mallocs,
				memStats.Frees,
				memStats.HeapObjects,
				version.Version,
			),
		)
		check(err)

		time.Sleep(logEvery)
	}
}
