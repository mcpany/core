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
	"google.golang.org/protobuf/proto"
)

func TestHandleTemplates_Detailed(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "templates-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	app := NewApplication()
	app.TemplateManager = NewTemplateManager(tmpDir)
	handler := app.handleTemplates()

	t.Run("List Templates Empty", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/templates", nil)
		w := httptest.NewRecorder()
		handler(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, "[]", w.Body.String())
	})

	t.Run("Create Template", func(t *testing.T) {
		body := `{"id": "tpl1", "name": "My Template"}`
		req := httptest.NewRequest(http.MethodPost, "/templates", bytes.NewReader([]byte(body)))
		w := httptest.NewRecorder()
		handler(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("List Templates Populated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/templates", nil)
		w := httptest.NewRecorder()
		handler(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var list []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &list)
		require.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, "tpl1", list[0]["id"])
	})

	t.Run("Create Invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/templates", bytes.NewReader([]byte("{invalid")))
		w := httptest.NewRecorder()
		handler(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/templates", nil)
		w := httptest.NewRecorder()
		handler(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestHandleTemplateDetail(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "templates-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	app := NewApplication()
	app.TemplateManager = NewTemplateManager(tmpDir)

	// Create a template
	tmpl := &configv1.UpstreamServiceConfig{
		Id:   proto.String("tpl1"),
		Name: proto.String("My Template"),
	}
	app.TemplateManager.SaveTemplate(tmpl)

	handler := app.handleTemplateDetail()

	t.Run("Delete Template", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/templates/tpl1", nil)
		w := httptest.NewRecorder()
		handler(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify gone
		list := app.TemplateManager.ListTemplates()
		assert.Len(t, list, 0)
	})

	t.Run("Delete Missing ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/templates/", nil)
		w := httptest.NewRecorder()
		handler(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/templates/tpl1", nil)
		w := httptest.NewRecorder()
		handler(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}
