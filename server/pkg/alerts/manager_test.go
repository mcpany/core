// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"testing"
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
	// Create some alerts
	for i := 0; i < 5; i++ {
		m.CreateAlert(&Alert{Title: "Test"})
	}
	list := m.ListAlerts()
	if len(list) != 5 {
		t.Errorf("expected 5 alerts, got %d", len(list))
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
