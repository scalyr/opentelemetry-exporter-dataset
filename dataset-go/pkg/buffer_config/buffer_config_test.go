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
	"fmt"
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
		WithMaxParallelOutgoing(40),
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
		MaxParallelOutgoing:  40,
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
		WithMaxParallelOutgoing(40),
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
		MaxParallelOutgoing:  40,
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
		WithMaxParallelOutgoing(240),
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
		MaxParallelOutgoing:  40,
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
		MaxParallelOutgoing:  240,
	}, *bufCfg2)
}

func TestDataConfigNewDefaultToString(t *testing.T) {
	cfg := NewDefaultDataSetBufferSettings()
	assert.Equal(t, "MaxLifetime: 5s, PurgeOlderThan: 30s, MaxSize: 6225920, GroupBy: [], RetryRandomizationFactor: 0.500000, RetryMultiplier: 1.500000, RetryInitialInterval: 5s, RetryMaxInterval: 30s, RetryMaxElapsedTime: 5m0s, RetryShutdownTimeout: 30s, MaxParallelOutgoing: 100", cfg.String())
}

func TestDataConfigNewDefaultIsValid(t *testing.T) {
	cfg := NewDefaultDataSetBufferSettings()
	assert.Nil(t, cfg.Validate())
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name     string
		options  []DataSetBufferSettingsOption
		expError error
	}{
		// default config is valid
		{
			name:     "default",
			options:  []DataSetBufferSettingsOption{},
			expError: nil,
		},
		{
			name: "shorter purge",
			options: []DataSetBufferSettingsOption{
				WithMaxLifetime(time.Second),
				WithPurgeOlderThan(time.Second),
			},
			expError: fmt.Errorf("MaxLifetime 1s has to be at least 3 times smaller than PurgeOlderThan 1s"),
		},
		{
			name: "max size too large",
			options: []DataSetBufferSettingsOption{
				WithMaxSize(LimitBufferSize + 1),
			},
			expError: fmt.Errorf("MaxSize has value %d which is more than %d", LimitBufferSize+1, LimitBufferSize),
		},
		{
			name: "max size too small",
			options: []DataSetBufferSettingsOption{
				WithMaxSize(1),
			},
			expError: fmt.Errorf("MaxSize has value %d which is less than %d", 1, MinimalMaxSize),
		},
		{
			name: "retry initial interval too short",
			options: []DataSetBufferSettingsOption{
				WithRetryInitialInterval(MinimalInitialInterval / 2),
			},
			expError: fmt.Errorf("RetryInitialInterval has value %s which is less than %s", MinimalInitialInterval/2, MinimalInitialInterval),
		},
		{
			name: "retry max interval too short",
			options: []DataSetBufferSettingsOption{
				WithRetryMaxInterval(MinimalMaxInterval / 2),
			},
			expError: fmt.Errorf("RetryMaxInterval has value %s which is less than %s", MinimalMaxInterval/2, MinimalMaxInterval),
		},
		{
			name: "retry max elapsed time too short",
			options: []DataSetBufferSettingsOption{
				WithRetryMaxElapsedTime(MinimalMaxElapsedTime / 2),
			},
			expError: fmt.Errorf("RetryMaxElapsedTime has value %s which is less than %s", MinimalMaxElapsedTime/2, MinimalMaxElapsedTime),
		},
		{
			name: "retry multiplier too small",
			options: []DataSetBufferSettingsOption{
				WithRetryMultiplier(MinimalMultiplier / 2),
			},
			expError: fmt.Errorf("RetryMultiplier has value %f which is less or equal than %f", MinimalMultiplier/2, MinimalMultiplier),
		},
		{
			name: "retry randomization factor too small",
			options: []DataSetBufferSettingsOption{
				WithRetryRandomizationFactor(MinimalRandomizationFactor / 2),
			},
			expError: fmt.Errorf("RetryRandomizationFactor has value %f which is less or equal than %f", MinimalRandomizationFactor/2, MinimalRandomizationFactor),
		},
		{
			name: "retry shutdown timeout too small",
			options: []DataSetBufferSettingsOption{
				WithRetryShutdownTimeout(MinimalRetryShutdownTimeout / 2),
			},
			expError: fmt.Errorf("RetryShutdownTimeout has value %s which is less than %s", MinimalRetryShutdownTimeout/2, MinimalRetryShutdownTimeout),
		},
		{
			name: "max size too large",
			options: []DataSetBufferSettingsOption{
				WithMaxParallelOutgoing(MaximalMaxParallelOutgoing + 1),
			},
			expError: fmt.Errorf("MaxParallelOutgoing has value %d which is more than %d", MaximalMaxParallelOutgoing+1, MaximalMaxParallelOutgoing),
		},
		{
			name: "max size too small",
			options: []DataSetBufferSettingsOption{
				WithMaxParallelOutgoing(0),
			},
			expError: fmt.Errorf("MaxParallelOutgoing has value %d which is less than %d", 0, MinimalMaxParallelOutgoing),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			defCfg := NewDefaultDataSetBufferSettings()
			cfg, err := (&defCfg).WithOptions(tt.options...)
			assert.Nil(t, err)
			err = cfg.Validate()
			assert.Equal(t, tt.expError, err)
		})
	}
}
