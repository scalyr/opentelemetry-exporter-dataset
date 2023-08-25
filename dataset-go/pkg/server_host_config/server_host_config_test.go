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
package server_host_config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDataSetServerHostSettingsWithProvidedOptions(t *testing.T) {
	// WHEN
	cfg5, err := New(
		WithUseHostName(true),
		WithServerHost("AAA"),
	)
	// THEN
	assert.Nil(t, err)
	assert.Equal(t, true, cfg5.UseHostName)
	assert.Equal(t, "AAA", cfg5.ServerHost)
}

func TestNewDataSetServerHostSettingsBasedOnExistingWithNewConfigOptions(t *testing.T) {
	// GIVEN
	cfg5, err := New(
		WithUseHostName(true),
		WithServerHost("AAA"),
	)
	assert.Nil(t, err)
	// AND

	cfg6, _ := cfg5.WithOptions(
		WithUseHostName(false),
		WithServerHost("BBB"),
	)
	// THEN original config is unchanged
	assert.Equal(t, true, cfg5.UseHostName)
	assert.Equal(t, "AAA", cfg5.ServerHost)

	// AND new config is changed
	assert.Equal(t, false, cfg6.UseHostName)
	assert.Equal(t, "BBB", cfg6.ServerHost)
}

func TestNewDefaultDataSetServerHostSettingsToString(t *testing.T) {
	cfg := NewDefaultDataSetServerHostSettings()
	assert.Equal(t, cfg.String(), "UseHostName: true, ServerHost: ")
}

func TestNewDefaultDataSetServerHostSettingsIsValid(t *testing.T) {
	cfg := NewDefaultDataSetServerHostSettings()
	assert.Nil(t, cfg.Validate())
}

func TestDataSetServerHostConfigValidationMissingServerHost(t *testing.T) {
	cfg := DataSetServerHostSettings{
		UseHostName: false,
		ServerHost:  "",
	}
	err := cfg.Validate()
	assert.NotNil(t, err)
	assert.Equal(t, "when UseHostName is False, then ServerHost has to be set", err.Error())
}
