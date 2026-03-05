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

func TestManager_Webhook(t *testing.T) {
	m := NewManager()
	url := "http://example.com/webhook"
	m.SetWebhookURL(url)

	if got := m.GetWebhookURL(); got != url {
		t.Errorf("expected webhook URL %s, got %s", url, got)
	}
}

func TestManager_GetAlertStats(t *testing.T) {
	m := NewManager()
	stats := m.GetAlertStats()
	if stats == nil {
		t.Error("expected non-nil stats")
	}

	// With the seeded data, we should have 1 active critical, 1 active warning, and at least some total today
	if stats.ActiveCritical != 1 {
		t.Errorf("expected 1 active critical alert, got %d", stats.ActiveCritical)
	}
	if stats.ActiveWarning != 1 {
		t.Errorf("expected 1 active warning alert, got %d", stats.ActiveWarning)
	}
	if stats.TotalToday < 1 {
		t.Errorf("expected >0 total today, got %d", stats.TotalToday)
	}
	if stats.MTTR == "" {
		t.Error("expected non-empty MTTR")
	}
}

func TestManager_WebhookTriggers(t *testing.T) {
	var requestCount int32
	alertChan := make(chan Alert, 1)

	// Create a test HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)

		if r.Method != http.MethodPost {
			t.Errorf("Expected method POST, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var lastAlert Alert
		err := json.NewDecoder(r.Body).Decode(&lastAlert)
		if err != nil {
			t.Errorf("Failed to decode webhook payload: %v", err)
		}

		// Simulate occasionally failing response
		if atomic.LoadInt32(&requestCount) == 2 {
			w.WriteHeader(http.StatusInternalServerError)
			alertChan <- lastAlert
			return
		}

		w.WriteHeader(http.StatusOK)
		alertChan <- lastAlert
	}))
	defer ts.Close()

	m := NewManager()
	m.SetWebhookURL(ts.URL)

	// Test 1: CreateAlert triggers webhook
	alert := &Alert{
		Title:    "Test Webhook Alert",
		Severity: SeverityCritical,
		Status:   StatusActive,
	}
	created := m.CreateAlert(alert)

	var receivedAlert Alert
	select {
	case receivedAlert = <-alertChan:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for webhook request")
	}

	if atomic.LoadInt32(&requestCount) != 1 {
		t.Errorf("Expected webhook request count to be 1, got %d", atomic.LoadInt32(&requestCount))
	}
	if receivedAlert.ID != created.ID {
		t.Errorf("Expected webhook to send alert ID %s, got %s", created.ID, receivedAlert.ID)
	}

	// Test 2: UpdateAlert triggers webhook (even if it returns 500, shouldn't panic/fail test)
	m.UpdateAlert(created.ID, &Alert{Status: StatusResolved})

	select {
	case receivedAlert = <-alertChan:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for webhook request")
	}

	if atomic.LoadInt32(&requestCount) != 2 {
		t.Errorf("Expected webhook request count to be 2, got %d", atomic.LoadInt32(&requestCount))
	}
	if receivedAlert.Status != StatusResolved {
		t.Errorf("Expected webhook to send updated status Resolved, got %s", receivedAlert.Status)
	}

	// Test 3: Invalid Webhook URL doesn't crash application
	m.SetWebhookURL("http://invalid-url-that-does-not-exist.local")
	m.CreateAlert(&Alert{Title: "Should fail gracefully"})

	// Wait a bit just to be sure there's no panic during execution
	time.Sleep(100 * time.Millisecond)

	// If we reach here, no panic occurred.

	// Test 4: Unexisting alert on Update
	nilRes := m.UpdateAlert("non-existent-id", &Alert{Status: StatusResolved})
	if nilRes != nil {
		t.Errorf("Expected nil when updating non-existent alert, got %+v", nilRes)
	}
}
