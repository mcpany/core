package alerts

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestManager_CreateAlert_Webhook(t *testing.T) {
	called := make(chan bool, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var a Alert
		if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
			t.Errorf("expected valid JSON: %v", err)
		}
		if a.Title != "Test Webhook Alert" {
			t.Errorf("expected title 'Test Webhook Alert', got '%s'", a.Title)
		}
		called <- true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	m := NewManager()
	m.SetWebhookURL(server.URL)

	alert := &Alert{
		Title:    "Test Webhook Alert",
		Severity: SeverityCritical,
		Status:   StatusActive,
	}
	m.CreateAlert(alert)

	select {
	case <-called:
	case <-time.After(2 * time.Second):
		t.Error("webhook was not called")
	}
}

func TestManager_UpdateAlert_Webhook(t *testing.T) {
	called := make(chan bool, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var a Alert
		if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
			t.Errorf("expected valid JSON: %v", err)
		}
		if a.Status != StatusResolved {
			t.Errorf("expected status 'Resolved', got '%s'", a.Status)
		}
		called <- true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	m := NewManager()
	alert := &Alert{
		Title:    "Test Update Alert",
		Severity: SeverityCritical,
		Status:   StatusActive,
	}
	created := m.CreateAlert(alert)

	m.SetWebhookURL(server.URL)
	m.UpdateAlert(created.ID, &Alert{Status: StatusResolved})

	select {
	case <-called:
	case <-time.After(2 * time.Second):
		t.Error("webhook was not called")
	}
}

func TestManager_UpdateAlert_NotFound(t *testing.T) {
	m := NewManager()
	updated := m.UpdateAlert("nonexistent", &Alert{Status: StatusResolved})
	if updated != nil {
		t.Error("expected nil when updating nonexistent alert")
	}
}

func TestManager_RuleOperations(t *testing.T) {
	m := NewManager()
	err := m.DeleteRule("1")
	if err != nil {
		t.Errorf("unexpected error deleting rule: %v", err)
	}
	updated := m.UpdateRule("nonexistent", &AlertRule{})
	if updated != nil {
		t.Error("expected nil when updating nonexistent rule")
	}
	rule := &AlertRule{
		Name:     "Test Rule",
		Metric:   "test_metric",
		Operator: ">",
		Threshold: 10,
	}
	created := m.CreateRule(rule)
	if created.ID == "" {
		t.Error("expected ID to be generated for rule")
	}
	created.Name = "Updated Rule"
	updated = m.UpdateRule(created.ID, created)
	if updated == nil {
		t.Error("expected rule to be updated")
	}
	if updated.Name != "Updated Rule" {
		t.Errorf("expected name 'Updated Rule', got '%s'", updated.Name)
	}
}

func TestManager_Webhook_Errors(t *testing.T) {
	m := NewManager()
	m.SetWebhookURL("://invalid-url")
	alert := &Alert{
		Title:    "Test Webhook Alert",
		Severity: SeverityCritical,
		Status:   StatusActive,
	}
	m.CreateAlert(alert)
	time.Sleep(100 * time.Millisecond)

	created := m.CreateAlert(alert)
	m.UpdateAlert(created.ID, &Alert{Status: StatusResolved})
	time.Sleep(100 * time.Millisecond)

	errServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer errServer.Close()
	m.SetWebhookURL(errServer.URL)
	m.CreateAlert(alert)
	m.UpdateAlert(created.ID, &Alert{Status: StatusResolved})
	time.Sleep(100 * time.Millisecond)
}

func TestManager_Webhook_RequestError(t *testing.T) {
	m := NewManager()
	m.SetWebhookURL("http://example.com/webhook\x00")
	alert := &Alert{
		Title:    "Test Webhook Alert",
		Severity: SeverityCritical,
		Status:   StatusActive,
	}
	created := m.CreateAlert(alert)
	time.Sleep(100 * time.Millisecond)
	m.UpdateAlert(created.ID, &Alert{Status: StatusResolved})
	time.Sleep(100 * time.Millisecond)
}
