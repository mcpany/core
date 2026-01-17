// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package doctor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func TestRunChecks_Http(t *testing.T) {
	// Start a mock HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: strPtr("valid-http"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: strPtr(ts.URL),
					},
				},
			},
			{
				Name: strPtr("invalid-http"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: strPtr("http://localhost:12345/nonexistent"),
					},
				},
			},
			{
				Name:    strPtr("disabled-service"),
				Disable: boolPtr(true),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: strPtr(ts.URL),
					},
				},
			},
		},
	}

	results := RunChecks(context.Background(), config)

	assert.Len(t, results, 3)

	// Check valid service
	assert.Equal(t, "valid-http", results[0].ServiceName)
	assert.Equal(t, StatusOk, results[0].Status)

	// Check invalid service
	assert.Equal(t, "invalid-http", results[1].ServiceName)
	assert.Equal(t, StatusError, results[1].Status)

	// Check disabled service
	assert.Equal(t, "disabled-service", results[2].ServiceName)
	assert.Equal(t, StatusSkipped, results[2].Status)
}

func TestRunChecks_Grpc(t *testing.T) {
	// We can cheat and use the HTTP listener address for TCP check
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()

	// Extract host:port from ts.URL (http://127.0.0.1:xxxxx)
	addr := ts.Listener.Addr().String()

	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: strPtr("valid-grpc"),
				ServiceConfig: &configv1.UpstreamServiceConfig_GrpcService{
					GrpcService: &configv1.GrpcUpstreamService{
						Address: strPtr(addr),
					},
				},
			},
			{
				Name: strPtr("invalid-grpc"),
				ServiceConfig: &configv1.UpstreamServiceConfig_GrpcService{
					GrpcService: &configv1.GrpcUpstreamService{
						Address: strPtr("localhost:1"), // Unlikely port
					},
				},
			},
		},
	}

	results := RunChecks(context.Background(), config)

	assert.Len(t, results, 2)
	assert.Equal(t, StatusOk, results[0].Status)
	assert.Equal(t, StatusError, results[1].Status)
}
