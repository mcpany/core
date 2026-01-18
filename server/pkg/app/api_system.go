// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/mcpany/core/server/pkg/appconsts"
)

// handleSystemStatus returns the current system status.
func (a *Application) handleSystemStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	warnings := []string{}
	// Check for insecure binding
	// Auth check:
	apiKey := a.SettingsManager.GetAPIKey()
	if apiKey == "" {
		warnings = append(warnings, "No API Key configured. Public access restricted to localhost.")
	}

	status := map[string]any{
		"uptime_seconds":     time.Since(a.startTime).Seconds(),
		"version":            appconsts.Version,
		"active_connections": atomic.LoadInt64(&a.activeConnections),
		"bound_http_port":    a.BoundHTTPPort,
		"bound_grpc_port":    a.BoundGRPCPort,
		"security_warnings":  warnings,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
