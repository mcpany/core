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

func TestHandleTemplates_List(t *testing.T) {
	tempDir := t.TempDir()
	tm := NewTemplateManager(tempDir)
	app := &Application{TemplateManager: tm}

	// Create a dummy template
	tmpl := &configv1.UpstreamServiceConfig{}
	tmpl.SetId("tmpl-1")
	tmpl.SetName("Template 1")
	require.NoError(t, tm.SaveTemplate(tmpl))

	req := httptest.NewRequest(http.MethodGet, "/templates", nil)
	w := httptest.NewRecorder()

	app.handleTemplates()(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var templates []map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &templates)
	require.NoError(t, err)

	// BuiltinTemplates are also listed by TemplateManager if it uses them as fallback or seed?
	// TemplateManager implementation reads from disk AND `BuiltinTemplates`.
	// Let's assume seeded ones are present.
	// But `NewTemplateManager` uses `dir`.
	// `ListTemplates` logic:
	// It lists files in dir.
	// Does it include builtins?
	// Let's assume only what's in dir + persistence.
	// Actually `seeds.go` says `BuiltinTemplates` are variables.
	// `TemplateManager` likely merges them.
	// If so, we should see `tmpl-1` plus builtins.

	found := false
	for _, t := range templates {
		if t["id"] == "tmpl-1" {
			found = true
			break
		}
	}
	assert.True(t, found, "Created template should be in list")
}

func TestHandleTemplates_Create(t *testing.T) {
	tempDir := t.TempDir()
	tm := NewTemplateManager(tempDir)
	app := &Application{TemplateManager: tm}

	tmpl := &configv1.UpstreamServiceConfig{}
	tmpl.SetName("New Template")
	httpSvc := &configv1.HttpUpstreamService{}
	httpSvc.SetAddress("http://localhost:8080")
	tmpl.SetHttpService(httpSvc)

	opts := protojson.MarshalOptions{UseProtoNames: true}
	body, _ := opts.Marshal(tmpl)

	req := httptest.NewRequest(http.MethodPost, "/templates", bytes.NewReader(body))
	w := httptest.NewRecorder()

	app.handleTemplates()(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp["id"])

	// Verify persistence
	list := tm.ListTemplates()
	found := false
	for _, t := range list {
		if t.GetName() == "New Template" {
			found = true
			break
		}
	}
	assert.True(t, found, "New template should be persisted")
}

func TestHandleTemplates_Create_InvalidJSON(t *testing.T) {
	app := &Application{}
	req := httptest.NewRequest(http.MethodPost, "/templates", bytes.NewReader([]byte("invalid")))
	w := httptest.NewRecorder()

	app.handleTemplates()(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleTemplates_MethodNotAllowed(t *testing.T) {
	app := &Application{}
	req := httptest.NewRequest(http.MethodPut, "/templates", nil)
	w := httptest.NewRecorder()

	app.handleTemplates()(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleTemplateDetail_Delete(t *testing.T) {
	tempDir := t.TempDir()
	tm := NewTemplateManager(tempDir)
	app := &Application{TemplateManager: tm}

	// Create a dummy template
	tmpl := &configv1.UpstreamServiceConfig{}
	tmpl.SetId("tmpl-delete")
	tmpl.SetName("Template Delete")
	require.NoError(t, tm.SaveTemplate(tmpl))

	req := httptest.NewRequest(http.MethodDelete, "/templates/tmpl-delete", nil)
	w := httptest.NewRecorder()

	app.handleTemplateDetail()(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify it's gone
	for _, tmpl := range tm.ListTemplates() {
		if tmpl.GetId() == "tmpl-delete" {
			assert.Fail(t, "Template should have been deleted")
		}
	}
}

func TestHandleTemplateDetail_MethodNotAllowed(t *testing.T) {
	app := &Application{}
	req := httptest.NewRequest(http.MethodGet, "/templates/id", nil)
	w := httptest.NewRecorder()

	app.handleTemplateDetail()(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleTemplateDetail_MissingID(t *testing.T) {
	app := &Application{}
	req := httptest.NewRequest(http.MethodDelete, "/templates/", nil)
	w := httptest.NewRecorder()

	app.handleTemplateDetail()(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
