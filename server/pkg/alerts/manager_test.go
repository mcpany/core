// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestManager_CreateAndGet(t *testing.T) {
	m := NewManager()
	alert := &Alert{
		Title:    "Test Alert",
		Severity: SeverityInfo,
		Status:   StatusActive,
	}
	created := m.CreateAlert(alert)
	if created.ID == "" {
		t.Error("expected ID to be generated")
	}

	got := m.GetAlert(created.ID)
	if got == nil {
		t.Error("expected to get alert")
	}
	if got.Title != "Test Alert" {
		t.Errorf("expected title 'Test Alert', got '%s'", got.Title)
	}
}

func TestManager_List(t *testing.T) {
	m := NewManager()
	// Should have seeded data (5 items)
	list := m.ListAlerts()
	if len(list) < 5 {
		t.Errorf("expected at least 5 seeded alerts, got %d", len(list))
	}
}

func TestManager_Update(t *testing.T) {
	m := NewManager()
	alert := &Alert{Title: "Test", Status: StatusActive}
	created := m.CreateAlert(alert)

	updated := m.UpdateAlert(created.ID, &Alert{Status: StatusResolved})
	if updated.Status != StatusResolved {
		t.Errorf("expected status Resolved, got %s", updated.Status)
	}

	got := m.GetAlert(created.ID)
	if got.Status != StatusResolved {
		t.Errorf("expected persisted status Resolved, got %s", got.Status)
	}
}

func TestManager_Webhooks(t *testing.T) {
	m := NewManager()

	wh1 := &Webhook{URL: "http://example.com/1", Events: []string{"all"}, Active: true}
	created1 := m.CreateWebhook(wh1)

	wh2 := &Webhook{URL: "http://example.com/2", Events: []string{"alerts"}, Active: true}
	created2 := m.CreateWebhook(wh2)

	list := m.ListWebhooks()
	if len(list) != 2 {
		t.Errorf("expected 2 webhooks, got %d", len(list))
	}

	got1 := m.GetWebhook(created1.ID)
	if got1 == nil || got1.URL != wh1.URL {
		t.Error("failed to get webhook 1")
	}

	updated := m.UpdateWebhook(created1.ID, &Webhook{URL: "http://updated.com", Events: []string{"all"}, Active: false})
	if updated.URL != "http://updated.com" || updated.Active != false {
		t.Error("failed to update webhook")
	}

	m.DeleteWebhook(created2.ID)
	if m.GetWebhook(created2.ID) != nil {
		t.Error("failed to delete webhook")
	}
}

func TestManager_Dispatch(t *testing.T) {
	var callCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		var a Alert
		if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
			t.Error(err)
		}
		if a.Title != "Dispatch Test" {
			t.Errorf("expected title 'Dispatch Test', got %s", a.Title)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	m := NewManager()
	m.CreateWebhook(&Webhook{URL: server.URL, Events: []string{"all"}, Active: true})
	m.CreateWebhook(&Webhook{URL: server.URL + "/ignore", Events: []string{"none"}, Active: true})      // Should be ignored
	m.CreateWebhook(&Webhook{URL: server.URL + "/inactive", Events: []string{"all"}, Active: false}) // Should be ignored

	m.CreateAlert(&Alert{Title: "Dispatch Test", Status: StatusActive})

	// Wait a bit for async dispatch
	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("expected 1 webhook call, got %d", callCount)
	}
}
