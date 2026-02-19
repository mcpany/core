// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/skill"
	"github.com/mcpany/core/server/pkg/validation"
)

// POST /api/v1/skills.
func (a *Application) handleSkills() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			skills, err := a.SkillManager.ListSkills()
			if err != nil {
				logging.GetLogger().Error("Failed to list skills", "error", err)
				http.Error(w, "Failed to list skills", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(skills); err != nil {
				logging.GetLogger().Error("Failed to encode skills", "error", err)
			}
		case http.MethodPost:
			var s skill.Skill
			// Limit body size
			r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB limit for metadata/instructions
			if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			if err := a.SkillManager.CreateSkill(&s); err != nil {
				logging.GetLogger().Warn("Failed to create skill", "name", s.Name, "error", err)
				http.Error(w, fmt.Sprintf("Failed to create skill: %v", err), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(s)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// DELETE /api/v1/skills/{name}.
func (a *Application) handleSkillDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract skill name from URL path
		// Path expected: /skills/{name} (relative to mux mount)
		// We need to be careful about strict path matching in mux vs here.
		// If mounted as /skills/, then TrimPrefix("/skills/") should work.

		// Check if this is an asset request
		if strings.HasSuffix(r.URL.Path, "/assets") {
			a.handleUploadSkillAsset().ServeHTTP(w, r)
			return
		}

		name := strings.TrimPrefix(r.URL.Path, "/skills/")
		if name == "" || name == "/" {
			http.Error(w, "Skill name required", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			s, err := a.SkillManager.GetSkill(name)
			if err != nil {
				http.Error(w, "Skill not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(s)
		case http.MethodPut:
			var s skill.Skill
			r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
			if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			// Use the name from URL as the source of truth for the update target
			// The body name might be different if we allow renaming (UpdateSkill handles it)
			if err := a.SkillManager.UpdateSkill(name, &s); err != nil {
				logging.GetLogger().Warn("Failed to update skill", "name", name, "error", err)
				http.Error(w, fmt.Sprintf("Failed to update skill: %v", err), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(s)
		case http.MethodDelete:
			if err := a.SkillManager.DeleteSkill(name); err != nil {
				logging.GetLogger().Warn("Failed to delete skill", "name", name, "error", err)
				http.Error(w, fmt.Sprintf("Failed to delete skill: %v", err), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

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
		// Path expected: /skills/{name}/assets (relative to mux mount)
		// r.URL.Path will be /skills/{name}/assets
		// Handle both /skills/{name}/assets and /api/v1/skills/{name}/assets
		path := strings.TrimPrefix(r.URL.Path, "/api/v1")

		// Path expected: /skills/{name}/assets
		parts := strings.Split(path, "/")
		// parts[0] = ""
		// parts[1] = "skills"
		// parts[2] = "{name}"
		// parts[3] = "assets"
		if len(parts) < 4 || parts[1] != "skills" || parts[3] != "assets" {
			http.Error(w, "Invalid URL format\n", http.StatusBadRequest)
			return
		}
		skillName := parts[2]

		assetPath := r.URL.Query().Get("path")
		if assetPath == "" {
			http.Error(w, "Asset path query parameter required", http.StatusBadRequest)
			return
		}

		if err := validation.IsSecureRelativePath(assetPath); err != nil {
			http.Error(w, "Invalid asset path", http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		if err := a.SkillManager.SaveAsset(skillName, assetPath, body); err != nil {
			logging.GetLogger().Error("Failed to save asset", "skill", skillName, "path", assetPath, "error", err)
			// Sentinel Security: Return generic error to avoid leaking path information or internal details
			http.Error(w, "Failed to save asset", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
