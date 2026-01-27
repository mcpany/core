// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/pkg/resource"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

// mockBlobResource for testing binary downloads
type mockBlobResource struct {
	uri  string
	blob []byte
}

func (m *mockBlobResource) Resource() *mcp.Resource {
	return &mcp.Resource{URI: m.uri, Name: "blob.bin"}
}
func (m *mockBlobResource) Service() string { return "mock" }
func (m *mockBlobResource) Read(ctx context.Context) (*mcp.ReadResourceResult, error) {
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      m.uri,
				Blob:     m.blob,
				MIMEType: "application/octet-stream",
			},
		},
	}, nil
}
func (m *mockBlobResource) Subscribe(ctx context.Context) error { return nil }

func TestHandleResourceDownload(t *testing.T) {
	app := NewApplication()
	app.ResourceManager = resource.NewManager()

	// Text Resource (using mockResource from api_test.go if available, or define here to be safe)
	// Since api_test.go is in the same package `app`, we can use it.
	textRes := &mockResource{uri: "mock://text-resource", content: "hello world"}
	app.ResourceManager.AddResource(textRes)

	// Binary Resource
	blobContent := "blob content"
	// Blob in ResourceContents is []byte (raw bytes)
	blobRes := &mockBlobResource{uri: "mock://blob-resource", blob: []byte(blobContent)}
	app.ResourceManager.AddResource(blobRes)

	handler := app.handleResourceDownload()

	t.Run("DownloadText", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/resources/download?uri=mock://text-resource", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "hello world", w.Body.String())
		assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment")
		// mockResource returns empty name in Resource(), so it falls back to Base(uri) -> "text-resource"
		assert.Contains(t, w.Header().Get("Content-Disposition"), "text-resource")
	})

	t.Run("DownloadBlob", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/resources/download?uri=mock://blob-resource", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, blobContent, w.Body.String())
		assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment")
		assert.Contains(t, w.Header().Get("Content-Disposition"), "blob.bin")
		assert.Equal(t, "application/octet-stream", w.Header().Get("Content-Type"))
	})

	t.Run("NotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/resources/download?uri=mock://nonexistent", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("MissingURI", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/resources/download", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
