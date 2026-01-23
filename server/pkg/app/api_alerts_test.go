// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/pkg/alerts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAlertsTestApp() *Application {
	return &Application{
		AlertsManager: alerts.NewManager(),
	}
}

func TestHandleAlerts(t *testing.T) {
	app := setupAlertsTestApp()

	t.Run("Get Alerts", func(t *testing.T) {
		// Seed some alerts
		app.AlertsManager.CreateAlert(&alerts.Alert{
			Title: "Test Alert 1",
		})
		app.AlertsManager.CreateAlert(&alerts.Alert{
			Title: "Test Alert 2",
		})

		req := httptest.NewRequest(http.MethodGet, "/alerts", nil)
		w := httptest.NewRecorder()

		app.handleAlerts().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var list []*alerts.Alert
		err := json.Unmarshal(w.Body.Bytes(), &list)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 2) // NewManager seeds data
	})

	t.Run("Create Alert", func(t *testing.T) {
		newAlert := &alerts.Alert{
			Title:    "New Alert",
			Message:  "Something happened",
			Severity: alerts.SeverityCritical,
			Status:   alerts.StatusActive,
		}
		body, _ := json.Marshal(newAlert)
		req := httptest.NewRequest(http.MethodPost, "/alerts", bytes.NewReader(body))
		w := httptest.NewRecorder()

		app.handleAlerts().ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var created alerts.Alert
		err := json.Unmarshal(w.Body.Bytes(), &created)
		require.NoError(t, err)
		assert.Equal(t, newAlert.Title, created.Title)
		assert.NotEmpty(t, created.ID)
		assert.False(t, created.Timestamp.IsZero())
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/alerts", nil)
		w := httptest.NewRecorder()

		app.handleAlerts().ServeHTTP(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("Bad Request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/alerts", bytes.NewReader([]byte("invalid json")))
		w := httptest.NewRecorder()

		app.handleAlerts().ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandleAlertDetail(t *testing.T) {
	app := setupAlertsTestApp()

	// Create a test alert
	alert := app.AlertsManager.CreateAlert(&alerts.Alert{
		Title:    "Detail Test Alert",
		Severity: alerts.SeverityInfo,
		Status:   alerts.StatusActive,
	})

	t.Run("Get Alert Detail", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/alerts/"+alert.ID, nil)
		w := httptest.NewRecorder()

		app.handleAlertDetail().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var fetched alerts.Alert
		err := json.Unmarshal(w.Body.Bytes(), &fetched)
		require.NoError(t, err)
		assert.Equal(t, alert.ID, fetched.ID)
		assert.Equal(t, "Detail Test Alert", fetched.Title)
	})

	t.Run("Get Not Found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/alerts/non-existent", nil)
		w := httptest.NewRecorder()

		app.handleAlertDetail().ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Patch Alert", func(t *testing.T) {
		update := &alerts.Alert{
			Status: alerts.StatusResolved,
		}
		body, _ := json.Marshal(update)
		req := httptest.NewRequest(http.MethodPatch, "/alerts/"+alert.ID, bytes.NewReader(body))
		w := httptest.NewRecorder()

		app.handleAlertDetail().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var updated alerts.Alert
		err := json.Unmarshal(w.Body.Bytes(), &updated)
		require.NoError(t, err)
		assert.Equal(t, alerts.StatusResolved, updated.Status)

		// Verify in manager
		stored := app.AlertsManager.GetAlert(alert.ID)
		assert.Equal(t, alerts.StatusResolved, stored.Status)
	})

	t.Run("Patch Not Found", func(t *testing.T) {
		update := &alerts.Alert{Status: alerts.StatusResolved}
		body, _ := json.Marshal(update)
		req := httptest.NewRequest(http.MethodPatch, "/alerts/non-existent", bytes.NewReader(body))
		w := httptest.NewRecorder()

		app.handleAlertDetail().ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Patch Bad Request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/alerts/"+alert.ID, bytes.NewReader([]byte("invalid")))
		w := httptest.NewRecorder()

		app.handleAlertDetail().ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Missing ID", func(t *testing.T) {
		// This case is tricky because usually router handles path params.
		// But handleAlertDetail uses strings.TrimPrefix(r.URL.Path, "/alerts/")
		// If we request /alerts/, prefix trim gives empty string.
		req := httptest.NewRequest(http.MethodGet, "/alerts/", nil)
		w := httptest.NewRecorder()

		app.handleAlertDetail().ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/alerts/"+alert.ID, nil)
		w := httptest.NewRecorder()

		app.handleAlertDetail().ServeHTTP(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestHandleAlertRules_Full(t *testing.T) {
	app := setupAlertsTestApp()

	t.Run("List Rules", func(t *testing.T) {
		app.AlertsManager.CreateRule(&alerts.AlertRule{Name: "Rule 1"})
		app.AlertsManager.CreateRule(&alerts.AlertRule{Name: "Rule 2"})

		req := httptest.NewRequest(http.MethodGet, "/alerts/rules", nil)
		w := httptest.NewRecorder()

		app.handleAlertRules().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var list []*alerts.AlertRule
		err := json.Unmarshal(w.Body.Bytes(), &list)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 2)
	})

	t.Run("Create Rule", func(t *testing.T) {
		rule := &alerts.AlertRule{Name: "New Rule", Metric: "cpu"}
		body, _ := json.Marshal(rule)
		req := httptest.NewRequest(http.MethodPost, "/alerts/rules", bytes.NewReader(body))
		w := httptest.NewRecorder()

		app.handleAlertRules().ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var created alerts.AlertRule
		err := json.Unmarshal(w.Body.Bytes(), &created)
		require.NoError(t, err)
		assert.Equal(t, "New Rule", created.Name)
		assert.NotEmpty(t, created.ID)
	})

	t.Run("Create Bad Request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/alerts/rules", bytes.NewReader([]byte("invalid")))
		w := httptest.NewRecorder()

		app.handleAlertRules().ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/alerts/rules", nil)
		w := httptest.NewRecorder()

		app.handleAlertRules().ServeHTTP(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestHandleAlertRuleDetail_Full(t *testing.T) {
	app := setupAlertsTestApp()
	rule := app.AlertsManager.CreateRule(&alerts.AlertRule{Name: "Test Rule"})

	t.Run("Get Rule", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/alerts/rules/"+rule.ID, nil)
		w := httptest.NewRecorder()

		app.handleAlertRuleDetail().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var fetched alerts.AlertRule
		err := json.Unmarshal(w.Body.Bytes(), &fetched)
		require.NoError(t, err)
		assert.Equal(t, rule.ID, fetched.ID)
	})

	t.Run("Get Not Found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/alerts/rules/non-existent", nil)
		w := httptest.NewRecorder()

		app.handleAlertRuleDetail().ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Update Rule", func(t *testing.T) {
		update := &alerts.AlertRule{Name: "Updated Rule", Metric: "mem"}
		body, _ := json.Marshal(update)
		req := httptest.NewRequest(http.MethodPut, "/alerts/rules/"+rule.ID, bytes.NewReader(body))
		w := httptest.NewRecorder()

		app.handleAlertRuleDetail().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var updated alerts.AlertRule
		err := json.Unmarshal(w.Body.Bytes(), &updated)
		require.NoError(t, err)
		assert.Equal(t, "Updated Rule", updated.Name)

		stored := app.AlertsManager.GetRule(rule.ID)
		assert.Equal(t, "Updated Rule", stored.Name)
	})

	t.Run("Update Not Found", func(t *testing.T) {
		update := &alerts.AlertRule{Name: "Updated Rule"}
		body, _ := json.Marshal(update)
		req := httptest.NewRequest(http.MethodPut, "/alerts/rules/non-existent", bytes.NewReader(body))
		w := httptest.NewRecorder()

		app.handleAlertRuleDetail().ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Update Bad Request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/alerts/rules/"+rule.ID, bytes.NewReader([]byte("invalid")))
		w := httptest.NewRecorder()

		app.handleAlertRuleDetail().ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Delete Rule", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/alerts/rules/"+rule.ID, nil)
		w := httptest.NewRecorder()

		app.handleAlertRuleDetail().ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Nil(t, app.AlertsManager.GetRule(rule.ID))
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/alerts/rules/"+rule.ID, nil)
		w := httptest.NewRecorder()

		app.handleAlertRuleDetail().ServeHTTP(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("Missing ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/alerts/rules/", nil)
		w := httptest.NewRecorder()

		app.handleAlertRuleDetail().ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
