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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandleUploadSkillAsset_DoS verifies that the upload handler rejects overly large bodies.
func TestHandleUploadSkillAsset_DoS(t *testing.T) {
	manager, _ := setupSkillManagerForHTTPTest(t)
	app := &Application{
		SkillManager: manager,
	}
	handler := app.handleUploadSkillAsset()

	// Create a skill first so the request is valid otherwise
	skillName := "dos-skill"
	err := manager.CreateSkill(&skill.Skill{
		Frontmatter: skill.Frontmatter{Name: skillName},
		Instructions: "Test instructions",
	})
	require.NoError(t, err)

	// Create a large body (11MB)
	// We check for rejection of > 10MB
	size := 11 * 1024 * 1024
	// Use a dummy reader
	largeBody := bytes.NewReader(make([]byte, size))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/"+skillName+"/assets?path=large.txt", largeBody)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Expect 413 Payload Too Large
	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
}

// TestHandleUploadSkillAsset_PathParsing verifies robust path handling.
func TestHandleUploadSkillAsset_PathParsing(t *testing.T) {
	manager, _ := setupSkillManagerForHTTPTest(t)
	app := &Application{
		SkillManager: manager,
	}
	handler := app.handleUploadSkillAsset()

	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{
			name:           "Double Slash",
			path:           "/api/v1/skills//assets", // Should fail due to strict parsing
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing Skill Name",
			path:           "/api/v1/skills//assets",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Incorrect Segment",
			path:           "/api/v1/skills/s1/foo",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Too Short",
			path:           "/api/v1/skills",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.path+"?path=test.txt", strings.NewReader("data"))
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Path %q: got %d, expected %d", tt.path, w.Code, tt.expectedStatus)
			}
		})
	}
}
