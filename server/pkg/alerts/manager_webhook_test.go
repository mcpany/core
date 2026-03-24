package alerts

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestManager_Webhook_CreateAlert(t *testing.T) {
	// Setup test server to mock webhook endpoint
	webhookCalled := make(chan bool, 1)
	var receivedAlert Alert
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		err := json.Unmarshal(body, &receivedAlert)
		if err != nil {
			t.Errorf("Failed to unmarshal webhook body: %v", err)
		}
		webhookCalled <- true
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	m := NewManager()
	m.SetWebhookURL(ts.URL)

	alert := &Alert{
		Title:    "Webhook Test Alert",
		Severity: SeverityCritical,
		Status:   StatusActive,
	}

	created := m.CreateAlert(alert)

	select {
	case <-webhookCalled:
		// Success
		if receivedAlert.ID != created.ID {
			t.Errorf("Webhook received alert with ID %s, expected %s", receivedAlert.ID, created.ID)
		}
		if receivedAlert.Title != "Webhook Test Alert" {
			t.Errorf("Webhook received alert with Title %s, expected 'Webhook Test Alert'", receivedAlert.Title)
		}
	case <-time.After(2 * time.Second):
		t.Error("Webhook was not called within timeout")
	}
}

func TestManager_Webhook_UpdateAlert(t *testing.T) {
	m := NewManager()

	// Create an alert first without a webhook
	alert := &Alert{
		Title:    "Update Webhook Test Alert",
		Severity: SeverityWarning,
		Status:   StatusActive,
	}
	created := m.CreateAlert(alert)

	// Now set the webhook and update the alert
	webhookCalled := make(chan bool, 1)
	var receivedAlert Alert
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedAlert)
		webhookCalled <- true
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	m.SetWebhookURL(ts.URL)

	m.UpdateAlert(created.ID, &Alert{Status: StatusResolved})

	select {
	case <-webhookCalled:
		// Success
		if receivedAlert.ID != created.ID {
			t.Errorf("Webhook received alert with ID %s, expected %s", receivedAlert.ID, created.ID)
		}
		if receivedAlert.Status != StatusResolved {
			t.Errorf("Webhook received alert with Status %s, expected %s", receivedAlert.Status, StatusResolved)
		}
	case <-time.After(2 * time.Second):
		t.Error("Webhook was not called within timeout for update")
	}
}

func TestManager_Webhook_UpdateAlert_NotFound(t *testing.T) {
	m := NewManager()

	// Set the webhook
	webhookCalled := make(chan bool, 1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		webhookCalled <- true
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	m.SetWebhookURL(ts.URL)

	res := m.UpdateAlert("nonexistent", &Alert{Status: StatusResolved})
	if res != nil {
		t.Errorf("Expected nil for nonexistent alert, got %v", res)
	}

	select {
	case <-webhookCalled:
		t.Error("Webhook should not be called for non-existent alert update")
	case <-time.After(500 * time.Millisecond):
		// Success, webhook was not called
	}
}
