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

// CheckResult - Auto-generated documentation.
//
// Summary: CheckResult represents a single check result.
//
// Fields:
//   - Various fields for CheckResult.
type CheckResult struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
	Diff    string `json:"diff,omitempty"`
}

// CheckFunc is a function that performs a health check.
//
// Summary: Function signature for a health check execution logic.
type CheckFunc func(context.Context) CheckResult

// DoctorReport - Auto-generated documentation.
//
// Summary: DoctorReport represents the full doctor report.
//
// Fields:
//   - Various fields for DoctorReport.
type DoctorReport struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]CheckResult `json:"checks"`
}

// Doctor - Auto-generated documentation.
//
// Summary: Doctor is the health check handler.
//
// Fields:
//   - Various fields for Doctor.
type Doctor struct {
	checks     map[string]CheckFunc
	mu         sync.RWMutex
	httpClient *http.Client
}

// NewDoctor - Auto-generated documentation.
//
// Summary: NewDoctor creates a new Doctor.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func NewDoctor() *Doctor {
	return &Doctor{
		checks:     make(map[string]CheckFunc),
		httpClient: http.DefaultClient,
	}
}

// AddCheck - Auto-generated documentation.
//
// Summary: AddCheck adds a named health check.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func (d *Doctor) AddCheck(name string, check CheckFunc) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.checks[name] = check
}

// Handler - Auto-generated documentation.
//
// Summary: Handler returns the http handler.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
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
