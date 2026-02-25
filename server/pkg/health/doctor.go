// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// CheckResult represents a single check result.
//
// Summary: Stores the outcome of a health check.
type CheckResult struct {
	// Status indicates the health status (e.g., "ok", "degraded", "unhealthy").
	Status string `json:"status"`
	// Message provides additional details or error descriptions.
	Message string `json:"message,omitempty"`
	// Latency records the duration of the check.
	Latency string `json:"latency,omitempty"`
	// Diff captures any state difference if applicable.
	Diff string `json:"diff,omitempty"`
}

// CheckFunc is a function that performs a health check.
//
// Summary: Function signature for health checks.
//
// Parameters:
//   - ctx (context.Context): The context for the check.
//
// Returns:
//   - CheckResult: The result of the health check.
type CheckFunc func(context.Context) CheckResult

// DoctorReport represents the full doctor report.
//
// Summary: Aggregated health report for all checks.
type DoctorReport struct {
	// Status is the overall system health status.
	Status string `json:"status"`
	// Timestamp indicates when the report was generated.
	Timestamp time.Time `json:"timestamp"`
	// Checks maps check names to their individual results.
	Checks map[string]CheckResult `json:"checks"`
}

// Doctor is the health check handler.
//
// Summary: Manages and executes health checks.
type Doctor struct {
	checks     map[string]CheckFunc
	mu         sync.RWMutex
	httpClient *http.Client
}

// NewDoctor creates a new Doctor instance.
//
// Summary: Initializes a new Doctor for health checking.
//
// Returns:
//   - *Doctor: A pointer to the initialized Doctor instance.
//
// Side Effects:
//   - Allocates memory for the Doctor struct and its internal map.
func NewDoctor() *Doctor {
	return &Doctor{
		checks:     make(map[string]CheckFunc),
		httpClient: http.DefaultClient,
	}
}

// AddCheck registers a new named health check.
//
// Summary: Adds a health check to the Doctor.
//
// Parameters:
//   - name (string): The unique name of the check resource.
//   - check (CheckFunc): The function to execute for the check.
//
// Returns:
//   - None.
//
// Side Effects:
//   - Modifies the internal checks map of the Doctor instance.
func (d *Doctor) AddCheck(name string, check CheckFunc) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.checks[name] = check
}

// Handler returns the HTTP handler for health reports.
//
// Summary: Provides an HTTP handler to expose health check results.
//
// Returns:
//   - http.HandlerFunc: An HTTP handler function that returns a JSON health report.
//
// Side Effects:
//   - None (the returned handler has side effects when executed, but this function does not).
func (d *Doctor) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		report := DoctorReport{
			Status:    "healthy",
			Timestamp: time.Now(),
			Checks:    make(map[string]CheckResult),
		}

		// Check Internet
		start := time.Now()
		req, err := http.NewRequestWithContext(r.Context(), "GET", "https://www.google.com", nil)
		if err == nil {
			var resp *http.Response
			resp, err = d.httpClient.Do(req)
			if resp != nil {
				defer func() {
					_ = resp.Body.Close()
				}()
			}
		}

		if err != nil {
			report.Checks["internet"] = CheckResult{
				Status:  "degraded",
				Message: err.Error(),
			}
			report.Status = "degraded"
		} else {
			report.Checks["internet"] = CheckResult{
				Status:  "ok",
				Latency: time.Since(start).String(),
			}
		}

		// Auth Checks
		authChecks := CheckAuth()
		for k, v := range authChecks {
			report.Checks[k] = v
		}

		// Execute dynamic checks
		d.mu.RLock()
		for name, check := range d.checks {
			res := check(r.Context())
			report.Checks[name] = res
			if res.Status != "ok" {
				report.Status = "degraded"
			}
		}
		d.mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(report)
	}
}
