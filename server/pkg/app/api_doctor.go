// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mcpany/core/server/pkg/doctor"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
)

// handleServiceDiagnose handles the diagnostic check for a specific service.
// POST /api/v1/services/{name}/diagnose.
func (a *Application) handleServiceDiagnose(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract service name from URL
		// Expected path: /api/v1/services/{name}/diagnose
		path := strings.TrimPrefix(r.URL.Path, "/services/")
		path = strings.TrimSuffix(path, "/diagnose")
		name := path

		if name == "" {
			http.Error(w, "service name is required", http.StatusBadRequest)
			return
		}

		svc, err := store.GetService(r.Context(), name)
		if err != nil {
			logging.GetLogger().Error("failed to get service for diagnosis", "name", name, "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if svc == nil {
			http.NotFound(w, r)
			return
		}

		// Run the check
		result := doctor.CheckService(r.Context(), svc)

		// Map to JSON-friendly struct
		resp := struct {
			ServiceName string `json:"service_name"`
			Status      string `json:"status"`
			Message     string `json:"message"`
			Error       string `json:"error,omitempty"`
		}{
			ServiceName: svc.GetName(),
			Status:      string(result.Status),
			Message:     result.Message,
		}

		if result.Error != nil {
			resp.Error = result.Error.Error()
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logging.GetLogger().Error("failed to encode diagnostic response", "error", err)
		}
	}
}
