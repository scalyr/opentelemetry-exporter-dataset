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

package client

import (
	"fmt"
	"net/http"
	"runtime"
	"testing"

	"github.com/scalyr/dataset-go/pkg/version"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/scalyr/dataset-go/pkg/config"
	"go.uber.org/zap"
)

func TestNewClient(t *testing.T) {
	t.Setenv("SCALYR_SERVER", "test")
	t.Setenv("SCALYR_WRITELOG_TOKEN", "writelog")
	t.Setenv("SCALYR_READLOG_TOKEN", "readlog")
	t.Setenv("SCALYR_WRITECONFIG_TOKEN", "writeconfig")
	t.Setenv("SCALYR_READCONFIG_TOKEN", "readconfig")
	cfg, err := config.New(config.FromEnv())
	assert.Nil(t, err)
	sc4, err := NewClient(cfg, nil, zap.Must(zap.NewDevelopment()), nil, nil)
	require.Nil(t, err)
	assert.Equal(t, sc4.Config.Tokens.ReadLog, "readlog")
	assert.Equal(t, sc4.Config.Tokens.WriteLog, "writelog")
	assert.Equal(t, sc4.Config.Tokens.ReadConfig, "readconfig")
	assert.Equal(t, sc4.Config.Tokens.WriteConfig, "writeconfig")
}

func TestHttpStatusCodes(t *testing.T) {
	tests := []struct {
		statusCode  uint32
		isOk        bool
		isRetryable bool
	}{
		// 200
		{
			statusCode:  http.StatusOK,
			isOk:        true,
			isRetryable: false,
		},
		{
			statusCode:  http.StatusCreated,
			isOk:        false,
			isRetryable: false,
		},
		{
			statusCode:  http.StatusAccepted,
			isOk:        false,
			isRetryable: false,
		},
		// 300
		{
			statusCode:  http.StatusMovedPermanently,
			isOk:        false,
			isRetryable: false,
		},
		// 400
		{
			statusCode:  http.StatusUnauthorized,
			isOk:        false,
			isRetryable: true,
		},
		{
			statusCode:  http.StatusForbidden,
			isOk:        false,
			isRetryable: true,
		},
		{
			statusCode:  http.StatusTooManyRequests,
			isOk:        false,
			isRetryable: true,
		},
		// 500
		{
			statusCode:  http.StatusInternalServerError,
			isOk:        false,
			isRetryable: true,
		},
		{
			statusCode:  http.StatusNotImplemented,
			isOk:        false,
			isRetryable: true,
		},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("HTTP Status code: %d", tt.statusCode)
		t.Run(name, func(*testing.T) {
			assert.Equal(t, tt.isOk, isOkStatus(tt.statusCode), name)
			assert.Equal(t, tt.isRetryable, isRetryableStatus(tt.statusCode), name)
		})
	}
}

func TestAddEventsEndpointUrlWithoutTrailingSlash(t *testing.T) {
	t.Setenv("SCALYR_SERVER", "https://app.scalyr.com")
	cfg, err := config.New(config.FromEnv())
	assert.Nil(t, err)
	sc, err := NewClient(cfg, nil, zap.Must(zap.NewDevelopment()), nil, nil)
	require.Nil(t, err)
	assert.Equal(t, sc.addEventsEndpointUrl, "https://app.scalyr.com/api/addEvents")
}

func TestAddEventsEndpointUrlWithTrailingSlash(t *testing.T) {
	t.Setenv("SCALYR_SERVER", "https://app.scalyr.com/")
	cfg2, err := config.New(config.FromEnv())
	assert.Nil(t, err)
	sc2, err := NewClient(cfg2, nil, zap.Must(zap.NewDevelopment()), nil, nil)
	require.Nil(t, err)
	assert.Equal(t, sc2.addEventsEndpointUrl, "https://app.scalyr.com/api/addEvents")
}

func TestUserAgent(t *testing.T) {
	t.Setenv("SCALYR_SERVER", "https://app.scalyr.com/")
	libraryConsumerUserAgentSuffix := "OtelCollector;0.80.0;traces"
	numCpu := fmt.Sprint(runtime.NumCPU())
	cfg, err := config.New(config.FromEnv())
	assert.Nil(t, err)
	client, err := NewClient(cfg, nil, zap.Must(zap.NewDevelopment()), &libraryConsumerUserAgentSuffix, nil)
	clientId := client.Id.String()
	require.Nil(t, err)
	assert.Equal(t, client.userAgent, "dataset-go;"+version.Version+";"+version.ReleasedDate+";"+clientId+";"+runtime.GOOS+";"+runtime.GOARCH+";"+numCpu+";"+libraryConsumerUserAgentSuffix)
}

func TestUserAgentWithoutCollectorAttrs(t *testing.T) {
	t.Setenv("SCALYR_SERVER", "https://app.scalyr.com/")
	numCpu := fmt.Sprint(runtime.NumCPU())
	cfg, err := config.New(config.FromEnv())
	assert.Nil(t, err)
	client, err := NewClient(cfg, nil, zap.Must(zap.NewDevelopment()), nil, nil)
	clientId := client.Id.String()
	require.Nil(t, err)
	assert.Equal(t, client.userAgent, "dataset-go;"+version.Version+";"+version.ReleasedDate+";"+clientId+";"+runtime.GOOS+";"+runtime.GOARCH+";"+numCpu)
}
