// Copyright 2025 Author(s) of MCP Any
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

func TestCheckConnectivity(t *testing.T) {
	// Start a test server that returns 200 OK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Case 1: Reachable Service
	t.Run("ReachableService", func(t *testing.T) {
		cfg := &configv1.McpAnyServerConfig{
			UpstreamServices: []*configv1.UpstreamServiceConfig{
				{
					Name: proto.String("reachable"),
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
	})

	// Case 2: Unreachable Service (Bad Port)
	t.Run("UnreachableService", func(t *testing.T) {
		// Use a port that is likely closed or reserved
		cfg := &configv1.McpAnyServerConfig{
			UpstreamServices: []*configv1.UpstreamServiceConfig{
				{
					Name: proto.String("unreachable"),
					ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
						HttpService: &configv1.HttpUpstreamService{
							Address: proto.String("http://localhost:12345"), // Assuming this is not bound
						},
					},
				},
			},
		}

		errors := CheckConnectivity(context.Background(), cfg)
		assert.NotEmpty(t, errors)
		// Error message depends on OS/network, but usually contains "connection refused" or "dial tcp"
		assert.Contains(t, errors[0].Error(), "connection refused")
	})

	// Case 3: Mixed Services
	t.Run("MixedServices", func(t *testing.T) {
		cfg := &configv1.McpAnyServerConfig{
			UpstreamServices: []*configv1.UpstreamServiceConfig{
				{
					Name: proto.String("reachable"),
					ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
						HttpService: &configv1.HttpUpstreamService{
							Address: proto.String(server.URL),
						},
					},
				},
				{
					Name: proto.String("unreachable"),
					ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
						HttpService: &configv1.HttpUpstreamService{
							Address: proto.String("http://localhost:54321"),
						},
					},
				},
			},
		}

		errors := CheckConnectivity(context.Background(), cfg)
		assert.Len(t, errors, 1)
		assert.Equal(t, "unreachable", errors[0].ServiceName)
	})
}
