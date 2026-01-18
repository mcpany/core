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

	// Test case: Filesystem service creation blocked via API
	t.Run("Block Filesystem Service", func(t *testing.T) {
		svc := &configv1.UpstreamServiceConfig{
			Name: proto.String("unsafe-fs"),
			ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
				FilesystemService: &configv1.FilesystemUpstreamService{
					RootPaths: map[string]string{
						"/": "/",
					},
					FilesystemType: &configv1.FilesystemUpstreamService_Os{
						Os: &configv1.OsFs{},
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
		// We updated the error message in the implementation
		assert.Contains(t, w.Body.String(), "unsafe services")
		assert.Contains(t, w.Body.String(), "filesystem")
	})

	// Test case: SQL service creation blocked via API
	t.Run("Block SQL Service", func(t *testing.T) {
		svc := &configv1.UpstreamServiceConfig{
			Name: proto.String("unsafe-sql"),
			ServiceConfig: &configv1.UpstreamServiceConfig_SqlService{
				SqlService: &configv1.SqlUpstreamService{
					Driver: proto.String("sqlite"),
					Dsn:    proto.String("file::memory:"),
				},
			},
		}
		body, _ := protojson.Marshal(svc)

		req := httptest.NewRequest("POST", "/services", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Assert that the request is rejected with Bad Request (400)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "unsafe services")
		assert.Contains(t, w.Body.String(), "sql")
	})
}
