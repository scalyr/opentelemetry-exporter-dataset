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

package add_events

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventBundle(t *testing.T) {
	event := &Event{
		Thread: "5",
		Sev:    3,
		Ts:     "0",
		Attrs: map[string]interface{}{
			"foo": "a",
			"bar": "b",
			"baz": "a",
		},
	}
	bundle := &EventBundle{Event: event}

	keyFoo := bundle.Key([]string{"foo"})
	keyBar := bundle.Key([]string{"bar"})
	keyBaz := bundle.Key([]string{"baz"})
	keyNotThere1 := bundle.Key([]string{"notThere1"})
	keyNotThere2 := bundle.Key([]string{"notThere2"})

	assert.Equal(t, "ef9faec68698672038857b2647429002", keyFoo)
	assert.Equal(t, "55a2f7ebf2af8927837c599131d32d07", keyBar)
	assert.Equal(t, "6dd515483537f552fd5fa604cd60f0d9", keyBaz)
	assert.Equal(t, "d41d8cd98f00b204e9800998ecf8427e", keyNotThere1)
	assert.Equal(t, "d41d8cd98f00b204e9800998ecf8427e", keyNotThere2)

	// although the value is same, key should be different because attributes differ
	assert.NotEqual(t, keyBaz, keyFoo)
	// non-existing attributes should have the same key
	assert.Equal(t, keyNotThere1, keyNotThere2)
}
