// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestManager_CRUD(t *testing.T) {
	m := NewManager()

	w := &WebhookConfig{
		URL:    "http://example.com",
		Events: []string{"test"},
		Active: true,
	}

	m.AddWebhook(w)

	list := m.ListWebhooks()
	if len(list) != 1 {
		t.Fatalf("expected 1 webhook, got %d", len(list))
	}
	if list[0].URL != "http://example.com" {
		t.Errorf("expected URL http://example.com, got %s", list[0].URL)
	}

	id := list[0].ID
	got, ok := m.GetWebhook(id)
	if !ok {
		t.Fatalf("expected webhook to exist")
	}
	if got.ID != id {
		t.Errorf("expected ID %s, got %s", id, got.ID)
	}

	m.DeleteWebhook(id)
	list = m.ListWebhooks()
	if len(list) != 0 {
		t.Errorf("expected 0 webhooks, got %d", len(list))
	}
}

func TestManager_TestWebhook(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	m := NewManager()
	w := &WebhookConfig{
		URL:    ts.URL,
		Events: []string{"test"},
		Active: true,
	}
	m.AddWebhook(w)
	id := m.ListWebhooks()[0].ID

	err := m.TestWebhook(context.Background(), id)
	if err != nil {
		t.Fatalf("TestWebhook failed: %v", err)
	}

	updated, _ := m.GetWebhook(id)
	if updated.Status != "success" {
		t.Errorf("expected status success, got %s", updated.Status)
	}
}

func TestManager_TestWebhook_Failure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	m := NewManager()
	w := &WebhookConfig{
		URL:    ts.URL,
		Events: []string{"test"},
		Active: true,
	}
	m.AddWebhook(w)
	id := m.ListWebhooks()[0].ID

	err := m.TestWebhook(context.Background(), id)
	if err == nil {
		t.Fatalf("TestWebhook expected error, got nil")
	}

	updated, _ := m.GetWebhook(id)
	if updated.Status != "failure" {
		t.Errorf("expected status failure, got %s", updated.Status)
	}
}

func TestManager_TestWebhook_NotFound(t *testing.T) {
	m := NewManager()
	err := m.TestWebhook(context.Background(), "nonexistent")
	if err == nil {
		t.Fatalf("TestWebhook expected error for nonexistent webhook, got nil")
	}
	if err.Error() != "webhook not found" {
		t.Errorf("expected error 'webhook not found', got %v", err)
	}
}

func TestManager_TestWebhook_InvalidURL(t *testing.T) {
	m := NewManager()
	w := &WebhookConfig{
		URL:    "http://192.168.0.%31/", // Invalid URL that fails http.NewRequestWithContext
		Events: []string{"test"},
		Active: true,
	}
	m.AddWebhook(w)
	id := m.ListWebhooks()[0].ID

	// Override URL to something that breaks url.Parse inside NewRequestWithContext
	w.URL = string([]byte{0x7f})

	err := m.TestWebhook(context.Background(), id)
	if err == nil {
		t.Fatalf("TestWebhook expected error for invalid URL, got nil")
	}
}

func TestManager_TestWebhook_ClientError(t *testing.T) {
	m := NewManager()
	w := &WebhookConfig{
		URL:    "http://127.0.0.1:0", // Invalid port that dials will fail
		Events: []string{"test"},
		Active: true,
	}
	m.AddWebhook(w)
	id := m.ListWebhooks()[0].ID

	err := m.TestWebhook(context.Background(), id)
	if err == nil {
		t.Fatalf("TestWebhook expected error for client error, got nil")
	}

	updated, _ := m.GetWebhook(id)
	if updated.Status != "failure" {
		t.Errorf("expected status failure, got %s", updated.Status)
	}
}
