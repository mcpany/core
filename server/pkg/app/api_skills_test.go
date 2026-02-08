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

func TestHandleUploadSkillAsset(t *testing.T) {
	manager, tmpDir := setupSkillManagerForHTTPTest(t)
	app := &Application{
		SkillManager: manager,
	}
	handler := app.handleUploadSkillAsset()

	// Helper to create a skill
	createSkill := func(name string) {
		err := manager.CreateSkill(&skill.Skill{
			Frontmatter: skill.Frontmatter{Name: name},
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

func (e *skillErrorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}
