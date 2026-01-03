// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/pkg/storage/memory"
	"github.com/mcpany/core/pkg/tool"
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
		svc := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("malicious-service"),
			HttpService: configv1.HttpUpstreamService_builder{
				Address: proto.String("gopher://malicious.com"),
			}.Build(),
		}.Build()
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
		svc := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("absolute-path-service"),
			McpService: configv1.McpUpstreamService_builder{
				BundleConnection: configv1.McpBundleConnection_builder{
					BundlePath: proto.String("/etc/passwd"),
				}.Build(),
			}.Build(),
		}.Build()
		body, _ := protojson.Marshal(svc)

		req := httptest.NewRequest("POST", "/services", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Assert that the request is rejected with Bad Request (400)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid service configuration")
	})
}
