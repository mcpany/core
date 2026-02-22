// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// NewMockUpstreamServer creates a new mock server with a mux.
// Handlers is a map of path to handler function.
// The caller is responsible for closing the server.
func NewMockUpstreamServer(t *testing.T, handlers map[string]http.HandlerFunc) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	for path, handler := range handlers {
		mux.HandleFunc(path, handler)
	}
	return httptest.NewServer(mux)
}
