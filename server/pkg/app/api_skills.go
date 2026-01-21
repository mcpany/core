// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"io"
	"net/http"
	"strings"

	"github.com/mcpany/core/server/pkg/logging"
)

// handleUploadSkillAsset returns a handler that saves an asset for a skill.
// POST /api/v1/skills/{name}/assets
// Query Param: path (relative path).
func (a *Application) handleUploadSkillAsset() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract skill name manually since we aren't using a router with params
		// Path expected: /api/v1/skills/{name}/assets or /v1/skills/{name}/assets
		parts := strings.Split(r.URL.Path, "/")
		var skillName string

		// Handle /api/v1 prefix or /v1 prefix
		switch {
		case len(parts) >= 6 && parts[5] == "assets" && parts[3] == "skills":
			skillName = parts[4]
		case len(parts) >= 5 && parts[4] == "assets" && parts[2] == "skills":
			skillName = parts[3]
		default:
			http.Error(w, "Invalid URL format", http.StatusBadRequest)
			return
		}

		if skillName == "" {
			http.Error(w, "Skill name required", http.StatusBadRequest)
			return
		}

		assetPath := r.URL.Query().Get("path")
		if assetPath == "" {
			http.Error(w, "Asset path query parameter required", http.StatusBadRequest)
			return
		}

		// Limit upload size to 25MB to prevent DoS
		r.Body = http.MaxBytesReader(w, r.Body, 25<<20)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			if strings.Contains(err.Error(), "request body too large") {
				http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
				return
			}
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		if err := a.SkillManager.SaveAsset(skillName, assetPath, body); err != nil {
			logging.GetLogger().Error("Failed to save asset", "skill", skillName, "path", assetPath, "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
