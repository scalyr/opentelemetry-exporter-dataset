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

func TestTrimAttrsKeepEverything(t *testing.T) {
	attrs := map[string]interface{}{
		"foo": "aaaaaa",
		"bar": "bbbb",
		"baz": "cc",
	}

	assert.Equal(t, attrs, TrimAttrs(attrs, 10000))
}

func TestTrimAttrsTrimLongest(t *testing.T) {
	attrs := map[string]interface{}{
		"foo": "aaaaaa",
		"bar": "bbbb",
		"baz": "cc",
	}

	expected := map[string]interface{}{
		"foo": "aaaa",
		"bar": "bbbb",
		"baz": "cc",
	}

	assert.Equal(t, expected, TrimAttrs(attrs, 44))
}

func TestTrimAttrsSkipLongest(t *testing.T) {
	attrs := map[string]interface{}{
		"foo": "aaaaaa",
		"bar": "bbbb",
		"baz": "cc",
	}

	expected := map[string]interface{}{
		"bar": "bbbb",
		"baz": "cc",
	}

	assert.Equal(t, expected, TrimAttrs(attrs, 36))
}

func TestTrimAttrsSkipLongestAndTrim(t *testing.T) {
	attrs := map[string]interface{}{
		"foo": "aaaaaa",
		"bar": "bbbb",
		"baz": "cc",
	}

	expected := map[string]interface{}{
		"bar": "bb",
		"baz": "cc",
	}

	assert.Equal(t, expected, TrimAttrs(attrs, 26))
}

func TestTrimAttrsSkipLongestAndTrimFully(t *testing.T) {
	attrs := map[string]interface{}{
		"foo": "aaaaaa",
		"bar": "bbbb",
		"baz": "cc",
	}

	expected := map[string]interface{}{
		"bar": "",
		"baz": "cc",
	}

	assert.Equal(t, expected, TrimAttrs(attrs, 24))
}

func TestTrimAttrsSkipAll(t *testing.T) {
	attrs := map[string]interface{}{
		"foo": "aaaaaa",
		"bar": "bbbb",
		"baz": "cc",
	}

	expected := map[string]interface{}{}

	assert.Equal(t, expected, TrimAttrs(attrs, 1))
}
