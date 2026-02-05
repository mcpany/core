package app

import (
	"io"
	"net/http"
	"strings"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/validation"
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
