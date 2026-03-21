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

// Summary: CheckResult represents a single check result. The outcome of a single health check execution.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type CheckResult struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
	Diff    string `json:"diff,omitempty"`
}

// Summary: CheckFunc is a function that performs a health check. Function signature for a health check execution logic.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type CheckFunc func(context.Context) CheckResult

// Summary: DoctorReport represents the full doctor report. Aggregated health report containing all check results.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type DoctorReport struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]CheckResult `json:"checks"`
}

// Summary: Doctor is the health check handler. Registry and handler for system health checks (Doctor).
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type Doctor struct {
	checks     map[string]CheckFunc
	mu         sync.RWMutex
	httpClient *http.Client
}

// Summary: NewDoctor creates a new Doctor. Initializes a new Doctor instance.
//
// Parameters:
//   - None.
//
// Returns:
//   - *Doctor: The resulting *Doctor.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewDoctor() *Doctor {
	return &Doctor{
		checks:     make(map[string]CheckFunc),
		httpClient: http.DefaultClient,
	}
}

// Summary: AddCheck adds a named health check. Registers a custom health check function.
//
// Parameters:
//   - name (string): The name parameter.
//   - check (CheckFunc): The check parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (d *Doctor) AddCheck(name string, check CheckFunc) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.checks[name] = check
}

// Summary: Handler returns the http handler. Returns an HTTP handler that runs all checks and returns a JSON report.
//
// Parameters:
//   - None.
//
// Returns:
//   - http.HandlerFunc: The resulting http.HandlerFunc.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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
