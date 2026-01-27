// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"os"
	"runtime"
	"strings"
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

// SystemDiagnosticsResponse represents the response from the system diagnostics API.
type SystemDiagnosticsResponse struct {
	OS      string          `json:"os"`
	Arch    string          `json:"arch"`
	Docker  bool            `json:"docker"`
	EnvVars map[string]bool `json:"env_vars"`
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

func (a *Application) handleSystemDiagnostics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check for Docker environment
	isDocker := false
	if _, err := os.Stat("/.dockerenv"); err == nil {
		isDocker = true
	} else {
		// Fallback check cgroups
		if content, err := os.ReadFile("/proc/1/cgroup"); err == nil {
			if strings.Contains(string(content), "docker") {
				isDocker = true
			}
		}
	}

	envVars := make(map[string]bool)
	checkEnvs := []string{"HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY", "ALL_PROXY", "http_proxy", "https_proxy", "no_proxy", "all_proxy"}
	for _, key := range checkEnvs {
		if val := os.Getenv(key); val != "" {
			envVars[strings.ToUpper(key)] = true
		} else {
			envVars[strings.ToUpper(key)] = false
		}
	}

	resp := SystemDiagnosticsResponse{
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
		Docker:  isDocker,
		EnvVars: envVars,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logging.GetLogger().Error("Failed to encode system diagnostics response", "error", err)
	}
}
