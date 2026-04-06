// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage/memory"
)

func TestUABMetricsEndpoint(t *testing.T) {
	app := &Application{}
	mux := http.NewServeMux()
	store := memory.NewMemoryStorage()
	_ = store.SaveService(context.Background(), &configv1.UpstreamServiceConfig{Name: "s1"})

	app.mountUAB(mux, store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/uab/metrics", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", res.Status)
	}

	var metrics map[string]int
	if err := json.NewDecoder(res.Body).Decode(&metrics); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Based on the in-memory mock registration in api_uab.go
	if metrics["sessions"] != 1 {
		t.Errorf("Expected 1 session, got %d", metrics["sessions"])
	}

	if metrics["transports"] != 2 {
		t.Errorf("Expected 2 transports, got %d", metrics["transports"])
	}
}
