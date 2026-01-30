// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/server/pkg/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleUploadSkillAsset(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	skillManager, err := skill.NewManager(tempDir)
	require.NoError(t, err)

	app := &Application{
		SkillManager: skillManager,
	}

	// Create a dummy skill to attach assets to
	testSkill := &skill.Skill{
		Frontmatter: skill.Frontmatter{
			Name: "test-skill",
		},
		Instructions: "Do nothing",
	}
	err = skillManager.CreateSkill(testSkill)
	require.NoError(t, err)

	handler := app.handleUploadSkillAsset()

	t.Run("Happy Path", func(t *testing.T) {
		body := []byte("print('hello world')")
		// The handler expects /api/v1/skills/{name}/assets
		req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/test-skill/assets?path=scripts/main.py", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		// Verify file content
		content, err := os.ReadFile(filepath.Join(tempDir, "test-skill", "scripts", "main.py"))
		require.NoError(t, err)
		assert.Equal(t, body, content)
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/skills/test-skill/assets?path=scripts/main.py", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	})

	t.Run("Invalid URL Format", func(t *testing.T) {
		// Missing 'assets' segment
		req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/test-skill/invalid", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Missing Path Query Param", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/test-skill/assets", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Insecure Path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/test-skill/assets?path=../secret.txt", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Skill Not Found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/unknown-skill/assets?path=test.py", bytes.NewReader([]byte("data")))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		// SaveAsset returns error if skill not found. Handler logs error and returns 500.
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
