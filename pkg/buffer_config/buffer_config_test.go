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
package buffer_config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfigWithOptions(t *testing.T) {
	bufCfg, errB := New(
		WithMaxLifetime(3*time.Second),
		WithPurgeOlderThan(20*time.Second),
		WithMaxSize(12345),
		WithGroupBy([]string{"aaa", "bbb"}),
		WithRetryInitialInterval(8*time.Second),
		WithRetryMaxInterval(30*time.Second),
		WithRetryMaxElapsedTime(10*time.Minute),
		WithRetryShutdownTimeout(2*time.Minute),
	)

	assert.Nil(t, errB)

	assert.Equal(t, DataSetBufferSettings{
		MaxLifetime:          3 * time.Second,
		PurgeOlderThan:       20 * time.Second,
		MaxSize:              12345,
		GroupBy:              []string{"aaa", "bbb"},
		RetryInitialInterval: 8 * time.Second,
		RetryMaxInterval:     30 * time.Second,
		RetryMaxElapsedTime:  10 * time.Minute,
		RetryShutdownTimeout: 2 * time.Minute,
	}, *bufCfg)
}

func TestDataConfigUpdate(t *testing.T) {
	bufCfg, errB := New(
		WithMaxLifetime(3*time.Second),
		WithPurgeOlderThan(20*time.Second),
		WithMaxSize(12345),
		WithGroupBy([]string{"aaa", "bbb"}),
		WithRetryInitialInterval(8*time.Second),
		WithRetryMaxInterval(30*time.Second),
		WithRetryMaxElapsedTime(10*time.Minute),
		WithRetryShutdownTimeout(2*time.Minute),
	)
	assert.Nil(t, errB)

	assert.Equal(t, DataSetBufferSettings{
		MaxLifetime:          3 * time.Second,
		PurgeOlderThan:       20 * time.Second,
		MaxSize:              12345,
		GroupBy:              []string{"aaa", "bbb"},
		RetryInitialInterval: 8 * time.Second,
		RetryMaxInterval:     30 * time.Second,
		RetryMaxElapsedTime:  10 * time.Minute,
		RetryShutdownTimeout: 2 * time.Minute,
	}, *bufCfg)

	bufCfg2, err := bufCfg.WithOptions(
		WithMaxLifetime(23*time.Second),
		WithPurgeOlderThan(220*time.Second),
		WithMaxSize(212345),
		WithGroupBy([]string{"2aaa", "2bbb"}),
		WithRetryInitialInterval(28*time.Second),
		WithRetryMaxInterval(230*time.Second),
		WithRetryMaxElapsedTime(210*time.Minute),
		WithRetryShutdownTimeout(5*time.Minute),
	)
	assert.Nil(t, err)

	// original config is unchanged
	assert.Equal(t, DataSetBufferSettings{
		MaxLifetime:          3 * time.Second,
		PurgeOlderThan:       20 * time.Second,
		MaxSize:              12345,
		GroupBy:              []string{"aaa", "bbb"},
		RetryInitialInterval: 8 * time.Second,
		RetryMaxInterval:     30 * time.Second,
		RetryMaxElapsedTime:  10 * time.Minute,
		RetryShutdownTimeout: 2 * time.Minute,
	}, *bufCfg)

	// new config is changed
	assert.Equal(t, DataSetBufferSettings{
		MaxLifetime:          23 * time.Second,
		PurgeOlderThan:       220 * time.Second,
		MaxSize:              212345,
		GroupBy:              []string{"2aaa", "2bbb"},
		RetryInitialInterval: 28 * time.Second,
		RetryMaxInterval:     230 * time.Second,
		RetryMaxElapsedTime:  210 * time.Minute,
		RetryShutdownTimeout: 5 * time.Minute,
	}, *bufCfg2)
}

func TestDataConfigNewDefaultToString(t *testing.T) {
	cfg := NewDefaultDataSetBufferSettings()
	assert.Equal(t, "MaxLifetime: 5s, PurgeOlderThan: 30s, MaxSize: 6225920, GroupBy: [], RetryRandomizationFactor: 0.500000, RetryMultiplier: 1.500000, RetryInitialInterval: 5s, RetryMaxInterval: 30s, RetryMaxElapsedTime: 5m0s, RetryShutdownTimeout: 30s", cfg.String())
}

func TestDataConfigNewDefaultIsValid(t *testing.T) {
	cfg := NewDefaultDataSetBufferSettings()
	assert.Nil(t, cfg.Validate())
}
