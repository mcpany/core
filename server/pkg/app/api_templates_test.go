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

	// Seed manual template
	tmpl := &configv1.UpstreamServiceConfig{}
	tmpl.SetName("Test Template")
	tmpl.SetId("tmpl-1")
	require.NoError(t, tm.SaveTemplate(tmpl))

	app := NewApplication()
	app.TemplateManager = tm
	handler := app.handleTemplates()

	req := httptest.NewRequest(http.MethodGet, "/templates", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var list []json.RawMessage
	err := json.Unmarshal(w.Body.Bytes(), &list)
	require.NoError(t, err)
	assert.NotEmpty(t, list)

	found := false
	for _, raw := range list {
		var tObj configv1.UpstreamServiceConfig
		err := protojson.Unmarshal(raw, &tObj)
		require.NoError(t, err)
		if tObj.GetId() == "tmpl-1" {
			found = true
			assert.Equal(t, "Test Template", tObj.GetName())
		}
	}
	assert.True(t, found, "Expected template not found")
}

func TestHandleTemplates_Create(t *testing.T) {
	tempDir := t.TempDir()
	tm := NewTemplateManager(tempDir)
	app := NewApplication()
	app.TemplateManager = tm
	handler := app.handleTemplates()

	tmpl := &configv1.UpstreamServiceConfig{}
	tmpl.SetName("New Template")
	// ID empty, should be generated

	opts := protojson.MarshalOptions{UseProtoNames: true}
	bodyBytes, err := opts.Marshal(tmpl)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/templates", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp["id"])

	// Verify saved
	savedList := tm.ListTemplates()
	found := false
	for _, s := range savedList {
		if s.GetId() == resp["id"] {
			found = true
			assert.Equal(t, "New Template", s.GetName())
		}
	}
	assert.True(t, found)
}

func TestHandleTemplates_Create_Invalid(t *testing.T) {
	tempDir := t.TempDir()
	tm := NewTemplateManager(tempDir)
	app := NewApplication()
	app.TemplateManager = tm
	handler := app.handleTemplates()

	req := httptest.NewRequest(http.MethodPost, "/templates", bytes.NewReader([]byte("invalid-json")))
	w := httptest.NewRecorder()
	handler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleTemplateDetail_Delete(t *testing.T) {
	tempDir := t.TempDir()
	tm := NewTemplateManager(tempDir)
	// Seed
	tmpl := &configv1.UpstreamServiceConfig{}
	tmpl.SetName("To Delete")
	tmpl.SetId("del-1")
	require.NoError(t, tm.SaveTemplate(tmpl))

	app := NewApplication()
	app.TemplateManager = tm
	handler := app.handleTemplateDetail()

	req := httptest.NewRequest(http.MethodDelete, "/templates/del-1", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify deleted
	savedList := tm.ListTemplates()
	found := false
	for _, s := range savedList {
		if s.GetId() == "del-1" {
			found = true
		}
	}
	assert.False(t, found)
}

func TestHandleTemplateDetail_Delete_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	tm := NewTemplateManager(tempDir)
	app := NewApplication()
	app.TemplateManager = tm
	handler := app.handleTemplateDetail()

	req := httptest.NewRequest(http.MethodDelete, "/templates/unknown-id", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestHandleTemplateDetail_MissingID(t *testing.T) {
	app := NewApplication()
	handler := app.handleTemplateDetail()

	req := httptest.NewRequest(http.MethodDelete, "/templates/", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleTemplates_InvalidMethod(t *testing.T) {
	app := NewApplication()
	handler := app.handleTemplates()
	req := httptest.NewRequest(http.MethodPut, "/templates", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleTemplateDetail_InvalidMethod(t *testing.T) {
	app := NewApplication()
	handler := app.handleTemplateDetail()
	req := httptest.NewRequest(http.MethodGet, "/templates/123", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}
