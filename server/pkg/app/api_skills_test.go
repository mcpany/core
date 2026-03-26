// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/server/pkg/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSkillManagerForHTTPTest(t *testing.T) (*skill.Manager, string) {
	tmpDir, err := os.MkdirTemp("", "skill_http_test")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	manager, err := skill.NewManager(tmpDir)
	require.NoError(t, err)

	return manager, tmpDir
}

// TestHandleSkills ...
// Summary: TestHandleSkills
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	manager, _ := setupSkillManagerForHTTPTest(t)
	app := &Application{
		SkillManager: manager,
	}
	handler := app.handleSkills()

	createSkill := func(name string) {
		err := manager.CreateSkill(&skill.Skill{
			Frontmatter:  skill.Frontmatter{Name: name},
			Instructions: "Test instructions",
		})
		require.NoError(t, err)
	}

	createSkill("test-skill-1")
	createSkill("test-skill-2")

	tests := []struct {
		name         string
		method       string
		path         string
		body         string
		expectedCode int
		expectedBody []string
	}{
		{
			name:         "MethodGet",
			method:       http.MethodGet,
			path:         "/api/v1/skills",
			expectedCode: http.StatusOK,
			expectedBody: []string{"test-skill-1", "test-skill-2"},
		},
		{
			name:         "MethodPost_Success",
			method:       http.MethodPost,
			path:         "/api/v1/skills",
			body:         `{"name":"test-skill-3","instructions":"Test instructions"}`,
			expectedCode: http.StatusCreated,
			expectedBody: []string{"test-skill-3"},
		},
		{
			name:         "MethodPost_InvalidBody",
			method:       http.MethodPost,
			path:         "/api/v1/skills",
			body:         `invalid json`,
			expectedCode: http.StatusBadRequest,
			expectedBody: []string{"Invalid request body"},
		},
		{
			name:         "MethodPost_CreationError",
			method:       http.MethodPost,
			path:         "/api/v1/skills",
			body:         `{"name":"invalid/name","instructions":"Test instructions"}`,
			expectedCode: http.StatusBadRequest,
			expectedBody: []string{"Failed to create skill"},
		},
		{
			name:         "MethodNotAllowed",
			method:       http.MethodPut,
			path:         "/api/v1/skills",
			expectedCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			for _, expectedStr := range tt.expectedBody {
				assert.Contains(t, w.Body.String(), expectedStr)
			}
		})
	}
}

// TestHandleSkillDetail ...
// Summary: TestHandleSkillDetail
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	manager, _ := setupSkillManagerForHTTPTest(t)
	app := &Application{
		SkillManager: manager,
	}
	handler := app.handleSkillDetail()

	createSkill := func(name string) {
		err := manager.CreateSkill(&skill.Skill{
			Frontmatter:  skill.Frontmatter{Name: name},
			Instructions: "Test instructions",
		})
		require.NoError(t, err)
	}

	createSkill("test-skill-get")
	createSkill("test-skill-put")
	createSkill("test-skill-put-invalid")
	createSkill("test-skill-put-error")
	createSkill("test-skill-delete")
	createSkill("test-skill-not-allowed")
	createSkill("test-skill-asset-route")

	tests := []struct {
		name         string
		method       string
		path         string
		body         string
		expectedCode int
		expectedBody []string
	}{
		{
			name:         "MethodGet_Success",
			method:       http.MethodGet,
			path:         "/skills/test-skill-get",
			expectedCode: http.StatusOK,
			expectedBody: []string{"test-skill-get"},
		},
		{
			name:         "MethodGet_NotFound",
			method:       http.MethodGet,
			path:         "/skills/non-existent-skill",
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "MethodPut_Success",
			method:       http.MethodPut,
			path:         "/skills/test-skill-put",
			body:         `{"name":"test-skill-put","instructions":"Updated instructions"}`,
			expectedCode: http.StatusOK,
			expectedBody: []string{"Updated instructions"},
		},
		{
			name:         "MethodPut_InvalidBody",
			method:       http.MethodPut,
			path:         "/skills/test-skill-put-invalid",
			body:         `invalid json`,
			expectedCode: http.StatusBadRequest,
			expectedBody: []string{"Invalid request body"},
		},
		{
			name:         "MethodPut_UpdateError",
			method:       http.MethodPut,
			path:         "/skills/test-skill-put-error",
			body:         `{"name":"invalid/name","instructions":"Updated instructions"}`,
			expectedCode: http.StatusBadRequest,
			expectedBody: []string{"Failed to update skill"},
		},
		{
			name:         "MethodDelete_Success",
			method:       http.MethodDelete,
			path:         "/skills/test-skill-delete",
			expectedCode: http.StatusNoContent,
		},
		{
			name:         "MethodDelete_NotFound",
			method:       http.MethodDelete,
			path:         "/skills/non-existent-skill",
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "MethodNotAllowed",
			method:       http.MethodPost,
			path:         "/skills/test-skill-not-allowed",
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "MissingSkillName",
			method:       http.MethodGet,
			path:         "/skills/",
			expectedCode: http.StatusBadRequest,
			expectedBody: []string{"Skill name required"},
		},
		{
			name:         "AssetRouting",
			method:       http.MethodGet,
			path:         "/skills/test-skill-asset-route/assets?path=test.txt",
			expectedCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			for _, expectedStr := range tt.expectedBody {
				assert.Contains(t, w.Body.String(), expectedStr)
			}
		})
	}
}

// TestHandleUploadSkillAsset ...
// Summary: TestHandleUploadSkillAsset
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	manager, tmpDir := setupSkillManagerForHTTPTest(t)
	app := &Application{
		SkillManager: manager,
	}
	handler := app.handleUploadSkillAsset()

	// Helper to create a skill
	createSkill := func(name string) {
		err := manager.CreateSkill(&skill.Skill{
			Frontmatter:  skill.Frontmatter{Name: name},
			Instructions: "Test instructions",
		})
		require.NoError(t, err)
	}

	t.Run("Success", func(t *testing.T) {
		skillName := "test-skill-success"
		createSkill(skillName)

		body := &bytes.Buffer{}
		// Just raw bytes body, not multipart form as per implementation of handleUploadSkillAsset
		// Implementation: body, err := io.ReadAll(r.Body)
		content := []byte("asset content")
		body.Write(content)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/"+skillName+"/assets?path=test.txt", body)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify file created
		assetPath := filepath.Join(tmpDir, skillName, "test.txt")
		savedContent, err := os.ReadFile(assetPath)
		require.NoError(t, err)
		assert.Equal(t, content, savedContent)
	})

	t.Run("MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/skills/s1/assets?path=test.txt", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("BadRequest_MissingPath", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/s1/assets", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Asset path query parameter required")
	})

	t.Run("BadRequest_InvalidPath", func(t *testing.T) {
		// Path traversal attempt
		req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/s1/assets?path=../evil.txt", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid asset path")
	})

	t.Run("InternalServerError_SkillNotFound", func(t *testing.T) {
		// Skill does not exist
		req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/non-existent/assets?path=test.txt", bytes.NewReader([]byte("data")))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		// handleUploadSkillAsset returns 500 on SaveAsset error
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to save asset")
	})

	t.Run("BadRequest_BodyReadError", func(t *testing.T) {
		// Create a request with a body that fails to read
		req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/s1/assets?path=test.txt", &skillErrorReader{})
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to read body")
	})
}

type skillErrorReader struct{}

// Read ...
// Summary: Read
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return 0, io.ErrUnexpectedEOF
}
