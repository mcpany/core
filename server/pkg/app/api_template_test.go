// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestHandleTemplates_MethodNotAllowed(t *testing.T) {
	app := &Application{TemplateManager: NewTemplateManager(t.TempDir())}
	req := httptest.NewRequest(http.MethodPut, "/templates", nil)
	w := httptest.NewRecorder()
	app.handleTemplates().ServeHTTP(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleTemplates_Create_MalformedJSON(t *testing.T) {
	app := &Application{TemplateManager: NewTemplateManager(t.TempDir())}
	req := httptest.NewRequest(http.MethodPost, "/templates", bytes.NewReader([]byte("invalid")))
	w := httptest.NewRecorder()
	app.handleTemplates().ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleTemplates_Create_SaveError(t *testing.T) {
	// Create a directory named "templates.json" so it can't create the file
	dir := t.TempDir()
	path := filepath.Join(dir, "templates.json")
	os.MkdirAll(path, 0755)

	app := &Application{TemplateManager: NewTemplateManager(dir)}

	tmpl := &configv1.UpstreamServiceConfig{Name: proto.String("test")}
	body, _ := protojson.Marshal(tmpl)
	req := httptest.NewRequest(http.MethodPost, "/templates", bytes.NewReader(body))
	w := httptest.NewRecorder()
	app.handleTemplates().ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleTemplateDetail_MethodNotAllowed(t *testing.T) {
	app := &Application{TemplateManager: NewTemplateManager(t.TempDir())}
	req := httptest.NewRequest(http.MethodGet, "/templates/123", nil)
	w := httptest.NewRecorder()
	app.handleTemplateDetail().ServeHTTP(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleTemplateDetail_Delete_Success(t *testing.T) {
	app := &Application{TemplateManager: NewTemplateManager(t.TempDir())}
	// Pre-populate
	app.TemplateManager.SaveTemplate(&configv1.UpstreamServiceConfig{
		Id:   proto.String("123"),
		Name: proto.String("test"),
	})

	req := httptest.NewRequest(http.MethodDelete, "/templates/123", nil)
	w := httptest.NewRecorder()
	app.handleTemplateDetail().ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify deletion
	assert.Empty(t, app.TemplateManager.ListTemplates())
}

func TestHandleTemplateDetail_Delete_Error(t *testing.T) {
	dir := t.TempDir()
	app := &Application{TemplateManager: NewTemplateManager(dir)}

	// Pre-populate
	app.TemplateManager.SaveTemplate(&configv1.UpstreamServiceConfig{
		Id:   proto.String("123"),
		Name: proto.String("test"),
	})

	// Induce error by replacing file with directory
	path := filepath.Join(dir, "templates.json")
	os.Remove(path)
	os.Mkdir(path, 0755)

	req := httptest.NewRequest(http.MethodDelete, "/templates/123", nil)
	w := httptest.NewRecorder()
	app.handleTemplateDetail().ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleTemplateDetail_MissingID(t *testing.T) {
	app := &Application{TemplateManager: NewTemplateManager(t.TempDir())}
	req := httptest.NewRequest(http.MethodDelete, "/templates/", nil)
	w := httptest.NewRecorder()
	app.handleTemplateDetail().ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
