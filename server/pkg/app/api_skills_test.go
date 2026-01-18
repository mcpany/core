// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mcpany/core/server/pkg/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleUploadSkillAsset(t *testing.T) {
	// Setup
	tmpDir, err := os.MkdirTemp("", "skills-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	skillManager, err := skill.NewManager(tmpDir)
	require.NoError(t, err)

	// Create a skill first so we can upload asset to it
	// SaveAsset checks if skill exists
	err = os.Mkdir(tmpDir+"/myskill", 0755)
	require.NoError(t, err)

	app := &Application{
		SkillManager: skillManager,
	}

	handler := app.handleUploadSkillAsset()

	// 1. Valid upload
	t.Run("Valid upload", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "/api/v1/skills/myskill/assets?path=test.txt", bytes.NewBufferString("hello world"))
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		// Verify file content
		content, err := os.ReadFile(tmpDir + "/myskill/test.txt")
		require.NoError(t, err)
		assert.Equal(t, "hello world", string(content))
	})

	// 2. Invalid method
	t.Run("Invalid method", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/api/v1/skills/myskill/assets?path=test.txt", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	// 3. Missing path param
	t.Run("Missing path param", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "/api/v1/skills/myskill/assets", bytes.NewBufferString("hello"))
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	// 4. Invalid URL format
	t.Run("Invalid URL format", func(t *testing.T) {
		// Path too short
		req, err := http.NewRequest(http.MethodPost, "/api/v1/skills/assets", bytes.NewBufferString("hello"))
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	// 5. Skill not found
	t.Run("Skill not found", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "/api/v1/skills/otherskill/assets?path=test.txt", bytes.NewBufferString("hello"))
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code) // SaveAsset returns error
	})
}
