// Copyright 2026 Author(s) of MCP Any
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

func TestAlertsAPI(t *testing.T) {
	app := setupTestApp()
	// AlertsManager is already initialized with seed data in setupTestApp -> NewApplication -> alerts.NewManager -> seedData

	// 1. List Alerts
	t.Run("list alerts", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/alerts", nil)
		rr := httptest.NewRecorder()
		app.handleAlerts()(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var list []*alerts.Alert
		err := json.Unmarshal(rr.Body.Bytes(), &list)
		require.NoError(t, err)
		assert.NotEmpty(t, list) // Seed data exists
	})

	// 2. Create Alert
	var createdID string
	t.Run("create alert", func(t *testing.T) {
		alert := &alerts.Alert{
			Title:    "Test Alert",
			Message:  "Test Message",
			Severity: alerts.SeverityInfo,
			Status:   alerts.StatusActive,
		}
		body, _ := json.Marshal(alert)
		req := httptest.NewRequest(http.MethodPost, "/alerts", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		app.handleAlerts()(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		var resp alerts.Alert
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "Test Alert", resp.Title)
		assert.NotEmpty(t, resp.ID)
		createdID = resp.ID
	})

	// 3. Get Alert Detail
	t.Run("get alert detail", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/alerts/"+createdID, nil)
		rr := httptest.NewRecorder()
		app.handleAlertDetail()(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp alerts.Alert
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, createdID, resp.ID)
	})

	// 4. Update Alert (Patch)
	t.Run("update alert", func(t *testing.T) {
		update := &alerts.Alert{
			Status: alerts.StatusResolved,
		}
		body, _ := json.Marshal(update)
		req := httptest.NewRequest(http.MethodPatch, "/alerts/"+createdID, bytes.NewReader(body))
		rr := httptest.NewRecorder()
		app.handleAlertDetail()(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp alerts.Alert
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, alerts.StatusResolved, resp.Status)
	})
}

func TestAlertRulesAPI(t *testing.T) {
	app := setupTestApp()

	// 1. List Rules
	t.Run("list rules", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/alerts/rules", nil)
		rr := httptest.NewRecorder()
		app.handleAlertRules()(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var list []*alerts.AlertRule
		err := json.Unmarshal(rr.Body.Bytes(), &list)
		require.NoError(t, err)
		assert.NotEmpty(t, list) // Seed data
	})

	// 2. Create Rule
	var createdID string
	t.Run("create rule", func(t *testing.T) {
		rule := &alerts.AlertRule{
			Name:      "Test Rule",
			Metric:    "cpu",
			Operator:  ">",
			Threshold: 80,
			Duration:  "1m",
			Severity:  alerts.SeverityWarning,
			Enabled:   true,
		}
		body, _ := json.Marshal(rule)
		req := httptest.NewRequest(http.MethodPost, "/alerts/rules", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		app.handleAlertRules()(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		var resp alerts.AlertRule
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "Test Rule", resp.Name)
		assert.NotEmpty(t, resp.ID)
		createdID = resp.ID
	})

	// 3. Get Rule Detail
	t.Run("get rule detail", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/alerts/rules/"+createdID, nil)
		rr := httptest.NewRecorder()
		app.handleAlertRuleDetail()(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp alerts.AlertRule
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, createdID, resp.ID)
	})

	// 4. Update Rule (Put)
	t.Run("update rule", func(t *testing.T) {
		rule := &alerts.AlertRule{
			Name:      "Updated Rule",
			Metric:    "cpu",
			Operator:  ">",
			Threshold: 90,
			Duration:  "2m",
			Severity:  alerts.SeverityCritical,
			Enabled:   false,
		}
		body, _ := json.Marshal(rule)
		req := httptest.NewRequest(http.MethodPut, "/alerts/rules/"+createdID, bytes.NewReader(body))
		rr := httptest.NewRecorder()
		app.handleAlertRuleDetail()(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp alerts.AlertRule
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "Updated Rule", resp.Name)
		assert.Equal(t, alerts.SeverityCritical, resp.Severity)
		assert.False(t, resp.Enabled)
	})

	// 5. Delete Rule
	t.Run("delete rule", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/alerts/rules/"+createdID, nil)
		rr := httptest.NewRecorder()
		app.handleAlertRuleDetail()(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)

		// Verify gone
		req = httptest.NewRequest(http.MethodGet, "/alerts/rules/"+createdID, nil)
		rr = httptest.NewRecorder()
		app.handleAlertRuleDetail()(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}
