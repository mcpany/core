// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/pkg/skill"
	"github.com/stretchr/testify/assert"
)

func TestHandleUploadSkillAsset_DoSProtection(t *testing.T) {
	tmpDir := t.TempDir()
	sm, _ := skill.NewManager(tmpDir)
	app := &Application{SkillManager: sm}

	testSkill := &skill.Skill{
		Frontmatter: skill.Frontmatter{Name: "test-skill"},
		Instructions: "Run this.",
	}
	sm.CreateSkill(testSkill)

	t.Run("Body Too Large", func(t *testing.T) {
		// 10MB + 1 byte
		size := 10*1024*1024 + 1
		// We use a dummy reader to simulate large body without allocating all of it if possible,
		// but httptest.NewRequest reads it? No, it takes io.Reader.

		// To avoid OOM in test runner if it's constrained, we can use a repeating reader?
		// But 10MB is small enough for modern machines.
		largeBody := make([]byte, size)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/test-skill/assets?path=large.txt", bytes.NewReader(largeBody))
		w := httptest.NewRecorder()

		app.handleUploadSkillAsset().ServeHTTP(w, req)

		assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
		assert.Contains(t, w.Body.String(), "Request body too large")
	})
}
