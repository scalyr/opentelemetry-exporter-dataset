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

package buffer

import (
	"encoding/json"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/scalyr/dataset-go/pkg/api/add_events"
	"github.com/stretchr/testify/assert"
)

func TestAddStatusString(t *testing.T) {
	assert.Equal(t, "Added", Added.String())
	assert.NotEmpty(t, "Skipped", Skipped.String())
	assert.NotEmpty(t, "TooMuch", TooMuch.String())
}

func loadJson(name string) string {
	dat, err := os.ReadFile("../../test/testdata/" + name)
	if err != nil {
		panic(err)
	}

	return strings.Join(strings.Fields(string(dat)), "")
}

func TestEmptyPayloadShouldFail(t *testing.T) {
	buffer, err := NewBuffer("id", "token", nil)
	assert.Nil(t, err)
	_, err = buffer.Payload()
	assert.EqualError(t, err, "there is no event")
}

func TestEmptyTokenShouldFail(t *testing.T) {
	buffer, err := NewBuffer("id", "", nil)
	assert.Nil(t, err)
	_, err = buffer.Payload()
	assert.EqualError(t, err, "token is missing")
}

func TestEmptySessionShouldFail(t *testing.T) {
	buffer, err := NewBuffer("", "token", nil)
	assert.Nil(t, err)
	_, err = buffer.Payload()
	assert.EqualError(t, err, "session is missing")
}

func createTestBundle() add_events.EventBundle {
	return add_events.EventBundle{
		Log: &add_events.Log{Id: "LId", Attrs: map[string]interface{}{
			"LAttr1": "LVal1",
			"LAttr2": "LVal2",
		}},
		Thread: &add_events.Thread{Id: "TId", Name: "TName"},
		Event: &add_events.Event{
			Thread: "TId",
			Sev:    3,
			Ts:     "0",
			Attrs: map[string]interface{}{
				"message": "test",
				"s1web":   "CentralizeSentinelOne-nativeendpoint,cloud,andidentitytelemetrywithanyopen,thirdpartydatafromyoursecurityecosystemintoonepowerfulplatform.Don’tstopatjustidentifyingmaliciousbehaviors.Blockandremediateadvancedattacksautonomously,atmachinespeed,withcross-platform,enterprise-scaledataanalytics.Empoweranalystswiththecontexttheyneed,faster,byautomaticallyconnecting,correlatingbenignandmaliciouseventsinoneillustrativeview.",
			},
		},
	}
}

func createSessionInfo() *add_events.SessionInfo {
	return &add_events.SessionInfo{
		"serverId":   "serverId",
		"serverType": "serverType",
		"region":     "region",
	}
}

func createEmptyBuffer() *Buffer {
	sessionInfo := createSessionInfo()
	session := "session"
	token := "token"
	buffer, err := NewBuffer(
		session,
		token,
		sessionInfo)
	if err != nil {
		return nil
	}
	return buffer
}

func TestAddBundle(t *testing.T) {
	tests := []struct {
		name      string
		bundle    add_events.EventBundle
		expStatus AddStatus
		expError  error
	}{
		{
			name: "Add small",
			bundle: add_events.EventBundle{
				Log: &add_events.Log{Id: "LId", Attrs: map[string]interface{}{
					"LAttr1": "LVal1",
					"LAttr2": "LVal2",
				}},
				Thread: &add_events.Thread{Id: "TId-2", Name: "TName"},
				Event: &add_events.Event{
					Thread: "TId",
					Sev:    3,
					Ts:     "0",
					Attrs: map[string]interface{}{
						"message": "test",
						"s1web":   "aaa",
					},
				},
			},
			expStatus: Added,
			expError:  nil,
		},
		{
			name: "Add with large thread name",
			bundle: add_events.EventBundle{
				Log: &add_events.Log{Id: "LId", Attrs: map[string]interface{}{
					"LAttr1": "LVal1",
					"LAttr2": "LVal2",
				}},
				Thread: &add_events.Thread{Id: "TId-2", Name: strings.Repeat("T", LimitBufferSize)},
				Event: &add_events.Event{
					Thread: "TId",
					Sev:    3,
					Ts:     "0",
					Attrs: map[string]interface{}{
						"message": "test",
						"s1web":   "aaa",
					},
				},
			},
			expStatus: TooMuch,
			expError:  nil,
		},
		{
			name: "Add with large log",
			bundle: add_events.EventBundle{
				Log: &add_events.Log{Id: "LId-2", Attrs: map[string]interface{}{
					"LAttr1": "LVal1",
					"LAttr2": "LVal2",
					"LAttr3": strings.Repeat("L", LimitBufferSize),
				}},
				Thread: &add_events.Thread{Id: "TId-2", Name: "Name"},
				Event: &add_events.Event{
					Thread: "TId",
					Sev:    3,
					Ts:     "0",
					Attrs: map[string]interface{}{
						"message": "test",
						"s1web":   "aaa",
					},
				},
			},
			expStatus: TooMuch,
			expError:  nil,
		},
		{
			name: "Add with large event",
			bundle: add_events.EventBundle{
				Log: &add_events.Log{Id: "LId-2", Attrs: map[string]interface{}{
					"LAttr1": "LVal1",
					"LAttr2": "LVal2",
				}},
				Thread: &add_events.Thread{Id: "TId-2", Name: "Name"},
				Event: &add_events.Event{
					Thread: "TId",
					Sev:    3,
					Ts:     "0",
					Attrs: map[string]interface{}{
						"message": "test",
						"s1web":   "aaa",
						"attr":    strings.Repeat("E", LimitBufferSize),
					},
				},
			},
			expStatus: TooMuch,
			expError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			buf := createEmptyBuffer()
			bundle := createTestBundle()
			status, err := buf.AddBundle(&bundle)
			assert.Equal(t, status, Added)
			assert.Nil(t, err)

			sizeBefore := buf.BufferLengths()

			status, err = buf.AddBundle(&tt.bundle)
			assert.Equal(t, tt.expStatus, status)
			assert.Equal(t, tt.expError, err)

			if status == TooMuch {
				assert.Equal(t, int32(1), buf.CountEvents())
				assert.Equal(t, sizeBefore, buf.BufferLengths())
			}
		})
	}
}

