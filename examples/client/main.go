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
	"fmt"
	"net/http"
	"time"

	"github.com/scalyr/dataset-go/pkg/server_host_config"

	"github.com/scalyr/dataset-go/pkg/buffer_config"

	"github.com/scalyr/dataset-go/pkg/api/add_events"
	"github.com/scalyr/dataset-go/pkg/client"
	"github.com/scalyr/dataset-go/pkg/config"
	"go.uber.org/zap"
)

const (
	BundleCount = 10000
	BatchSize   = 100
	BatchDelay  = time.Second
)

func main() {
	// read configuration from env variables
	cfg, err := config.New(config.FromEnv())
	if err != nil {
		panic(err)
	}

	// manually adjust delay between sending buffers
	bufCfg, err := cfg.BufferSettings.WithOptions(buffer_config.WithMaxLifetime(3 * BatchDelay))
	if err != nil {
		panic(err)
	}
	cfg, err = cfg.WithOptions(
		config.WithBufferSettings(*bufCfg),
		config.WithServerHostSettings(server_host_config.NewDefaultDataSetServerHostSettings()),
	)
	if err != nil {
		panic(err)
	}

	libraryConsumerUserAgentSuffix := "OtelCollector;1.2.3"

	// build client
	cl, err := client.NewClient(
		cfg,
		&http.Client{},
		zap.Must(zap.NewDevelopment()),
		&libraryConsumerUserAgentSuffix,
	)
	if err != nil {
		panic(err)
	}

	// build some bundles
	bundles := make([]*add_events.EventBundle, 0)
	for i := 0; i < BundleCount; i++ {
		event := &add_events.Event{
			Thread: "T",
			Log:    "L",
			Sev:    3,
			Ts:     fmt.Sprintf("%d", time.Now().Nanosecond()),
			Attrs: map[string]interface{}{
				"message": fmt.Sprintf(
					"dataset-go library - test message - %d",
					i,
				),
				"attribute": i,
			},
		}
		thread := &add_events.Thread{Id: "T", Name: "thread-1"}
		log := &add_events.Log{
			Id: "L",
			Attrs: map[string]interface{}{
				"attr 1": fmt.Sprintf("value %d", i),
			},
		}

		eventBundle := &add_events.EventBundle{
			Event:  event,
			Thread: thread,
			Log:    log,
		}

		bundles = append(bundles, eventBundle)
	}

	// send bundles in batches
	for j := 0; j <= BundleCount; j += BatchSize {
		from := j
		to := j + BatchSize
		if to >= BundleCount {
			to = BundleCount
		}
		fmt.Printf("Sending batch: %5d - %5d\n", from, to)
		err := cl.AddEvents(bundles[from:to])
		if err != nil {
			fmt.Printf("There was problem in batch starting at %d: %v", j, err)
		}
		time.Sleep(BatchDelay)
	}

	shutdownError := cl.Shutdown()
	if shutdownError != nil {
		panic(shutdownError)
	}
}
