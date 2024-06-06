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

package meter_config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
)

func TestNewMeterConfig(t *testing.T) {
	meter := otel.Meter("AAA")
	cfg := NewMeterConfig(&meter, "entity", "name")
	assert.Equal(t, "entity", cfg.Entity())
	assert.Equal(t, "name", cfg.Name())
	assert.NotNil(t, cfg.Meter())
}
