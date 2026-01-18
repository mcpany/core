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
	// Setup temporary directory for skills
	tmpDir, err := os.MkdirTemp("", "skills-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Initialize SkillManager
	sm, err := skill.NewManager(tmpDir)
	require.NoError(t, err)

	// Initialize Application
	app := &Application{
		SkillManager: sm,
	}

	// Create a test skill
	testSkill := &skill.Skill{
		Frontmatter: skill.Frontmatter{
			Name:        "test-skill",
			Description: "A test skill",
		},
		Instructions: "Run this.",
	}
	err = sm.CreateSkill(testSkill)
	require.NoError(t, err)

	tests := []struct {
		name           string
		method         string
		url            string
		body           []byte
		expectedStatus int
	}{
		{
			name:           "Valid Upload",
			method:         http.MethodPost,
			url:            "/api/v1/skills/test-skill/assets?path=scripts/test.py",
			body:           []byte("print('hello')"),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Wrong Method",
			method:         http.MethodGet,
			url:            "/api/v1/skills/test-skill/assets?path=scripts/test.py",
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Invalid URL Format",
			method:         http.MethodPost,
			url:            "/api/v1/skills/assets", // Missing skill name
			body:           []byte("data"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing Path Param",
			method:         http.MethodPost,
			url:            "/api/v1/skills/test-skill/assets",
			body:           []byte("data"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Skill Not Found",
			method:         http.MethodPost,
			url:            "/api/v1/skills/unknown-skill/assets?path=test.txt",
			body:           []byte("data"),
			expectedStatus: http.StatusInternalServerError, // SaveAsset returns error for non-existent skill
		},
		{
			name:           "Invalid Path (Traversal)",
			method:         http.MethodPost,
			url:            "/api/v1/skills/test-skill/assets?path=../test.txt",
			body:           []byte("data"),
			expectedStatus: http.StatusInternalServerError, // SaveAsset validates path and returns error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, bytes.NewReader(tt.body))
			w := httptest.NewRecorder()

			// The handler expects manually parsed URL parts or router usage.
			// The implementation manually parses r.URL.Path assuming /api/v1/skills/{name}/assets
			// So we must use that structure.

			handler := app.handleUploadSkillAsset()
			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				// Verify file was written
				// "scripts/test.py"
				assetPath := filepath.Join(tmpDir, "test-skill", "scripts", "test.py")
				content, err := os.ReadFile(assetPath)
				require.NoError(t, err)
				assert.Equal(t, tt.body, content)
			}
		})
	}
}
