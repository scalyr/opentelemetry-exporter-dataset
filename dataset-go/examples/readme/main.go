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

	"github.com/scalyr/dataset-go/pkg/meter_config"

	"go.opentelemetry.io/otel"

	"github.com/scalyr/dataset-go/pkg/server_host_config"

	"github.com/scalyr/dataset-go/pkg/buffer_config"

	"github.com/scalyr/dataset-go/pkg/api/add_events"
	"github.com/scalyr/dataset-go/pkg/client"
	"github.com/scalyr/dataset-go/pkg/config"
	"go.uber.org/zap"
)

func makeBundles() []*add_events.EventBundle {
	event := &add_events.Event{
		Thread: "T",
		Log:    "L",
		Sev:    3,
		Ts:     fmt.Sprintf("%d", time.Now().Nanosecond()),
		Attrs: map[string]interface{}{
			"message": "dataset-go library - test message",
		},
	}
	thread := &add_events.Thread{Id: "T", Name: "thread-1"}
	log := &add_events.Log{
		Id: "L",
		Attrs: map[string]interface{}{
			"attr 1": "value 1",
		},
	}

	eventBundle := &add_events.EventBundle{
		Event:  event,
		Thread: thread,
		Log:    log,
	}

	return []*add_events.EventBundle{eventBundle}
}

func main() {
	logger := zap.Must(zap.NewDevelopment())

	// read configuration from env variables
	cfg, err := config.New(config.FromEnv())
	if err != nil {
		panic(err)
	}
	bufferCfg, err := cfg.BufferSettings.WithOptions(
		buffer_config.WithMaxLifetime(time.Second),
		buffer_config.WithRetryInitialInterval(time.Second),
		buffer_config.WithRetryMaxInterval(2*time.Second),
		buffer_config.WithRetryMaxElapsedTime(10*time.Second),
	)
	cfg, err = cfg.WithOptions(
		config.WithBufferSettings(*bufferCfg),
		config.WithServerHostSettings(server_host_config.NewDefaultDataSetServerHostSettings()),
	)
	if err != nil {
		panic(err)
	}
	libraryConsumerUserAgentSuffix := "OtelCollector-readme;1.2.3"
	meter := otel.Meter("example.readme")
	// build client
	cl, err := client.NewClient(cfg, &http.Client{}, logger, &libraryConsumerUserAgentSuffix, meter_config.NewMeterConfig(&meter, "all", "example"))
	if err != nil {
		panic(err)
	}

	// send bundles
	err = cl.AddEvents(makeBundles())
	if err != nil {
		panic(err)
	}

	// Finish event processing
	err = cl.Shutdown()
	if err != nil {
		panic(err)
	}
}
