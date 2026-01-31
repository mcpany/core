// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestHandleTemplates_Get(t *testing.T) {
	tmpDir := t.TempDir()
	tm := NewTemplateManager(tmpDir)
	app := &Application{
		TemplateManager: tm,
	}

	// Seed a template
	tmpl := &configv1.UpstreamServiceConfig{}
	tmpl.SetId("test-id")
	tmpl.SetName("Test Template")

	err := tm.SaveTemplate(tmpl)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/templates", nil)
	w := httptest.NewRecorder()

	handler := app.handleTemplates()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var list []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&list)
	require.NoError(t, err)

	// BuiltinTemplates are seeded by NewTemplateManager.
	// We added one more.
	assert.True(t, len(list) >= 1)

	// Verify our seeded template is present
	found := false
	for _, item := range list {
		if id, ok := item["id"].(string); ok && id == "test-id" {
			found = true
			assert.Equal(t, "Test Template", item["name"])
			break
		}
	}
	assert.True(t, found, "Seeded template not found")
}

func TestHandleTemplates_Post(t *testing.T) {
	tmpDir := t.TempDir()
	tm := NewTemplateManager(tmpDir)
	app := &Application{
		TemplateManager: tm,
	}

	newTmpl := &configv1.UpstreamServiceConfig{}
	newTmpl.SetName("New Template")
	newTmpl.SetConfigurationSchema("{}")

	body, err := protojson.Marshal(newTmpl)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/templates", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler := app.handleTemplates()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var resMap map[string]string
	err = json.NewDecoder(resp.Body).Decode(&resMap)
	require.NoError(t, err)
	assert.NotEmpty(t, resMap["id"])

	// Check via ListTemplates
	all := tm.ListTemplates()
	foundTmpl := false
	for _, item := range all {
		if item.GetId() == resMap["id"] {
			foundTmpl = true
			assert.Equal(t, "New Template", item.GetName())
			break
		}
	}
	assert.True(t, foundTmpl)
}

func TestHandleTemplates_Post_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	tm := NewTemplateManager(tmpDir)
	app := &Application{
		TemplateManager: tm,
	}

	req := httptest.NewRequest(http.MethodPost, "/templates", bytes.NewReader([]byte("{invalid-json")))
	w := httptest.NewRecorder()

	handler := app.handleTemplates()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandleTemplateDetail_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	tm := NewTemplateManager(tmpDir)
	app := &Application{
		TemplateManager: tm,
	}

	// Seed a template
	tmpl := &configv1.UpstreamServiceConfig{}
	tmpl.SetId("delete-me")
	tmpl.SetName("To Delete")

	err := tm.SaveTemplate(tmpl)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete, "/templates/delete-me", nil)
	w := httptest.NewRecorder()

	handler := app.handleTemplateDetail()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify it's gone
	all := tm.ListTemplates()
	for _, item := range all {
		assert.NotEqual(t, "delete-me", item.GetId())
	}
}

func TestHandleTemplateDetail_MissingID(t *testing.T) {
	tmpDir := t.TempDir()
	tm := NewTemplateManager(tmpDir)
	app := &Application{
		TemplateManager: tm,
	}

	req := httptest.NewRequest(http.MethodDelete, "/templates/", nil)
	w := httptest.NewRecorder()

	handler := app.handleTemplateDetail()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
