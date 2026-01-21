// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleDashboardMetrics_Coverage(t *testing.T) {
	app := NewApplication()

	req := httptest.NewRequest("GET", "/dashboard/metrics", nil)
	rr := httptest.NewRecorder()

	// We need to use the handler that calls handleDashboardMetrics
	// But handleDashboardMetrics is private?
	// It's in dashboard.go: `func (a *Application) handleDashboardMetrics(w http.ResponseWriter, r *http.Request)`
	// Wait, is it a method on Application?
	// `dashboard.go` package is `app`? Yes.

	// Check if handleDashboardMetrics is exported or not.
	// `func (a *Application) handleDashboardMetrics` -> lower case.
	// But I am in `package app` (same package in `server_extras_test.go` if I use `package app`).
	// `server_extras_test.go` uses `package app`.

	// So I can call it.
	handler := app.handleDashboardMetrics()
	handler.ServeHTTP(rr, req)

	// It relies on `a.activeConnections` etc.
	// Should return 200 and JSON.
	assert.Equal(t, http.StatusOK, rr.Code)
	// Response is a list of metrics, check for "Active Services" label
	assert.Contains(t, rr.Body.String(), "Active Services")
}
