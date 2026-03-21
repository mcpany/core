// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHandleUploadSkillAsset_PathDisclosure(t *testing.T) {
	// Create a directory
	tmpDir := t.TempDir()

	// Create manager with this root
	sm, err := skill.NewManager(tmpDir)
	require.NoError(t, err)

	// Create a skill
	skillName := "leak-skill"
	err = sm.CreateSkill(&skill.Skill{Frontmatter: skill.Frontmatter{Name: skillName}, Instructions: "inst"})
	require.NoError(t, err)

	// Now make the skill directory read-only to cause SaveAsset to fail
	skillDir := filepath.Join(tmpDir, skillName)
	// Make it 0500 (read-only, executable for traversal)
	err = os.Chmod(skillDir, 0500)
	require.NoError(t, err)

	// Ensure we can clean up
	t.Cleanup(func() {
		_ = os.Chmod(skillDir, 0755)
	})

	app := &Application{SkillManager: sm}

	// Try to upload an asset
	// Path must match: /api/v1/skills/{name}/assets
	req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/"+skillName+"/assets?path=test.txt", bytes.NewReader([]byte("data")))
	w := httptest.NewRecorder()
	app.handleUploadSkillAsset().ServeHTTP(w, req)

	// We expect 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Check for leak
	// The error likely contains the full path if not handled correctly
	if strings.Contains(w.Body.String(), tmpDir) {
		t.Errorf("Security Leak! Response body contains server path: %s", w.Body.String())
	} else {
		t.Logf("Response body is safe: %s", w.Body.String())
	}

	assert.Contains(t, w.Body.String(), "Failed to save asset")
}

func TestHandleSecrets_RBAC_Enforcement(t *testing.T) {
	app, store := setupApiTestApp()

	// Add test secrets
	secret := configv1.Secret_builder{Id: proto.String("test-secret"), Name: proto.String("test")}.Build()
	require.NoError(t, store.SaveSecret(context.Background(), secret))

	// Setup handler directly with RBAC middleware simulating api.go
	handler := app.createAPIHandler(store)

	t.Run("Non-Admin user cannot access /api/v1/secrets", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/secrets", nil)
		ctx := auth.ContextWithUser(req.Context(), "victim-user")
		ctx = auth.ContextWithRoles(ctx, []string{"user"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Admin user can access /api/v1/secrets", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/secrets", nil)
		ctx := auth.ContextWithUser(req.Context(), "admin-user")
		ctx = auth.ContextWithRoles(ctx, []string{"admin"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
