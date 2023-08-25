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

package util

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetEnvWithDefaultIsSet(t *testing.T) {
	key := fmt.Sprintf("ENV_VARIABLE_%d", time.Now().Nanosecond())
	t.Setenv(key, "foo")
	assert.Equal(t, "foo", GetEnvVariableOrDefault(key, "bar"))
}

func TestGetEnvWithDefaultUseDefault(t *testing.T) {
	key := fmt.Sprintf("ENV_VARIABLE_%d", time.Now().Nanosecond())
	assert.Equal(t, "bar", GetEnvVariableOrDefault(key, "bar"))
}
