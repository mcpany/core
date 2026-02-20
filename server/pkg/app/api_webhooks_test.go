// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/pkg/webhooks"
)

func TestHandleWebhooks(t *testing.T) {
	app := NewApplication()
	ts := httptest.NewServer(app.handleWebhooks())
	defer ts.Close()

	// POST
	cfg := webhooks.WebhookConfig{
		URL:    "http://example.com",
		Events: []string{"test"},
	}
	body, _ := json.Marshal(cfg)
	resp, err := http.Post(ts.URL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201 Created, got %d", resp.StatusCode)
	}

	// GET
	resp, err = http.Get(ts.URL)
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	var list []*webhooks.WebhookConfig
	json.NewDecoder(resp.Body).Decode(&list)
	if len(list) != 1 {
		t.Errorf("expected 1 webhook, got %d", len(list))
	}
}

func TestHandleWebhookDetail(t *testing.T) {
	app := NewApplication()
	// Pre-populate
	w := &webhooks.WebhookConfig{
		ID:     "wh-123",
		URL:    "http://example.com",
		Events: []string{"test"},
	}
	app.WebhooksManager.AddWebhook(w)

	// Mock server for detail handler. Since it strips prefix, we need to be careful with URL construction
	// or just test the handler function directly.
	// But `handleWebhookDetail` expects URL path to contain ID.
	handler := app.handleWebhookDetail()

	// Test GET
	req := httptest.NewRequest("GET", "/webhooks/wh-123", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET expected 200, got %d", rr.Code)
	}
	var got webhooks.WebhookConfig
	json.NewDecoder(rr.Body).Decode(&got)
	if got.ID != "wh-123" {
		t.Errorf("GET expected ID wh-123, got %s", got.ID)
	}

	// Test DELETE
	req = httptest.NewRequest("DELETE", "/webhooks/wh-123", nil)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("DELETE expected 204, got %d", rr.Code)
	}
	if _, ok := app.WebhooksManager.GetWebhook("wh-123"); ok {
		t.Errorf("DELETE expected webhook to be gone")
	}
}
