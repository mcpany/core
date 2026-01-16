// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package connectivity

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func strPtr(s string) *string {
	return &s
}

func TestCheckHTTP(t *testing.T) {
	// Start a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create a config pointing to the test server
	cfg := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: strPtr("test-http"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: strPtr(server.URL),
					},
				},
			},
		},
	}

	// Run check
	results := Check(context.Background(), cfg)

	// Assertions
	assert.Len(t, results, 1)
	assert.Equal(t, "test-http", results[0].ServiceName)
	assert.True(t, results[0].Status)
	assert.Equal(t, "HTTP", results[0].Type)
}

func TestCheckHTTP_Fail(t *testing.T) {
	// Create a config pointing to a non-existent server
	// Use a reserved port or localhost with random closed port
	cfg := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: strPtr("test-http-fail"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: strPtr("http://localhost:59999"), // unlikely to be open
					},
				},
			},
		},
	}

	// Run check
	results := Check(context.Background(), cfg)

	// Assertions
	assert.Len(t, results, 1)
	assert.False(t, results[0].Status)
	assert.Error(t, results[0].Error)
}

func TestCheckTCP(t *testing.T) {
	// Start a TCP listener
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	// Create config for gRPC
	cfg := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: strPtr("test-grpc"),
				ServiceConfig: &configv1.UpstreamServiceConfig_GrpcService{
					GrpcService: &configv1.GrpcUpstreamService{
						Address: strPtr(ln.Addr().String()),
					},
				},
			},
		},
	}

	results := Check(context.Background(), cfg)
	assert.Len(t, results, 1)
	assert.True(t, results[0].Status)
}
