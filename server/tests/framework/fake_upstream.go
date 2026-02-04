// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// NewFakeUpstream creates a new fake upstream server that responds with the provided JSON.
func NewFakeUpstream(t *testing.T, response interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
}

// NewFakeUpstreamHandler creates a new fake upstream server with a custom handler.
func NewFakeUpstreamHandler(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}
