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

package statistics

import (
	"math/rand"
	"testing"
	"time"

	"github.com/scalyr/dataset-go/pkg/meter_config"

	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
)

var (
	meter = otel.Meter("test")
	tests = []struct {
		name  string
		meter *meter_config.MeterConfig
	}{
		{
			name:  "no config",
			meter: nil,
		},
		{
			name:  "no meter",
			meter: meter_config.NewMeterConfig(nil, "e", "n"),
		},
		{
			name:  "with meter",
			meter: meter_config.NewMeterConfig(&meter, "e", "n"),
		},
	}
)

func TestBuffersEnqueued(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			stats, err := NewStatistics(tt.meter, zap.Must(zap.NewDevelopment()))
			require.Nil(t, err)
			v := uint64(rand.Int())
			stats.BuffersEnqueuedAdd(v)
			assert.Equal(t, v, stats.BuffersEnqueued())
		})
	}
}

func TestBuffersProcessed(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			stats, err := NewStatistics(tt.meter, zap.Must(zap.NewDevelopment()))

			require.Nil(t, err)
			v := uint64(rand.Int())
			stats.BuffersProcessedAdd(v)
			assert.Equal(t, v, stats.BuffersProcessed())
		})
	}
}

func TestBuffersDropped(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			stats, err := NewStatistics(tt.meter, zap.Must(zap.NewDevelopment()))

			require.Nil(t, err)
			v := uint64(rand.Int())
			stats.BuffersDroppedAdd(v)
			assert.Equal(t, v, stats.BuffersDropped())
		})
	}
}

func TestBuffersBroken(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			stats, err := NewStatistics(tt.meter, zap.Must(zap.NewDevelopment()))
			require.Nil(t, err)
			v := uint64(rand.Int())
			stats.BuffersBrokenAdd(v)
			assert.Equal(t, v, stats.BuffersBroken())
		})
	}
}

func TestEventsEnqueued(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			stats, err := NewStatistics(tt.meter, zap.Must(zap.NewDevelopment()))
			require.Nil(t, err)
			v := uint64(rand.Int())
			stats.EventsEnqueuedAdd(v)
			assert.Equal(t, v, stats.EventsEnqueued())
		})
	}
}

func TestEventsProcessed(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			stats, err := NewStatistics(tt.meter, zap.Must(zap.NewDevelopment()))

			require.Nil(t, err)
			v := uint64(rand.Int())
			stats.EventsProcessedAdd(v)
			assert.Equal(t, v, stats.EventsProcessed())
		})
	}
}

func TestEventsDropped(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			stats, err := NewStatistics(tt.meter, zap.Must(zap.NewDevelopment()))

			require.Nil(t, err)
			v := uint64(rand.Int())
			stats.EventsDroppedAdd(v)
			assert.Equal(t, v, stats.EventsDropped())
		})
	}
}

func TestEventsBroken(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			stats, err := NewStatistics(tt.meter, zap.Must(zap.NewDevelopment()))
			require.Nil(t, err)
			v := uint64(rand.Int())
			stats.EventsBrokenAdd(v)
			assert.Equal(t, v, stats.EventsBroken())
		})
	}
}

func TestBytesAPISent(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			stats, err := NewStatistics(tt.meter, zap.Must(zap.NewDevelopment()))
			require.Nil(t, err)
			v := uint64(rand.Int())
			stats.BytesAPISentAdd(v)
			assert.Equal(t, v, stats.BytesAPISent())
		})
	}
}

func TestBytesAPIAccepted(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			stats, err := NewStatistics(tt.meter, zap.Must(zap.NewDevelopment()))
			require.Nil(t, err)
			v := uint64(rand.Int())
			stats.BytesAPIAcceptedAdd(v)
			assert.Equal(t, v, stats.BytesAPIAccepted())
		})
	}
}

func TestSessionsOpened(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			stats, err := NewStatistics(tt.meter, zap.Must(zap.NewDevelopment()))
			require.Nil(t, err)
			v := uint64(rand.Int())
			stats.SessionsOpenedAdd(v)
			assert.Equal(t, v, stats.SessionsOpened())
		})
	}
}

