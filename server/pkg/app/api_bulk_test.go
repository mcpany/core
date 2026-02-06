// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestHandleServicesBulk(t *testing.T) {
	// Setup
	store := memory.NewStore()
	app := NewApplication()
	app.Storage = store
	app.fs = afero.NewMemMapFs()
	app.configPaths = []string{}

	// Create valid dummy services
	svc1 := &configv1.UpstreamServiceConfig{}
	svc1.SetName("svc1")
	svc1.SetDisable(false)
	// Make svc1 safe too just in case
	http1 := &configv1.HttpUpstreamService{}
	http1.SetAddress("http://example.com/1")
	svc1.SetHttpService(http1)

	svc2 := &configv1.UpstreamServiceConfig{}
	svc2.SetName("svc2")
	svc2.SetDisable(true)
	// Use safe service to avoid admin role requirement for enable
	http2 := &configv1.HttpUpstreamService{}
	http2.SetAddress("http://example.com/2")
	svc2.SetHttpService(http2)

	_ = store.SaveService(context.Background(), svc1)
	_ = store.SaveService(context.Background(), svc2)

	handler := app.handleServicesBulk(store)

	t.Run("Enable Services", func(t *testing.T) {
		reqBody := BulkServiceActionRequest{
			Action:   "enable",
			Services: []string{"svc2"},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/services/bulk", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify store update
		updated, _ := store.GetService(context.Background(), "svc2")
		assert.False(t, updated.GetDisable())
	})

	t.Run("Disable Services", func(t *testing.T) {
		reqBody := BulkServiceActionRequest{
			Action:   "disable",
			Services: []string{"svc1"},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/services/bulk", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		updated, _ := store.GetService(context.Background(), "svc1")
		assert.True(t, updated.GetDisable())
	})

	t.Run("Delete Services", func(t *testing.T) {
		reqBody := BulkServiceActionRequest{
			Action:   "delete",
			Services: []string{"svc1"},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/services/bulk", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		deleted, _ := store.GetService(context.Background(), "svc1")
		assert.Nil(t, deleted)
	})

	t.Run("Invalid Action", func(t *testing.T) {
		reqBody := BulkServiceActionRequest{
			Action:   "explode",
			Services: []string{"svc2"},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/services/bulk", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
