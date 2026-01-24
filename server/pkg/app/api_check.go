// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"

	"github.com/mcpany/core/server/pkg/doctor"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
)

// handleServiceCheck handles the request to check the health of a specific service.
func (a *Application) handleServiceCheck(w http.ResponseWriter, r *http.Request, name string, store storage.Storage) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Get the service configuration
	svc, err := store.GetService(r.Context(), name)
	if err != nil {
		logging.GetLogger().Error("failed to get service for check", "name", name, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if svc == nil {
		http.NotFound(w, r)
		return
	}

	// 2. Run the doctor check
	result := doctor.CheckService(r.Context(), svc)

	// 3. Map the result to a JSON-friendly response
	// doctor.CheckResult has an error field which doesn't marshal well.
	type CheckResultJSON struct {
		ServiceName string `json:"service_name"`
		Status      string `json:"status"`
		Message     string `json:"message"`
		Error       string `json:"error,omitempty"`
	}

	jsonResult := CheckResultJSON{
		ServiceName: result.ServiceName,
		Status:      string(result.Status),
		Message:     result.Message,
	}
	if result.Error != nil {
		jsonResult.Error = result.Error.Error()
	}

	// 4. Respond
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(jsonResult); err != nil {
		logging.GetLogger().Error("failed to encode check result", "error", err)
	}
}
