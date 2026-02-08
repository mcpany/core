package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/alerts"
	"github.com/stretchr/testify/assert"
)

func TestHandleAlerts(t *testing.T) {
	app := NewApplication()
	// Use default seeded data from NewManager
	app.AlertsManager = alerts.NewManager()

	t.Run("ListAlerts", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/alerts", nil)
		w := httptest.NewRecorder()

		app.handleAlerts()(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var list []*alerts.Alert
		err := json.Unmarshal(w.Body.Bytes(), &list)
		assert.NoError(t, err)
		// Seed data has 5 alerts
		assert.GreaterOrEqual(t, len(list), 5)
	})

	t.Run("CreateAlert", func(t *testing.T) {
		newAlert := &alerts.Alert{
			Title:    "Test Alert",
			Message:  "Something happened",
			Severity: alerts.SeverityInfo,
			Service:  "test-service",
			Source:   "test",
		}
		body, _ := json.Marshal(newAlert)
		req := httptest.NewRequest(http.MethodPost, "/alerts", bytes.NewReader(body))
		w := httptest.NewRecorder()

		app.handleAlerts()(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var created alerts.Alert
		err := json.Unmarshal(w.Body.Bytes(), &created)
		assert.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.Equal(t, "Test Alert", created.Title)
	})

	t.Run("MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/alerts", nil)
		w := httptest.NewRecorder()
		app.handleAlerts()(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestHandleAlertDetail(t *testing.T) {
	app := NewApplication()
	// Ensure we have a known alert
	alert := &alerts.Alert{
		ID:       "test-id",
		Title:    "Test",
		Severity: alerts.SeverityWarning,
		Status:   alerts.StatusActive,
	}
	app.AlertsManager.CreateAlert(alert)

	t.Run("GetAlert", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/alerts/test-id", nil)
		w := httptest.NewRecorder()

		app.handleAlertDetail()(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var got alerts.Alert
		err := json.Unmarshal(w.Body.Bytes(), &got)
		assert.NoError(t, err)
		assert.Equal(t, "test-id", got.ID)
	})

	t.Run("UpdateAlert", func(t *testing.T) {
		updates := &alerts.Alert{
			Status: alerts.StatusResolved,
		}
		body, _ := json.Marshal(updates)
		req := httptest.NewRequest(http.MethodPatch, "/alerts/test-id", bytes.NewReader(body))
		w := httptest.NewRecorder()

		app.handleAlertDetail()(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var updated alerts.Alert
		err := json.Unmarshal(w.Body.Bytes(), &updated)
		assert.NoError(t, err)
		assert.Equal(t, alerts.StatusResolved, updated.Status)
	})

	t.Run("NotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/alerts/non-existent", nil)
		w := httptest.NewRecorder()
		app.handleAlertDetail()(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("MissingID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/alerts/", nil)
		w := httptest.NewRecorder()
		app.handleAlertDetail()(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandleAlertRules(t *testing.T) {
	app := NewApplication()

	t.Run("ListRules", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/alerts/rules", nil)
		w := httptest.NewRecorder()

		app.handleAlertRules()(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var list []*alerts.AlertRule
		err := json.Unmarshal(w.Body.Bytes(), &list)
		assert.NoError(t, err)
		// Seed data has 2 rules
		assert.GreaterOrEqual(t, len(list), 2)
	})

	t.Run("CreateRule", func(t *testing.T) {
		newRule := &alerts.AlertRule{
			Name:      "New Rule",
			Metric:    "cpu",
			Operator:  ">",
			Threshold: 80,
		}
		body, _ := json.Marshal(newRule)
		req := httptest.NewRequest(http.MethodPost, "/alerts/rules", bytes.NewReader(body))
		w := httptest.NewRecorder()

		app.handleAlertRules()(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var created alerts.AlertRule
		err := json.Unmarshal(w.Body.Bytes(), &created)
		assert.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.Equal(t, "New Rule", created.Name)
	})
}

func TestHandleAlertRuleDetail(t *testing.T) {
	app := NewApplication()
	rule := &alerts.AlertRule{
		ID:   "rule-123",
		Name: "Test Rule",
	}
	app.AlertsManager.CreateRule(rule)

	t.Run("GetRule", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/alerts/rules/rule-123", nil)
		w := httptest.NewRecorder()

		app.handleAlertRuleDetail()(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var got alerts.AlertRule
		err := json.Unmarshal(w.Body.Bytes(), &got)
		assert.NoError(t, err)
		assert.Equal(t, "rule-123", got.ID)
	})

	t.Run("UpdateRule", func(t *testing.T) {
		updates := &alerts.AlertRule{
			Name:      "Updated Rule",
			LastUpdated: time.Now(),
		}
		body, _ := json.Marshal(updates)
		req := httptest.NewRequest(http.MethodPut, "/alerts/rules/rule-123", bytes.NewReader(body))
		w := httptest.NewRecorder()

		app.handleAlertRuleDetail()(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var updated alerts.AlertRule
		err := json.Unmarshal(w.Body.Bytes(), &updated)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Rule", updated.Name)
	})

	t.Run("DeleteRule", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/alerts/rules/rule-123", nil)
		w := httptest.NewRecorder()

		app.handleAlertRuleDetail()(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Nil(t, app.AlertsManager.GetRule("rule-123"))
	})

	t.Run("NotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/alerts/rules/non-existent", nil)
		w := httptest.NewRecorder()
		app.handleAlertRuleDetail()(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestHandleAlertWebhook(t *testing.T) {
	app := NewApplication()

	t.Run("GetWebhookURL_Default", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/alerts/webhook", nil)
		w := httptest.NewRecorder()

		app.handleAlertWebhook()(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "", resp["url"])
	})

	t.Run("SetWebhookURL", func(t *testing.T) {
		payload := map[string]string{"url": "http://example.com/webhook"}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/alerts/webhook", bytes.NewReader(body))
		w := httptest.NewRecorder()

		app.handleAlertWebhook()(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "http://example.com/webhook", resp["url"])

		// Verify it persisted
		assert.Equal(t, "http://example.com/webhook", app.AlertsManager.GetWebhookURL())
	})
}
