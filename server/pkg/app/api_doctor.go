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

func (a *Application) handleServiceDiagnose(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/services/")
		parts := strings.Split(path, "/")
		if len(parts) < 1 || parts[0] == "" {
			http.Error(w, "name required", http.StatusBadRequest)
			return
		}
		name := parts[0]

		svc, err := store.GetService(r.Context(), name)
		if err != nil {
			logging.GetLogger().Error("failed to get service for diagnostics", "name", name, "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if svc == nil {
			http.NotFound(w, r)
			return
		}

		// Run diagnostics
		result := doctor.CheckService(r.Context(), svc)
		result.ServiceName = svc.GetName()

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			logging.GetLogger().Error("failed to encode diagnostic result", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}
