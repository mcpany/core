// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mcpany/core/server/pkg/audit"
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
		// Path expected: /api/v1/skills/{name}/assets
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 6 || parts[5] != "assets" {
			http.Error(w, "Invalid URL format", http.StatusBadRequest)
			return
		}
		skillName := parts[4]

		assetPath := r.URL.Query().Get("path")
		if assetPath == "" {
			http.Error(w, "Asset path query parameter required", http.StatusBadRequest)
			return
		}

		// Limit the request body size to 10MB to prevent DoS attacks
		r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

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

			// Audit log failure
			if a.standardMiddlewares != nil && a.standardMiddlewares.Audit != nil {
				args := map[string]string{
					"skill": skillName,
					"path":  assetPath,
				}
				argsBytes, _ := json.Marshal(args)
				entry := audit.Entry{
					ToolName:  "SkillAssetUpload",
					Arguments: json.RawMessage(argsBytes),
					Error:     err.Error(),
				}
				_ = a.standardMiddlewares.Audit.Log(r.Context(), entry)
			}

			// Sentinel Security: Return generic error to avoid leaking path information or internal details
			http.Error(w, "Failed to save asset", http.StatusInternalServerError)
			return
		}

		// Audit log success
		if a.standardMiddlewares != nil && a.standardMiddlewares.Audit != nil {
			args := map[string]string{
				"skill": skillName,
				"path":  assetPath,
				"size":  fmt.Sprintf("%d", len(body)),
			}
			argsBytes, _ := json.Marshal(args)
			entry := audit.Entry{
				ToolName:  "SkillAssetUpload",
				Arguments: json.RawMessage(argsBytes),
				Result:    "success",
			}
			_ = a.standardMiddlewares.Audit.Log(r.Context(), entry)
		}

		w.WriteHeader(http.StatusOK)
	}
}
