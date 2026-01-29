// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/pkg/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleUploadSkillAsset_Comprehensive(t *testing.T) {
	// Setup a temporary directory for skills
	tmpDir := t.TempDir()
	sm, err := skill.NewManager(tmpDir)
	require.NoError(t, err)

	app := &Application{SkillManager: sm}

	// Create a test skill to upload assets to
	testSkill := &skill.Skill{
		Frontmatter: skill.Frontmatter{Name: "test-skill"},
		Instructions: "Run this.",
	}
	require.NoError(t, sm.CreateSkill(testSkill))

	handler := app.handleUploadSkillAsset()

	t.Run("Valid Upload", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/test-skill/assets?path=scripts/test.py", bytes.NewReader([]byte("print('hello')")))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// Verify file exists
		assert.FileExists(t, tmpDir+"/test-skill/scripts/test.py")
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}
		for _, method := range methods {
			t.Run(method, func(t *testing.T) {
				req := httptest.NewRequest(method, "/api/v1/skills/test-skill/assets", nil)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)
				assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
			})
		}
	})

	t.Run("Invalid URL Format", func(t *testing.T) {
		// Missing skill name or assets segment
		invalidURLs := []string{
			"/api/v1/skills/assets",
			"/api/v1/skills/test-skill",
			"/api/v1/skills/test-skill/other",
		}
		for _, url := range invalidURLs {
			t.Run(url, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodPost, url, nil)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)
				assert.Equal(t, http.StatusBadRequest, w.Code)
			})
		}
	})

	t.Run("Missing Path Query Parameter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/test-skill/assets", bytes.NewReader([]byte("content")))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Asset path query parameter required")
	})

	t.Run("Insecure Path Traversal", func(t *testing.T) {
		insecurePaths := []struct {
			path string
			code int
		}{
			{"../secret.txt", http.StatusBadRequest},
			{"scripts/../../secret.txt", http.StatusBadRequest},
			// Absolute paths are caught by SkillManager, returning 500
			{"/etc/passwd", http.StatusInternalServerError},
		}
		for _, tc := range insecurePaths {
			t.Run(tc.path, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/test-skill/assets?path="+tc.path, bytes.NewReader([]byte("content")))
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)
				assert.Equal(t, tc.code, w.Code)
				if tc.code == http.StatusBadRequest {
					assert.Contains(t, w.Body.String(), "Invalid asset path")
				}
			})
		}
	})

	t.Run("Skill Not Found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/unknown-skill/assets?path=test.txt", bytes.NewReader([]byte("content")))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Body Read Error", func(t *testing.T) {
		// Use a custom error reader
		req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/test-skill/assets?path=test.txt", &errReader{})
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to read body")
	})

    t.Run("Empty Body", func(t *testing.T) {
        // Empty body should be allowed (creates empty file)
        req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/test-skill/assets?path=empty.txt", nil)
        w := httptest.NewRecorder()
        handler.ServeHTTP(w, req)
        assert.Equal(t, http.StatusOK, w.Code)
        assert.FileExists(t, tmpDir+"/test-skill/empty.txt")
    })
}

type errReader struct{}

func (e *errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("simulated read error")
}
