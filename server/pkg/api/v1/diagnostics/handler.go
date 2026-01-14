// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package diagnostics provides the HTTP handler for the diagnostics API.
package diagnostics

import (
	"encoding/json"
	"net/http"

	pkg_diagnostics "github.com/mcpany/core/server/pkg/diagnostics"
	"github.com/mcpany/core/server/pkg/logging"
)

// Handler handles diagnostics requests.
type Handler struct {
	diagnosticsService *pkg_diagnostics.Service
}

// NewHandler creates a new diagnostics handler.
func NewHandler(diagnosticsService *pkg_diagnostics.Service) *Handler {
	return &Handler{
		diagnosticsService: diagnosticsService,
	}
}

// ServeHTTP handles the diagnostics request.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	report, err := h.diagnosticsService.GenerateReport(r.Context())
	if err != nil {
		logging.GetLogger().Error("Failed to generate diagnostics report", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		logging.GetLogger().Error("Failed to encode diagnostics report", "error", err)
	}
}
