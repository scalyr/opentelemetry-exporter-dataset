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
package config

import (
	"testing"
	"time"

	"github.com/scalyr/dataset-go/pkg/server_host_config"

	"github.com/stretchr/testify/assert"

	"github.com/scalyr/dataset-go/pkg/buffer_config"
)

func TestNewConfigFromEnvUsesDefaultsIfNotSet(t *testing.T) {
	cfg1, err := New(FromEnv())
	assert.Nil(t, err)
	assert.Equal(t, "", cfg1.Endpoint)
}

func TestNewConfigFromEnvUsingEnvServer(t *testing.T) {
	t.Setenv("SCALYR_SERVER", "http://test")
	cfg2, err := New(FromEnv())
	assert.Nil(t, err)
	assert.Equal(t, cfg2.Endpoint, "http://test")
}

func TestNewConfigFromEnvUsingEnvTokens(t *testing.T) {
	// GIVEN
	t.Setenv("SCALYR_WRITELOG_TOKEN", "writelog")
	t.Setenv("SCALYR_READLOG_TOKEN", "readlog")
	t.Setenv("SCALYR_WRITECONFIG_TOKEN", "writeconfig")
	t.Setenv("SCALYR_READCONFIG_TOKEN", "readconfig")
	// WHEN
	cfg4, err := New(FromEnv())
	// THEN
	assert.Nil(t, err)
	assert.Equal(t, "readlog", cfg4.Tokens.ReadLog)
	assert.Equal(t, "writelog", cfg4.Tokens.WriteLog)
	assert.Equal(t, "readconfig", cfg4.Tokens.ReadConfig)
	assert.Equal(t, "writeconfig", cfg4.Tokens.WriteConfig)
}

func TestNewConfigWithProvidedOptions(t *testing.T) {
	// GIVEN
	bufCfg, errB := buffer_config.New(
		buffer_config.WithMaxLifetime(3*time.Second),
		buffer_config.WithMaxSize(12345),
		buffer_config.WithGroupBy([]string{"aaa", "bbb"}),
		buffer_config.WithRetryInitialInterval(8*time.Second),
		buffer_config.WithRetryMaxInterval(30*time.Second),
		buffer_config.WithRetryMaxElapsedTime(10*time.Minute),
	)
	assert.Nil(t, errB)
	// WHEN
	cfg5, err := New(
		WithEndpoint("https://fooOpt"),
		WithTokens(DataSetTokens{WriteLog: "writeLogOpt"}),
		WithBufferSettings(*bufCfg),
	)
	// THEN
	assert.Nil(t, err)
	assert.Equal(t, "https://fooOpt", cfg5.Endpoint)
	assert.Equal(t, DataSetTokens{WriteLog: "writeLogOpt"}, cfg5.Tokens)
	assert.Equal(t, buffer_config.DataSetBufferSettings{
		MaxLifetime:          3 * time.Second,
		MaxSize:              12345,
		GroupBy:              []string{"aaa", "bbb"},
		RetryInitialInterval: 8 * time.Second,
		RetryMaxInterval:     30 * time.Second,
		RetryMaxElapsedTime:  10 * time.Minute,
	}, cfg5.BufferSettings)
}