func TestSessionsClosed(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			stats, err := NewStatistics(tt.meter, zap.Must(zap.NewDevelopment()))
			require.Nil(t, err)
			v := uint64(rand.Int())
			stats.SessionsClosedAdd(v)
			assert.Equal(t, v, stats.SessionsClosed())
		})
	}
}

func TestExport(t *testing.T) {
	stats, err := NewStatistics(nil, zap.Must(zap.NewDevelopment()))
	require.Nil(t, err)

	stats.BuffersEnqueuedAdd(1000)
	stats.BuffersProcessedAdd(100)
	stats.BuffersDroppedAdd(10)
	stats.BuffersBrokenAdd(1)

	stats.EventsEnqueuedAdd(2000)
	stats.EventsProcessedAdd(200)
	stats.EventsDroppedAdd(20)
	stats.EventsBrokenAdd(2)

	stats.BytesAPISentAdd(3000)
	stats.BytesAPIAcceptedAdd(300)

	stats.SessionsOpenedAdd(4000)
	stats.SessionsClosedAdd(400)

	exp := stats.Export(time.Second)
	assert.Equal(t, &ExportedStatistics{
		Buffers: QueueStats{
			enqueued:       1000,
			processed:      100,
			dropped:        10,
			broken:         1,
			processingTime: time.Second,
		},
		Events: QueueStats{
			enqueued:       2000,
			processed:      200,
			dropped:        20,
			broken:         2,
			processingTime: time.Second,
		},
		Transfer: TransferStats{
			bytesSent:        3000,
			bytesAccepted:    300,
			buffersProcessed: 100,
			processingTime:   time.Second,
		},
		Sessions: SessionsStats{
			sessionsOpened: 4000,
			sessionsClosed: 400,
		},
	}, exp)

	assert.Equal(t, exp.Buffers.Enqueued(), uint64(1000))
	assert.Equal(t, exp.Buffers.Processed(), uint64(100))
	assert.Equal(t, exp.Buffers.Dropped(), uint64(10))
	assert.Equal(t, exp.Buffers.Broken(), uint64(1))
	assert.Equal(t, exp.Buffers.Waiting(), uint64(889))
	assert.Equal(t, exp.Buffers.ProcessingTime(), time.Second)
	assert.Equal(t, exp.Buffers.SuccessRate(), .89)

	assert.Equal(t, exp.Events.Enqueued(), uint64(2000))
	assert.Equal(t, exp.Events.Processed(), uint64(200))
	assert.Equal(t, exp.Events.Dropped(), uint64(20))
	assert.Equal(t, exp.Events.Broken(), uint64(2))
	assert.Equal(t, exp.Events.Waiting(), uint64(1778))
	assert.Equal(t, exp.Events.ProcessingTime(), time.Second)
	assert.Equal(t, exp.Events.SuccessRate(), .89)

	assert.Equal(t, exp.Transfer.BytesSent(), uint64(3000))
	assert.Equal(t, exp.Transfer.BytesAccepted(), uint64(300))
	assert.Equal(t, exp.Transfer.BuffersProcessed(), uint64(100))
	assert.Equal(t, exp.Transfer.SuccessRate(), float64(0.1))
	assert.Equal(t, exp.Transfer.ThroughputBpS(), float64(300))
	assert.Equal(t, exp.Transfer.AvgBufferBytes(), float64(3))
	assert.Equal(t, exp.Transfer.ProcessingTime(), time.Second)

	assert.Equal(t, exp.Sessions.SessionsOpened(), uint64(4000))
	assert.Equal(t, exp.Sessions.SessionsClosed(), uint64(400))
	assert.Equal(t, exp.Sessions.SessionsActive(), uint64(3600))
}

func TestExportNoTraffic(t *testing.T) {
	stats, err := NewStatistics(nil, zap.Must(zap.NewDevelopment()))
	require.Nil(t, err)

	exp := stats.Export(time.Second)
	assert.Equal(t, exp.Events.SuccessRate(), float64(0))
	assert.Equal(t, exp.Buffers.SuccessRate(), float64(0))
	assert.Equal(t, exp.Transfer.SuccessRate(), float64(0))
	assert.Equal(t, exp.Transfer.ThroughputBpS(), float64(0))
	assert.Equal(t, exp.Transfer.AvgBufferBytes(), float64(0))
}
