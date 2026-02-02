// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestHandleTemplates(t *testing.T) {
	// Setup temp dir for TemplateManager
	tempDir, err := os.MkdirTemp("", "mcp-templates-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tm := NewTemplateManager(tempDir)
	app := &Application{
		TemplateManager: tm,
	}

	t.Run("Create Template", func(t *testing.T) {
		tmpl := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test-template"),
			Id:   proto.String("tpl-123"),
		}.Build()

		opts := protojson.MarshalOptions{UseProtoNames: true}
		body, err := opts.Marshal(tmpl)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/templates", bytes.NewReader(body))
		w := httptest.NewRecorder()

		app.handleTemplates()(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "tpl-123", resp["id"])

		// Verify it was saved
		saved := tm.ListTemplates()
		found := false
		for _, s := range saved {
			if s.GetId() == "tpl-123" {
				found = true
				assert.Equal(t, "test-template", s.GetName())
				break
			}
		}
		assert.True(t, found, "Template not found in manager")
	})

	t.Run("List Templates", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/templates", nil)
		w := httptest.NewRecorder()

		app.handleTemplates()(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Check response
		var list []json.RawMessage
		err := json.Unmarshal(w.Body.Bytes(), &list)
		require.NoError(t, err)
		// We expect at least the one we created + builtins
		assert.GreaterOrEqual(t, len(list), 1)

		found := false
		for _, item := range list {
			var tmpl configv1.UpstreamServiceConfig
			err = protojson.Unmarshal(item, &tmpl)
			require.NoError(t, err)
			if tmpl.GetId() == "tpl-123" {
				found = true
				break
			}
		}
		assert.True(t, found, "Created template not found in list response")
	})

	t.Run("Create Template with Auto ID", func(t *testing.T) {
		tmpl := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("auto-id-template"),
		}.Build()

		opts := protojson.MarshalOptions{UseProtoNames: true}
		body, err := opts.Marshal(tmpl)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/templates", bytes.NewReader(body))
		w := httptest.NewRecorder()

		app.handleTemplates()(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.NotEmpty(t, resp["id"])
	})

	t.Run("Create Invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/templates", bytes.NewBufferString("{invalid"))
		w := httptest.NewRecorder()

		app.handleTemplates()(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/templates", nil)
		w := httptest.NewRecorder()

		app.handleTemplates()(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestHandleTemplateDetail(t *testing.T) {
	// Setup temp dir for TemplateManager
	tempDir, err := os.MkdirTemp("", "mcp-templates-detail-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tm := NewTemplateManager(tempDir)
	app := &Application{
		TemplateManager: tm,
	}

	// Seed a template
	tmpl := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("to-delete"),
		Id:   proto.String("del-123"),
	}.Build()
	require.NoError(t, tm.SaveTemplate(tmpl))

	t.Run("Delete Template", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/templates/del-123", nil)
		w := httptest.NewRecorder()

		app.handleTemplateDetail()(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify deletion
		saved := tm.ListTemplates()
		found := false
		for _, s := range saved {
			if s.GetId() == "del-123" {
				found = true
				break
			}
		}
		assert.False(t, found, "Template should have been deleted")
	})

	t.Run("Delete Missing ID", func(t *testing.T) {
		// The handler extracts ID from URL path trimming prefix "/templates/"
		// If we pass just "/templates/", id is empty.
		req := httptest.NewRequest(http.MethodDelete, "/templates/", nil)
		w := httptest.NewRecorder()

		app.handleTemplateDetail()(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/templates/123", nil)
		w := httptest.NewRecorder()

		app.handleTemplateDetail()(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}
