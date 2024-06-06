// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package logs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/open-telemetry/opentelemetry-collector-contrib/cmd/telemetrygen/internal/common"
)

type exporter interface {
	export(plog.Logs) error
}

func newExporter(ctx context.Context, cfg *Config) (exporter, error) {

	// Exporter with HTTP
	if cfg.UseHTTP {
		if cfg.Insecure {
			return &httpClientExporter{
				client: http.DefaultClient,
				cfg:    cfg,
			}, nil
		}
		creds, err := common.GetTLSCredentialsForHTTPExporter(cfg.CaFile, cfg.ClientAuth)
		if err != nil {
			return nil, fmt.Errorf("failed to get TLS credentials: %w", err)
		}
		return &httpClientExporter{
			client: &http.Client{Transport: &http.Transport{TLSClientConfig: creds}},
			cfg:    cfg,
		}, nil
	}

	// Exporter with GRPC
	var err error
	var clientConn *grpc.ClientConn
	if cfg.Insecure {
		clientConn, err = grpc.DialContext(ctx, cfg.Endpoint(), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
	} else {
		creds, err := common.GetTLSCredentialsForGRPCExporter(cfg.CaFile, cfg.ClientAuth)
		if err != nil {
			return nil, fmt.Errorf("failed to get TLS credentials: %w", err)
		}
		clientConn, err = grpc.DialContext(ctx, cfg.Endpoint(), grpc.WithTransportCredentials(creds))
		if err != nil {
			return nil, err
		}
	}
	return &gRPCClientExporter{client: plogotlp.NewGRPCClient(clientConn)}, nil
}

type gRPCClientExporter struct {
	client plogotlp.GRPCClient
}

func (e *gRPCClientExporter) export(logs plog.Logs) error {
	req := plogotlp.NewExportRequestFromLogs(logs)
	if _, err := e.client.Export(context.Background(), req); err != nil {
		return err
	}
	return nil
}

type httpClientExporter struct {
	client *http.Client
	cfg    *Config
}

func (e *httpClientExporter) export(logs plog.Logs) error {
	scheme := "https"
	if e.cfg.Insecure {
		scheme = "http"
	}
	path := e.cfg.HTTPPath
	url := fmt.Sprintf("%s://%s%s", scheme, e.cfg.Endpoint(), path)

	req := plogotlp.NewExportRequestFromLogs(logs)
	body, err := req.MarshalProto()
	if err != nil {
		return fmt.Errorf("failed to marshal logs to protobuf: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create logs HTTP request: %w", err)
	}
	for k, v := range e.cfg.Headers {
		httpReq.Header.Set(k, v)
	}
	httpReq.Header.Set("Content-Type", "application/x-protobuf")
	resp, err := e.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute logs HTTP request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var respData bytes.Buffer
		_, _ = io.Copy(&respData, resp.Body)
		return fmt.Errorf("log request failed with status %s (%s)", resp.Status, respData.String())
	}

	return nil
}
