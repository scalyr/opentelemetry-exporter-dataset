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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/scalyr/dataset-go/pkg/server_host_config"

	"github.com/scalyr/dataset-go/pkg/version"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/scalyr/dataset-go/pkg/buffer_config"

	"github.com/scalyr/dataset-go/pkg/api/add_events"
	"github.com/scalyr/dataset-go/pkg/buffer"
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

func TestClientBuffer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, "{ \"hello\": \"yo\" }")
		assert.Nil(t, err)
	}))

	token := "token-test"
	sc, err := NewClient(&config.DataSetConfig{
		Endpoint:           ts.URL,
		Tokens:             config.DataSetTokens{WriteLog: token},
		BufferSettings:     buffer_config.NewDefaultDataSetBufferSettings(),
		ServerHostSettings: server_host_config.NewDefaultDataSetServerHostSettings(),
	}, &http.Client{}, zap.Must(zap.NewDevelopment()), nil, nil)
	require.Nil(t, err)

	sessionInfo := add_events.SessionInfo{
		"ServerId":   "serverId",
		"ServerType": "testing",
	}

	event1 := &add_events.Event{
		Thread: "TId",
		Sev:    3,
		Ts:     "1",
		Attrs: map[string]interface{}{
			"message": "test - 1",
			"meh":     1,
		},
	}

	sc.newBufferForEvents("aaa", &sessionInfo)
	buffer1 := sc.getBuffer("aaa")
	added, err := buffer1.AddBundle(&add_events.EventBundle{Event: event1})
	assert.Nil(t, err)
	assert.Equal(t, added, buffer.Added)

	payload1, err := buffer1.Payload()
	assert.Nil(t, err)
	params1 := add_events.AddEventsRequest{}
	err = json.Unmarshal(payload1, &params1)
	assert.Nil(t, err)

	event2 := &add_events.Event{
		Thread: "TId",
		Sev:    3,
		Ts:     "2",
		Attrs: map[string]interface{}{
			"message": "test - 2",
			"meh":     1,
		},
	}

	sc.publishBuffer(buffer1)
	buffer2 := sc.getBuffer("aaa")
	added2, err := buffer2.AddBundle(&add_events.EventBundle{Event: event2})
	assert.Nil(t, err)
	assert.Equal(t, added2, buffer.Added)

	payload2, err := buffer2.Payload()
	assert.Nil(t, err)
	params2 := add_events.AddEventsRequest{}
	err = json.Unmarshal(payload2, &params2)
	assert.Nil(t, err)

	assert.Equal(t, params1.Token, params2.Token)
	assert.Equal(t, params1.Session, params2.Session)
	assert.Equal(t, params1.SessionInfo, params2.SessionInfo)

	assert.NotEqual(t, (params1.Events)[0], (params2.Events)[0])
	assert.Equal(t, (params1.Events)[0].Ts, "1")
	assert.Equal(t, (params2.Events)[0].Ts, "2")
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
