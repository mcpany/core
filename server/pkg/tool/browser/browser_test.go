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

	// 1. Functional Test (Bypass SSRF check to test parsing)
	pFunc := NewProvider(WithClient(http.DefaultClient))

	// Test Tool Definition
	def := pFunc.ToolDefinition()
	assert.Equal(t, "browse_page", def["name"])

	// Test BrowsePage
	content, err := pFunc.BrowsePage(context.Background(), server.URL)
	assert.NoError(t, err)
	assert.Contains(t, content, "# Hello World")
	assert.Contains(t, content, "This is a test.")
	assert.Contains(t, content, server.URL)

	// Test Error Case
	_, err = pFunc.BrowsePage(context.Background(), "")
	assert.Error(t, err)

	// 2. Security Test (Verify SSRF protection)
	// Create provider with default security enabled
	pSec := NewProvider()

	// Attempt to access the mock server (which is on localhost/127.0.0.1)
	// This MUST fail with our security policy
	_, err = pSec.BrowsePage(context.Background(), server.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no safe public IP found")
}
