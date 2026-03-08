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

func TestHandleSkills_Get(t *testing.T) {
	manager, _ := setupSkillManagerForHTTPTest(t)
	app := &Application{SkillManager: manager}

	// Create a dummy skill
	err := manager.CreateSkill(&skill.Skill{Frontmatter: skill.Frontmatter{Name: "test-skill-get"}})
	require.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/skills", nil)
	rr := httptest.NewRecorder()

	app.handleSkills().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"name":"test-skill-get"`)
}

func TestHandleSkills_Post(t *testing.T) {
	manager, _ := setupSkillManagerForHTTPTest(t)
	app := &Application{SkillManager: manager}

	body := []byte(`{"name":"test-skill-post"}`)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/skills", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	app.handleSkills().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	// Verify it was created
	s, err := manager.GetSkill("test-skill-post")
	require.NoError(t, err)
	assert.Equal(t, "test-skill-post", s.Name)
}

func TestHandleSkillDetail_Get(t *testing.T) {
	manager, _ := setupSkillManagerForHTTPTest(t)
	app := &Application{SkillManager: manager}

	// Create a dummy skill
	err := manager.CreateSkill(&skill.Skill{Frontmatter: skill.Frontmatter{Name: "test-skill-detail-get"}})
	require.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, "/skills/test-skill-detail-get", nil)
	rr := httptest.NewRecorder()

	app.handleSkillDetail().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"name":"test-skill-detail-get"`)
}

func TestHandleSkillDetail_Put(t *testing.T) {
	manager, _ := setupSkillManagerForHTTPTest(t)
	app := &Application{SkillManager: manager}

	// Create a dummy skill
	err := manager.CreateSkill(&skill.Skill{Frontmatter: skill.Frontmatter{Name: "test-skill-detail-put"}})
	require.NoError(t, err)

	body := []byte(`{"name":"test-skill-detail-put", "description":"updated desc"}`)
	req, _ := http.NewRequest(http.MethodPut, "/skills/test-skill-detail-put", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	app.handleSkillDetail().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify it was updated
	s, err := manager.GetSkill("test-skill-detail-put")
	require.NoError(t, err)
	assert.Equal(t, "updated desc", s.Description)
}

func TestHandleSkillDetail_Delete(t *testing.T) {
	manager, _ := setupSkillManagerForHTTPTest(t)
	app := &Application{SkillManager: manager}

	// Create a dummy skill
	err := manager.CreateSkill(&skill.Skill{Frontmatter: skill.Frontmatter{Name: "test-skill-detail-del"}})
	require.NoError(t, err)

	req, _ := http.NewRequest(http.MethodDelete, "/skills/test-skill-detail-del", nil)
	rr := httptest.NewRecorder()

	app.handleSkillDetail().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)

	// Verify it was deleted
	_, err = manager.GetSkill("test-skill-detail-del")
	assert.Error(t, err) // Should not be found
}

func TestHandleSkills_Methods(t *testing.T) {
	manager, _ := setupSkillManagerForHTTPTest(t)
	app := &Application{SkillManager: manager}

	req, _ := http.NewRequest(http.MethodPut, "/api/v1/skills", nil)
	rr := httptest.NewRecorder()

	app.handleSkills().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleSkills_PostInvalid(t *testing.T) {
	manager, _ := setupSkillManagerForHTTPTest(t)
	app := &Application{SkillManager: manager}

	body := []byte(`{"name":`) // invalid json
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/skills", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	app.handleSkills().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleSkillDetail_NotFound(t *testing.T) {
	manager, _ := setupSkillManagerForHTTPTest(t)
	app := &Application{SkillManager: manager}

	req, _ := http.NewRequest(http.MethodGet, "/skills/not-found", nil)
	rr := httptest.NewRecorder()

	app.handleSkillDetail().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandleSkillDetail_MissingName(t *testing.T) {
	manager, _ := setupSkillManagerForHTTPTest(t)
	app := &Application{SkillManager: manager}

	req, _ := http.NewRequest(http.MethodGet, "/skills/", nil)
	rr := httptest.NewRecorder()

	app.handleSkillDetail().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleSkillDetail_Methods(t *testing.T) {
	manager, _ := setupSkillManagerForHTTPTest(t)
	app := &Application{SkillManager: manager}

	err := manager.CreateSkill(&skill.Skill{Frontmatter: skill.Frontmatter{Name: "test-skill-detail-methods"}})
	require.NoError(t, err)

	req, _ := http.NewRequest(http.MethodPatch, "/skills/test-skill-detail-methods", nil)
	rr := httptest.NewRecorder()

	app.handleSkillDetail().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleSkillDetail_PutInvalid(t *testing.T) {
	manager, _ := setupSkillManagerForHTTPTest(t)
	app := &Application{SkillManager: manager}

	err := manager.CreateSkill(&skill.Skill{Frontmatter: skill.Frontmatter{Name: "test-skill-detail-put-inv"}})
	require.NoError(t, err)

	body := []byte(`{"name":`) // invalid json
	req, _ := http.NewRequest(http.MethodPut, "/skills/test-skill-detail-put-inv", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	app.handleSkillDetail().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}


func TestHandleUploadSkillAsset_InvalidMethod(t *testing.T) {
	manager, _ := setupSkillManagerForHTTPTest(t)
	app := &Application{SkillManager: manager}

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/skills/test-skill/assets", nil)
	rr := httptest.NewRecorder()

	app.handleUploadSkillAsset().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleUploadSkillAsset_NoPath(t *testing.T) {
	manager, _ := setupSkillManagerForHTTPTest(t)
	app := &Application{SkillManager: manager}

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/skills/test-skill/assets", bytes.NewReader([]byte("test")))
	rr := httptest.NewRecorder()

	app.handleUploadSkillAsset().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleUploadSkillAsset_MissingSkill(t *testing.T) {
	manager, _ := setupSkillManagerForHTTPTest(t)
	app := &Application{SkillManager: manager}

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/skills/non-existent-skill/assets?path=test.txt", bytes.NewReader([]byte("test")))
	rr := httptest.NewRecorder()

	app.handleUploadSkillAsset().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