func TestNewConfigBasedOnExistingWithNewConfigOptions(t *testing.T) {
	// GIVEN
	bufCfg, _ := buffer_config.New(
		buffer_config.WithMaxLifetime(3*time.Second),
		buffer_config.WithMaxSize(12345),
		buffer_config.WithGroupBy([]string{"aaa", "bbb"}),
		buffer_config.WithRetryInitialInterval(8*time.Second),
		buffer_config.WithRetryMaxInterval(30*time.Second),
		buffer_config.WithRetryMaxElapsedTime(10*time.Minute),
	)
	hostCfg, _ := server_host_config.New(
		server_host_config.WithServerHost("AAA"),
		server_host_config.WithUseHostName(false),
	)
	cfg5, _ := New(
		WithEndpoint("https://fooOpt1"),
		WithTokens(DataSetTokens{WriteLog: "writeLogOpt1"}),
		WithBufferSettings(*bufCfg),
		WithServerHostSettings(*hostCfg),
		WithDebug(true),
	)
	// AND
	bufCfg2, _ := buffer_config.New(
		buffer_config.WithMaxLifetime(23*time.Second),
		buffer_config.WithMaxSize(212345),
		buffer_config.WithGroupBy([]string{"2aaa", "2bbb"}),
		buffer_config.WithRetryInitialInterval(28*time.Second),
		buffer_config.WithRetryMaxInterval(230*time.Second),
		buffer_config.WithRetryMaxElapsedTime(210*time.Minute),
	)
	hostCfg2, _ := server_host_config.New(
		server_host_config.WithServerHost("BBB"),
		server_host_config.WithUseHostName(true),
	)
	// WHEN
	cfg6, _ := cfg5.WithOptions(
		WithEndpoint("https://fooOpt2"),
		WithTokens(DataSetTokens{WriteLog: "writeLogOpt2"}),
		WithBufferSettings(*bufCfg2),
		WithServerHostSettings(*hostCfg2),
	)
	// THEN original config is unchanged
	assert.Equal(t, "https://fooOpt1", cfg5.Endpoint)
	assert.Equal(t, DataSetTokens{WriteLog: "writeLogOpt1"}, cfg5.Tokens)
	assert.Equal(t, buffer_config.DataSetBufferSettings{
		MaxLifetime:          3 * time.Second,
		MaxSize:              12345,
		GroupBy:              []string{"aaa", "bbb"},
		RetryInitialInterval: 8 * time.Second,
		RetryMaxInterval:     30 * time.Second,
		RetryMaxElapsedTime:  10 * time.Minute,
	}, cfg5.BufferSettings)
	assert.Equal(t, server_host_config.DataSetServerHostSettings{
		ServerHost:  "AAA",
		UseHostName: false,
	}, cfg5.ServerHostSettings)
	assert.Equal(t, true, cfg5.Debug)

	// AND new config is changed
	assert.Equal(t, "https://fooOpt2", cfg6.Endpoint)
	assert.Equal(t, DataSetTokens{WriteLog: "writeLogOpt2"}, cfg6.Tokens)
	assert.Equal(t, buffer_config.DataSetBufferSettings{
		MaxLifetime:          23 * time.Second,
		MaxSize:              212345,
		GroupBy:              []string{"2aaa", "2bbb"},
		RetryInitialInterval: 28 * time.Second,
		RetryMaxInterval:     230 * time.Second,
		RetryMaxElapsedTime:  210 * time.Minute,
	}, cfg6.BufferSettings)
	assert.Equal(t, server_host_config.DataSetServerHostSettings{
		ServerHost:  "BBB",
		UseHostName: true,
	}, cfg6.ServerHostSettings)
	assert.Equal(t, true, cfg6.Debug)
}

func TestNewDefaultDataSetConfigToString(t *testing.T) {
	cfg := NewDefaultDataSetConfig()
	assert.Equal(t, cfg.String(), "Endpoint: https://app.scalyr.com, Tokens: (WriteLog: false, ReadLog: false, WriteConfig: false, ReadConfig: false), BufferSettings: (MaxLifetime: 5s, PurgeOlderThan: 30s, MaxSize: 6225920, GroupBy: [], RetryRandomizationFactor: 0.500000, RetryMultiplier: 1.500000, RetryInitialInterval: 5s, RetryMaxInterval: 30s, RetryMaxElapsedTime: 5m0s, RetryShutdownTimeout: 30s), ServerHostSettings: (UseHostName: true, ServerHost: ), Debug: (false)")
}

func TestNewDefaultDataSetConfigIsValid(t *testing.T) {
	cfg := NewDefaultDataSetConfig()
	assert.Nil(t, cfg.Validate())
}

func TestDataSetConfigValidationEmptyEndpoint(t *testing.T) {
	cfg := DataSetConfig{
		Endpoint:       "",
		Tokens:         DataSetTokens{},
		BufferSettings: buffer_config.NewDefaultDataSetBufferSettings(),
	}
	err := cfg.Validate()
	assert.NotNil(t, err)
	assert.Equal(t, "endpoint cannot be empty", err.Error())
}