func TestAddBundleWithBufferTrimming(t *testing.T) {
	buf := createEmptyBuffer()
	bundle := add_events.EventBundle{
		Log: &add_events.Log{Id: "LId-2", Attrs: map[string]interface{}{
			"LAttr1": "LVal1",
			"LAttr2": "LVal2",
		}},
		Thread: &add_events.Thread{Id: "TId-2", Name: "Name"},
		Event: &add_events.Event{
			Thread: "TId",
			Sev:    3,
			Ts:     "0",
			Attrs: map[string]interface{}{
				"message": "test",
				"s1web":   "aaa",
				"attr":    strings.Repeat("E", LimitBufferSize),
			},
		},
	}
	status, err := buf.AddBundle(&bundle)
	assert.Nil(t, err)
	assert.Equal(t, status, Added)
	assert.Equal(t, int32(1), buf.CountEvents())
	assert.Equal(t, int32(6225911), buf.BufferLengths())
}

func TestPayloadFull(t *testing.T) {
	buffer := createEmptyBuffer()
	assert.NotNil(t, buffer)

	bundle := createTestBundle()
	added, err := buffer.AddBundle(&bundle)
	assert.Nil(t, err)
	assert.Equal(t, added, Added)
	assert.Equal(t, buffer.countLogs, int32(1))
	assert.Equal(t, buffer.lenLogs, int32(57))
	assert.Equal(t, buffer.countThreads, int32(1))
	assert.Equal(t, buffer.lenThreads, int32(28))

	payload, err := buffer.Payload()
	assert.Nil(t, err)

	params := add_events.AddEventsRequest{}
	err = json.Unmarshal(payload, &params)
	assert.Nil(t, err)
	assert.Equal(t, params.Session, buffer.Session)
	assert.Equal(t, params.Token, buffer.Token)
	assert.Equal(t, params.SessionInfo, buffer.sessionInfo)

	expected := loadJson("buffer_test_payload_full.json")

	assert.Equal(t, len(payload), len(expected), "Length differs")

	upperBound := int(math.Min(float64(len(expected)), float64(len(payload))))
	for i := 0; i < upperBound; i++ {
		if expected[i] != (payload)[i] {
			assert.Equal(t, string(payload[0:i]), expected[0:i], "Pos: %d", i)
		}
	}
	assert.Equal(t, payload, []byte(expected))
}

