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

func TestHandleAlertRules(t *testing.T) {
	app := NewApplication()
	// NewApplication creates a default AlertsManager with seeded data.
	// We can use it directly or mock it.
	// For integration test of handler <-> manager, using the real one is fine.

	// 1. List Rules (Initial Seed)
	t.Run("List Rules", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/alerts/rules", nil)
		rr := httptest.NewRecorder()
		app.handleAlertRules().ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var rules []*alerts.AlertRule
		err := json.Unmarshal(rr.Body.Bytes(), &rules)
		require.NoError(t, err)
		assert.NotEmpty(t, rules, "Expected seeded rules")
	})

	// 2. Create Rule
	var createdID string
	t.Run("Create Rule", func(t *testing.T) {
		rule := &alerts.AlertRule{
			Name:      "API Error Rate",
			Metric:    "error_rate",
			Operator:  ">",
			Threshold: 0.05,
			Enabled:   true,
			Severity:  alerts.SeverityCritical,
		}
		body, _ := json.Marshal(rule)
		req := httptest.NewRequest(http.MethodPost, "/alerts/rules", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		app.handleAlertRules().ServeHTTP(rr, req)

		require.Equal(t, http.StatusCreated, rr.Code)
		var created alerts.AlertRule
		err := json.Unmarshal(rr.Body.Bytes(), &created)
		require.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.Equal(t, "API Error Rate", created.Name)
		createdID = created.ID
	})

	// 3. Get Rule Detail
	t.Run("Get Rule Detail", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/alerts/rules/"+createdID, nil)
		rr := httptest.NewRecorder()
		app.handleAlertRuleDetail().ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var fetched alerts.AlertRule
		err := json.Unmarshal(rr.Body.Bytes(), &fetched)
		require.NoError(t, err)
		assert.Equal(t, createdID, fetched.ID)
	})

	// 4. Update Rule
	t.Run("Update Rule", func(t *testing.T) {
		update := &alerts.AlertRule{
			Name:      "API Error Rate Updated",
			Metric:    "error_rate",
			Operator:  ">",
			Threshold: 0.10,
			Enabled:   false,
			Severity:  alerts.SeverityWarning,
		}
		body, _ := json.Marshal(update)
		req := httptest.NewRequest(http.MethodPut, "/alerts/rules/"+createdID, bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		app.handleAlertRuleDetail().ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var updated alerts.AlertRule
		err := json.Unmarshal(rr.Body.Bytes(), &updated)
		require.NoError(t, err)
		assert.Equal(t, "API Error Rate Updated", updated.Name)
		assert.Equal(t, 0.10, updated.Threshold)
	})

	// 5. Delete Rule
	t.Run("Delete Rule", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/alerts/rules/"+createdID, nil)
		rr := httptest.NewRecorder()
		app.handleAlertRuleDetail().ServeHTTP(rr, req)

		require.Equal(t, http.StatusNoContent, rr.Code)

		// Verify deletion
		reqGet := httptest.NewRequest(http.MethodGet, "/alerts/rules/"+createdID, nil)
		rrGet := httptest.NewRecorder()
		app.handleAlertRuleDetail().ServeHTTP(rrGet, reqGet)
		require.Equal(t, http.StatusNotFound, rrGet.Code) // Should be 404
	})
}
