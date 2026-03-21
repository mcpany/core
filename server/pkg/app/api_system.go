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
}

func (a *Application) handleSystemStatus(w http.ResponseWriter, _ *http.Request) {
	uptime := int64(time.Since(a.startTime).Seconds())
	activeConns := atomic.LoadInt32(&a.activeConnections)

	warnings := []string{}
	if a.SettingsManager.GetAPIKey() == "" {
		warnings = append(warnings, "No API Key configured")
	}

	// Check if listening on all interfaces is complex due to config structure.
	// For now, focus on API key warning.

	resp := SystemStatusResponse{
		UptimeSeconds:     uptime,
		ActiveConnections: activeConns,
		BoundHTTPPort:     int(a.BoundHTTPPort.Load()),
		BoundGRPCPort:     int(a.BoundGRPCPort.Load()),
		Version:           appconsts.Version,
		SecurityWarnings:  warnings,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logging.GetLogger().Error("Failed to encode system status response", "error", err)
	}
}
