// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/mcpany/core/server/pkg/appconsts"
	"github.com/mcpany/core/server/pkg/logging"
)

// SystemStatusResponse represents the response from the system status API.
type SystemStatusResponse struct {
	UptimeSeconds     int64    `json:"uptime_seconds"`
	ActiveConnections int32    `json:"active_connections"`
	BoundHTTPPort     int      `json:"bound_http_port"`
	BoundGRPCPort     int      `json:"bound_grpc_port"`
	Version           string   `json:"version"`
	SecurityWarnings  []string `json:"security_warnings"`
	ConfigStatus      string   `json:"config_status,omitempty"`
	ConfigError       string   `json:"last_reload_error,omitempty"`
	ConfigReloadTime  string   `json:"last_reload_time,omitempty"`
}

func (a *Application) handleSystemStatus(w http.ResponseWriter, r *http.Request) {
	uptime := int64(time.Since(a.startTime).Seconds())
	activeConns := atomic.LoadInt32(&a.activeConnections)

	warnings := []string{}
	if a.SettingsManager.GetAPIKey() == "" {
		warnings = append(warnings, "No API Key configured")
	}

	// Check config health
	// We reuse the logic from configHealthCheck but expose it directly in the status
	// so the UI can show a banner without polling the heavier doctor endpoint.
	check := a.configHealthCheck(r.Context())

	resp := SystemStatusResponse{
		UptimeSeconds:     uptime,
		ActiveConnections: activeConns,
		BoundHTTPPort:     int(a.BoundHTTPPort.Load()),
		BoundGRPCPort:     int(a.BoundGRPCPort.Load()),
		Version:           appconsts.Version,
		SecurityWarnings:  warnings,
		ConfigStatus:      check.Status,
		ConfigError:       check.Message,
		ConfigReloadTime:  check.Latency, // Latency field in checkResult is actually "time since reload"
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logging.GetLogger().Error("Failed to encode system status response", "error", err)
	}
}
