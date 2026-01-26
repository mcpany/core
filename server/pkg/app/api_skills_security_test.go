// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/audit"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type mockAuditStore struct {
	entries []audit.Entry
}

func (m *mockAuditStore) Write(ctx context.Context, entry audit.Entry) error {
	m.entries = append(m.entries, entry)
	return nil
}

func (m *mockAuditStore) Read(ctx context.Context, filter audit.Filter) ([]audit.Entry, error) {
	return m.entries, nil
}

func (m *mockAuditStore) Close() error {
	return nil
}

func TestHandleUploadSkillAsset_Security(t *testing.T) {
	// Setup Skill Manager
	tempDir, err := os.MkdirTemp("", "app-skills-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	sm, err := skill.NewManager(tempDir)
	require.NoError(t, err)
	err = sm.CreateSkill(&skill.Skill{Frontmatter: skill.Frontmatter{Name: "test-skill"}})
	require.NoError(t, err)

	// Setup Audit Middleware
	auditConfig := &configv1.AuditConfig{
		Enabled: proto.Bool(true),
	}
	am, err := middleware.NewAuditMiddleware(auditConfig)
	require.NoError(t, err)
	mockStore := &mockAuditStore{}
	am.SetStore(mockStore)

	// Setup Application
	app := &Application{
		SkillManager: sm,
		standardMiddlewares: &middleware.StandardMiddlewares{
			Audit: am,
		},
	}

	t.Run("BodySizeLimit", func(t *testing.T) {
		// Create large body (11MB)
		largeBody := make([]byte, 11*1024*1024)
		req := httptest.NewRequest("POST", "/api/v1/skills/test-skill/assets?path=large.bin", bytes.NewReader(largeBody))
		w := httptest.NewRecorder()

		handler := app.handleUploadSkillAsset()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	})

	t.Run("AuditLogging", func(t *testing.T) {
		body := []byte("test content")
		req := httptest.NewRequest("POST", "/api/v1/skills/test-skill/assets?path=test.txt", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler := app.handleUploadSkillAsset()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Check audit logs
		require.Len(t, mockStore.entries, 1)
		assert.Equal(t, "SkillAssetUpload", mockStore.entries[0].ToolName)
		assert.Contains(t, string(mockStore.entries[0].Arguments), "test-skill")
		assert.Equal(t, "success", mockStore.entries[0].Result)
	})
}
