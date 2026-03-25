// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"context"
	"encoding/json"
	"net/http"
// CheckResult represents a single check result.
//
// Summary: The outcome of a single health check execution.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
type CheckResult struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
// CheckFunc is a function that performs a health check.
// DoctorReport represents the full doctor report.
//
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// Summary: Aggregated health report containing all check results.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
type DoctorReport struct {
// NewDoctor creates a new Doctor.
//
// Summary: Initializes a new Doctor instance.
//
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// Returns:
// AddCheck adds a named health check.
//
// Summary: Registers a custom health check function.
// Handler returns the http handler.
// Parameters:
//   - standard arguments based on function signature.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Summary: Returns an HTTP handler that runs all checks and returns a JSON report.
//
// Returns:
//   - http.HandlerFunc: The HTTP handler.
//
// Parameters:
//   - standard arguments based on function signature.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Side Effects:
//   - Executes all registered health checks.
//   - Makes an external network call to google.com (connectivity check).
//   - Reads environment variables (Auth checks).
//   - Writes JSON response to the client.
// Parameters:
//   - standard arguments based on function signature.
// Errors:
//   - triggers relevant error states on failure.
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
