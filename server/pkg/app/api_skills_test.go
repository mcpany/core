// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mcpany/core/server/pkg/skill"
)

func TestHandleUploadSkillAsset_Limit(t *testing.T) {
	// Mock SkillManager is needed, or we just rely on validation failure if possible?
	// But we want to test the body reading limit which happens BEFORE SkillManager usage (mostly).
	// Actually, the body read happens before SaveAsset.
	// So if we send a large body, it should fail at ReadAll/MaxBytesReader.

	// Setup application with minimal dependencies
	app := &Application{
		SkillManager: &skill.Manager{}, // Dummy manager, we expect to fail before usage or we mock it if needed
	}

	// Create a large body > 10MB
	largeBody := make([]byte, 11*1024*1024) // 11MB
	req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/test-skill/assets?path=test.png", bytes.NewReader(largeBody))
	w := httptest.NewRecorder()

	// Handler
	handler := app.handleUploadSkillAsset()
	handler.ServeHTTP(w, req)

	// Expectation: 400 Bad Request (due to body too large)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	// Check body for "request body too large" or similar standard error from http.MaxBytesReader?
	// http.MaxBytesReader sets error on Read. io.ReadAll returns it.
	// The handler checks err != nil and returns "Failed to read body".
	if !strings.Contains(w.Body.String(), "Failed to read body") {
		t.Errorf("Expected error message 'Failed to read body', got %q", w.Body.String())
	}
}

func TestHandleUploadSkillAsset_Success(t *testing.T) {
	// Setup temporary directory for skills
	tmpDir := t.TempDir()
	sm, err := skill.NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create skill manager: %v", err)
	}

	// Create a skill first
	sk := &skill.Skill{
		Frontmatter: skill.Frontmatter{
			Name: "test-skill",
		},
	}
	if err := sm.CreateSkill(sk); err != nil {
		t.Fatalf("Failed to create skill: %v", err)
	}

	app := &Application{
		SkillManager: sm,
	}

	// Create a small body
	body := []byte("test content")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/test-skill/assets?path=test.txt", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Handler
	handler := app.handleUploadSkillAsset()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}
