// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/mcpany/core/server/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestAPIHandler_SecurityValidation(t *testing.T) {
	// Create a memory store
	store := memory.NewStore()
	app := &Application{
		ToolManager: tool.NewManager(nil), // Need ToolManager to avoid panic in ReloadConfig log
	}

	handler := app.createAPIHandler(store)

	// Test case: Invalid URL scheme in http_service
	t.Run("Invalid URL Scheme", func(t *testing.T) {
		svc := &configv1.UpstreamServiceConfig{
			Name: proto.String("malicious-service"),
			ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
				HttpService: &configv1.HttpUpstreamService{
					Address: proto.String("gopher://malicious.com"),
				},
			},
		}
		body, _ := protojson.Marshal(svc)

		req := httptest.NewRequest("POST", "/services", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Assert that the request is rejected with Bad Request (400)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid service configuration")
	})

	// Test case: Absolute path in bundle_path (path traversal/security risk)
	t.Run("Absolute Bundle Path", func(t *testing.T) {
		svc := &configv1.UpstreamServiceConfig{
			Name: proto.String("absolute-path-service"),
			ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
				McpService: &configv1.McpUpstreamService{
					ConnectionType: &configv1.McpUpstreamService_BundleConnection{
						BundleConnection: &configv1.McpBundleConnection{
							BundlePath: proto.String("/etc/passwd"),
						},
					},
				},
			},
		}
		body, _ := protojson.Marshal(svc)

		req := httptest.NewRequest("POST", "/services", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Assert that the request is rejected with Bad Request (400)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid service configuration")
	})
}
