// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestCheckConnectivity_HTTP_Success(t *testing.T) {
	// Start a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("valid-http"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: proto.String(server.URL),
					},
				},
			},
		},
	}

	errors := CheckConnectivity(context.Background(), cfg)
	assert.Empty(t, errors)
}

func TestCheckConnectivity_HTTP_Failure(t *testing.T) {
	// Find a free port but don't listen on it
	// Actually, just picking a random high port is usually safe enough for tests to fail
	address := "http://localhost:59999"

	cfg := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("broken-http"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: proto.String(address),
					},
				},
			},
		},
	}

	errors := CheckConnectivity(context.Background(), cfg)
	assert.NotEmpty(t, errors)
	assert.Equal(t, "broken-http", errors[0].ServiceName)
	assert.Contains(t, errors[0].Err.Error(), "connection refused")
}

func TestCheckConnectivity_Disabled(t *testing.T) {
	// Invalid address but service is disabled
	cfg := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name:    proto.String("disabled-service"),
				Disable: proto.Bool(true),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: proto.String("http://localhost:59999"),
					},
				},
			},
		},
	}

	errors := CheckConnectivity(context.Background(), cfg)
	assert.Empty(t, errors)
}
