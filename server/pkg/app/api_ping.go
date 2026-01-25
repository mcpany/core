// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
)

// handleServicePing handles the ping request for a service.
// POST /services/{name}/ping
func (a *Application) handleServicePing(w http.ResponseWriter, r *http.Request, name string, store storage.Storage) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Get Service Config
	svc, err := store.GetService(r.Context(), name)
	if err != nil {
		logging.GetLogger().Error("failed to get service for ping", "name", name, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if svc == nil {
		http.NotFound(w, r)
		return
	}

	// 2. Perform Connectivity Check
	valid, details, latency, err := performConnectivityCheck(r.Context(), svc)

	// 3. Return Result
	response := map[string]any{
		"status":  "ok",
		"latency": latency.String(),
		"details": details,
	}

	if !valid {
		response["status"] = "error"
		if err != nil {
			response["error"] = err.Error()
		} else {
			response["error"] = details
		}
	} else {
		// If valid, we can imply it's reachable.
		// Latency measurement would be nice to return from performConnectivityCheck,
		// but for now we just say "ok".
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
