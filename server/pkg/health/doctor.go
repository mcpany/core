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
type CheckResult struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
	Diff    string `json:"diff,omitempty"`
}

// CheckFunc is a function that performs a health check.
type CheckFunc func(context.Context) CheckResult

// DoctorReport represents the full doctor report.
type DoctorReport struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]CheckResult `json:"checks"`
}

// Doctor is the health check handler.
type Doctor struct {
	checks     map[string]CheckFunc
	mu         sync.RWMutex
	httpClient *http.Client
}

// NewDoctor initializes a new Doctor instance for performing health checks.
// It sets up the default checks and HTTP client.
//
// Returns:
//   - *Doctor: A new Doctor instance.
func NewDoctor() *Doctor {
	return &Doctor{
		checks:     make(map[string]CheckFunc),
		httpClient: http.DefaultClient,
	}
}

// AddCheck registers a custom health check function with a given name.
//
// Parameters:
//   - name: string. The unique name of the check (e.g., "database", "redis").
//   - check: CheckFunc. The function to execute for the check.
func (d *Doctor) AddCheck(name string, check CheckFunc) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.checks[name] = check
}

// Handler returns an HTTP handler that executes all registered health checks.
// It returns a JSON report containing the status of each check and the overall system health.
//
// Returns:
//   - http.HandlerFunc: The HTTP handler function.
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
