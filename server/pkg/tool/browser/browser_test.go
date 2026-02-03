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
	// Setup Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body><h1>Hello World</h1><p>This is a test.</p></body></html>"))
	}))
	defer server.Close()

	// Use default client to bypass SSRF check for testing against localhost
	p := NewProvider(WithClient(http.DefaultClient))

	// Test Tool Definition
	def := p.ToolDefinition()
	assert.Equal(t, "browse_page", def["name"])

	// Test BrowsePage
	content, err := p.BrowsePage(context.Background(), server.URL)
	assert.NoError(t, err)
	assert.Contains(t, content, "# Hello World")
	assert.Contains(t, content, "This is a test.")
	assert.Contains(t, content, server.URL)

	// Test Error Case
	_, err = p.BrowsePage(context.Background(), "")
	assert.Error(t, err)
}
