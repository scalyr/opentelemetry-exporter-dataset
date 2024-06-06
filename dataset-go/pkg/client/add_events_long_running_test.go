//go:build long_running

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

package client

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/scalyr/dataset-go/pkg/buffer_config"
	"github.com/scalyr/dataset-go/pkg/config"
	"github.com/scalyr/dataset-go/pkg/server_host_config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/scalyr/dataset-go/pkg/api/add_events"
)

func TestAddEventsManyLogsShouldSucceed(t *testing.T) {
	const MaxDelay = 200 * time.Millisecond
	const PurgeOlderThan = 15 * MaxDelay

	const Cycles = 3
	const MaxBatchCount = 400
	const LogsPerBatch = 500
	const ExpectedLogs = uint64(Cycles * MaxBatchCount * LogsPerBatch)

	attempt := atomic.Uint64{}
	lastCall := atomic.Int64{}
	processedEvents := atomic.Uint64{}
	seenKeys := make(map[string]int64)
	expectedKeys := make(map[string]int64)
	seenMutex := &sync.RWMutex{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		attempt.Add(1)
		cer, err := extract(req)

		assert.Nil(t, err, "Error reading request: %v", err)

		for _, ev := range cer.Events {
			processedEvents.Add(1)
			key := ev.Attrs["body.str"]
			seenMutex.Lock()
			sKey := key.(string)
			_, f := seenKeys[sKey]
			if !f {
				seenKeys[sKey] = 0
			}
			seenKeys[sKey] += 1
			seenMutex.Unlock()
		}

		batch := (*cer.SessionInfo)["batch"]
		t.Logf("Accepting batch: %s", batch.(string))

		lastCall.Store(time.Now().UnixNano())
		time.Sleep(time.Duration(float64(MaxDelay) * 0.6))
		payload, err := json.Marshal(map[string]interface{}{
			"status":       "success",
			"bytesCharged": 42,
		})
		l, err := w.Write(payload)
		assert.Greater(t, l, 1)
		assert.NoError(t, err)
	}))
	defer server.Close()

	config := &config.DataSetConfig{
		Endpoint: server.URL,
		Tokens:   config.DataSetTokens{WriteLog: "AAAA"},
		BufferSettings: buffer_config.DataSetBufferSettings{
			MaxSize:                  1000,
			GroupBy:                  []string{"batch"},
			MaxLifetime:              5 * MaxDelay,
			PurgeOlderThan:           PurgeOlderThan,
			RetryRandomizationFactor: 1.0,
			RetryMultiplier:          1.0,
			RetryInitialInterval:     RetryBase,
			RetryMaxInterval:         RetryBase,
			RetryMaxElapsedTime:      10 * RetryBase,
			RetryShutdownTimeout:     50 * RetryBase,
		},
		ServerHostSettings: server_host_config.NewDefaultDataSetServerHostSettings(),
	}
	sc, err := NewClient(config, &http.Client{}, zap.Must(zap.NewDevelopment()), nil, nil)
	require.Nil(t, err)

	lastCall.Store(time.Now().UnixNano())
	for cI := 0; cI < Cycles; cI++ {
		for bI := 0; bI < MaxBatchCount; bI++ {
			batch := make([]*add_events.EventBundle, 0)
			batchKey := fmt.Sprintf("%d", bI)
			for lI := 0; lI < LogsPerBatch; lI++ {
				key := fmt.Sprintf("%04d-%04d-%06d", cI, bI, lI)
				attrs := make(map[string]interface{})
				attrs["batch"] = batchKey
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
				expectedKeys[key] = 1
			}

			t.Logf("Adding batch: %s (%d)", batchKey, cI)
			go (func(batch []*add_events.EventBundle) {
				err := sc.AddEvents(batch)
				assert.Nil(t, err)
			})(batch)
			time.Sleep(MaxDelay)
		}
		time.Sleep(2 * PurgeOlderThan)
		stats := sc.Statistics()
		assert.Greater(t, stats.Sessions.SessionsClosed(), uint64(0))
		assert.LessOrEqual(t, stats.Sessions.SessionsClosed(), stats.Sessions.SessionsOpened())
	}

	for {
		if time.Now().UnixNano()-lastCall.Load() > 5*time.Second.Nanoseconds() {
			break
		}
		time.Sleep(time.Second)
	}

	err = sc.Shutdown()
	assert.Nil(t, err, err)

	stats := sc.Statistics()
	assert.Equal(t, uint64(ExpectedLogs), stats.Events.Enqueued())
	assert.Equal(t, uint64(ExpectedLogs), stats.Events.Processed())
	assert.Equal(t, uint64(0), stats.Events.Waiting())
	assert.Equal(t, uint64(0), stats.Events.Dropped())
	assert.Equal(t, uint64(0), stats.Events.Broken())
	assert.Equal(t, 1.0, stats.Events.SuccessRate())

	assert.Equal(t, uint64(0), stats.Buffers.Waiting())
	assert.Equal(t, uint64(0), stats.Buffers.Dropped())
	assert.Equal(t, uint64(0), stats.Buffers.Broken())
	assert.Equal(t, 1.0, stats.Buffers.SuccessRate())

	assert.Equal(t, 1.0, stats.Transfer.SuccessRate())

	assert.Equal(t, uint64(Cycles*MaxBatchCount), stats.Sessions.SessionsOpened())
	assert.Greater(t, stats.Sessions.SessionsClosed(), uint64(0))
	assert.LessOrEqual(t, stats.Sessions.SessionsClosed(), stats.Sessions.SessionsOpened())

	assert.Equal(t, seenKeys, expectedKeys)
	assert.Equal(t, int(processedEvents.Load()), int(ExpectedLogs), "processed items")
	assert.Equal(t, int(len(seenKeys)), int(ExpectedLogs), "unique items")
}
