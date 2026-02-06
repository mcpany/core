// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package browser

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBrowserProvider(t *testing.T) {
	// Create a mock HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<html><body><h1>Hello World</h1><p>This is a test.</p></body></html>`))
	}))
	defer ts.Close()

	// Use standard client for tests to allow access to local test server
	p := NewProvider(WithClient(&http.Client{}))

	// Test ToolDefinition
	def := p.ToolDefinition()
	assert.Equal(t, "browse_page", def["name"])

	// Test BrowsePage with mock server
	content, err := p.BrowsePage(context.Background(), ts.URL)
	assert.NoError(t, err)
	// md converter converts <h1> to # and <p> to text
	assert.Contains(t, content, "# Hello World")
	assert.Contains(t, content, "This is a test.")

	// Test Empty URL
	_, err = p.BrowsePage(context.Background(), "")
	assert.Error(t, err)

	// Test 404
	ts404 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts404.Close()
	_, err = p.BrowsePage(context.Background(), ts404.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}
