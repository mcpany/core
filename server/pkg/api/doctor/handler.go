// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package doctor provides a diagnostic endpoint for the server.
package doctor

import (
	"encoding/json"
	"net/http"
	"os"
	"runtime"
	"strconv"
)

// DiagnosticReport contains information about the server environment and runtime.
type DiagnosticReport struct {
	Status      string            `json:"status"`
	Environment map[string]string `json:"environment"`
	Runtime     map[string]string `json:"runtime"`
}

// Handler handles the /api/v1/health/doctor endpoint.
// It returns a DiagnosticReport with system information.
func Handler(w http.ResponseWriter, _ *http.Request) {
	report := DiagnosticReport{
		Status: "ok",
		Environment: map[string]string{
			"GOOS":   runtime.GOOS,
			"GOARCH": runtime.GOARCH,
			"PID":    strconv.Itoa(os.Getpid()),
		},
		Runtime: map[string]string{
			"Version":      runtime.Version(),
			"NumGoroutine": strconv.Itoa(runtime.NumGoroutine()),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
