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

// CheckResult defines the core structure for check result within the system.
//
// Summary: CheckResult defines the core structure for check result within the system.
//
// Fields:
//   - Contains the configuration and state properties required for CheckResult functionality.
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

// DoctorReport defines the core structure for doctor report within the system.
//
// Summary: DoctorReport defines the core structure for doctor report within the system.
//
// Fields:
//   - Contains the configuration and state properties required for DoctorReport functionality.
type DoctorReport struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]CheckResult `json:"checks"`
}

// Doctor is the health check handler. Summary: Registry and handler for system health checks (Doctor).
//
// Summary: Doctor is the health check handler. Summary: Registry and handler for system health checks (Doctor).
//
// Fields:
//   - Contains the configuration and state properties required for Doctor functionality.
type Doctor struct {
	checks     map[string]CheckFunc
	mu         sync.RWMutex
	httpClient *http.Client
}

// NewDoctor creates a new Doctor. Summary: Initializes a new Doctor instance. Returns: - *Doctor: The initialized doctor registry. Side Effects: - Initializes internal maps and HTTP client.
//
// Summary: NewDoctor creates a new Doctor. Summary: Initializes a new Doctor instance. Returns: - *Doctor: The initialized doctor registry. Side Effects: - Initializes internal maps and HTTP client.
//
// Parameters:
//   - None.
//
// Returns:
//   - (*Doctor): The resulting Doctor object containing the requested data.
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

// AddCheck adds a named health check. Summary: Registers a custom health check function. Parameters: - name: string. The unique name of the check. - check: CheckFunc. The function to execute. Side Effects: - Updates the internal checks map.
//
// Summary: AddCheck adds a named health check. Summary: Registers a custom health check function. Parameters: - name: string. The unique name of the check. - check: CheckFunc. The function to execute. Side Effects: - Updates the internal checks map.
//
// Parameters:
//   - name (string): The name parameter used in the operation.
//   - check (CheckFunc): The check parameter used in the operation.
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

// Handler returns the http handler. Summary: Returns an HTTP handler that runs all checks and returns a JSON report. Returns: - http.HandlerFunc: The HTTP handler. Side Effects: - Executes all registered health checks. - Makes an external network call to google.com (connectivity check). - Reads environment variables (Auth checks). - Writes JSON response to the client.
//
// Summary: Handler returns the http handler. Summary: Returns an HTTP handler that runs all checks and returns a JSON report. Returns: - http.HandlerFunc: The HTTP handler. Side Effects: - Executes all registered health checks. - Makes an external network call to google.com (connectivity check). - Reads environment variables (Auth checks). - Writes JSON response to the client.
//
// Parameters:
//   - None.
//
// Returns:
//   - (http.HandlerFunc): The resulting http.HandlerFunc object containing the requested data.
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
