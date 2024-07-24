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

package session_manager

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestBasicFlow(t *testing.T) {
	// GIVEN
	seenKeys := make(map[string]int64)
	seenMutex := &sync.Mutex{}

	callback := func(key string, bundlesChannel <-chan interface{}) {
		for {
			msg, ok := <-bundlesChannel
			if !ok {
				return
			}
			seenMutex.Lock()
			seenKeys[fmt.Sprintf("%s-%s", key, msg)] = 1
			seenMutex.Unlock()
		}
		// do nothing
	}
	logger := zap.NewNop()
	manager := New(logger, callback)

	// WHEN
	manager.Sub("aaa")
	manager.Pub("aaa", "val-1")
	manager.Pub("aaa", "val-2")
	manager.Sub("bbb")
	manager.Pub("bbb", "val-1")
	manager.Pub("bbb", "val-2")
	manager.Unsub("aaa")
	manager.Unsub("bbb")

	// THEN
	time.Sleep(time.Second)
	seenMutex.Lock()
	assert.Equal(t, map[string]int64{"aaa-val-1": 1, "aaa-val-2": 1, "bbb-val-1": 1, "bbb-val-2": 1}, seenKeys)
	seenMutex.Unlock()
}