func TestPayloadInjection(t *testing.T) {
	sessionInfo := &add_events.SessionInfo{
		"serverId":   "serverId\",\"sI\":\"I",
		"serverType": "serverType\",\"sT\":\"T",
		"region":     "region\",\"r\":\"R",
	}
	session := "session\",\"s\":\"S"
	token := "token\",\"events\":[{}],\"foo\":\"bar"
	buffer, err := NewBuffer(
		session,
		token,
		sessionInfo)

	assert.Nil(t, err)
	bundle := &add_events.EventBundle{
		Log: &add_events.Log{
			Id: "LId\",\"i\":\"i",
			Attrs: map[string]interface{}{
				"LAttr1\",\"i\":\"i": "LVal1\",\"i\":\"i",
				"LAttr2\",\"i\":\"i": "LVal2\",\"i\":\"i",
			},
		},
		Thread: &add_events.Thread{Id: "TId\",\"i\":\"i", Name: "TName\",\"i\":\"i"},
		Event: &add_events.Event{
			Thread: "TId\",\"i\":\"i",
			Sev:    3,
			Ts:     "0",
			Attrs: map[string]interface{}{
				"message\",\"i\":\"i": "test\",\"i\":\"i",
				"meh\",\"i\":\"i":     1.0,
			},
		},
	}
	added, err := buffer.AddBundle(bundle)
	assert.Nil(t, err)
	assert.Equal(t, added, Added)

	assert.Equal(t, buffer.countLogs, int32(1))
	assert.Equal(t, buffer.lenLogs, int32(117))

	assert.Equal(t, buffer.countThreads, int32(1))
	assert.Equal(t, buffer.lenThreads, int32(52))

	payload, err := buffer.Payload()
	assert.Nil(t, err)

	params := add_events.AddEventsRequest{}
	err = json.Unmarshal(payload, &params)
	assert.Nil(t, err)
	assert.Equal(t, params.Session, session)
	assert.Equal(t, params.Token, token)
	assert.Equal(t, len(params.Events), 1)

	expected := loadJson("buffer_test_payload_injection.json")

	assert.Equal(t, len(payload), len(expected), "Length differs")

	upperBound := int(math.Min(float64(len(expected)), float64(len(payload))))
	for i := 0; i < upperBound; i++ {
		if expected[i] != (payload)[i] {
			assert.Equal(t, payload[0:i], expected[0:i], "Pos: %d", i)
		}
	}
	assert.Equal(t, payload, []byte(expected))
}

func TestAddEventWithShouldSendAge(t *testing.T) {
	buffer := createEmptyBuffer()
	assert.NotNil(t, buffer)

	bundle := createTestBundle()
	added, err := buffer.AddBundle(&bundle)
	assert.Nil(t, err)
	assert.Equal(t, added, Added)
	assert.False(t, buffer.ShouldSendAge(100*time.Millisecond))

	time.Sleep(10 * time.Millisecond)

	assert.True(t, buffer.ShouldSendAge(time.Millisecond))
}

func TestAddEventWithShouldPurgeAge(t *testing.T) {
	buffer := createEmptyBuffer()
	assert.NotNil(t, buffer)

	bundle := createTestBundle()
	assert.False(t, buffer.ShouldPurgeAge(100*time.Millisecond))

	time.Sleep(10 * time.Millisecond)

	assert.True(t, buffer.ShouldPurgeAge(time.Millisecond))

	added, err := buffer.AddBundle(&bundle)
	assert.Nil(t, err)
	assert.Equal(t, added, Added)
	assert.False(t, buffer.ShouldPurgeAge(100*time.Millisecond))

	time.Sleep(10 * time.Millisecond)

	assert.False(t, buffer.ShouldPurgeAge(time.Millisecond))
}

func TestAddEventWithShouldSendSize(t *testing.T) {
	buffer := createEmptyBuffer()
	assert.NotNil(t, buffer)

	for {
		bundle := createTestBundle()
		added, err := buffer.AddBundle(&bundle)
		assert.Nil(t, err)
		if added != Added {
			break
		}
		assert.Equal(t, added, Added)
		time.Sleep(25 * time.Microsecond)
	}

	assert.True(t, buffer.ShouldSendSize())
	assert.True(t, buffer.HasEvents())
	assert.Greater(t, buffer.CountEvents(), int32(10))
	assert.Greater(t, buffer.BufferLengths(), int32(ShouldSentBufferSize-1000))
}

func TestNewEmpty(t *testing.T) {
	buffer := createEmptyBuffer()
	assert.NotNil(t, buffer)

	bundle := createTestBundle()
	added, err := buffer.AddBundle(&bundle)
	assert.Nil(t, err)
	assert.Equal(t, added, Added)
	assert.Equal(t, int32(1), buffer.CountEvents())
	assert.Equal(t, int32(646), buffer.BufferLengths())

	newBuf, err := buffer.NewEmpty()
	assert.Nil(t, err)
	assert.NotNil(t, newBuf)
	assert.Equal(t, int32(0), newBuf.CountEvents())
	assert.Equal(t, int32(67), newBuf.BufferLengths())
}

func TestZapStats(t *testing.T) {
	buf := createEmptyBuffer()
	bundle := createTestBundle()
	status, err := buf.AddBundle(&bundle)
	assert.Nil(t, err)
	assert.Equal(t, status, Added)
	stats := buf.ZapStats()
	assert.Equal(t, 8, len(stats))
	assert.Equal(t, zap.String("session", "session"), stats[1])
	assert.Equal(t, zap.Int32("logs", 1), stats[2])
	assert.Equal(t, zap.Int32("threads", 1), stats[3])
	assert.Equal(t, zap.Int32("events", 1), stats[4])
	assert.Equal(t, zap.Int32("bufferLength", 646), stats[5])
	assert.Equal(t, zap.Float64("bufferRatio", float64(646)/ShouldSentBufferSize), stats[6])
}
