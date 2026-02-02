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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleServicesBulk(t *testing.T) {
	app, store := setupApiTestApp()

	handler := app.handleServicesBulk(store)

	// Seed some services
	svc1 := configv1.UpstreamServiceConfig_builder{Name: stringPtr("svc1"), Disable: boolPtr(false)}.Build()
	svc2 := configv1.UpstreamServiceConfig_builder{Name: stringPtr("svc2"), Disable: boolPtr(true)}.Build()
	require.NoError(t, store.SaveService(context.Background(), svc1))
	require.NoError(t, store.SaveService(context.Background(), svc2))

	t.Run("Bulk Delete", func(t *testing.T) {
		reqBody := BulkServiceActionRequest{
			Action:   "delete",
			Services: []string{"svc1"},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/services/bulk", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// Check deletion
		s1, err := store.GetService(context.Background(), "svc1")
		if err == nil {
			assert.Nil(t, s1)
		} else {
			assert.Error(t, err)
		}

		// Check svc2 still exists
		s2, err := store.GetService(context.Background(), "svc2")
		assert.NoError(t, err)
		assert.NotNil(t, s2)
	})

	t.Run("Bulk Enable", func(t *testing.T) {
		reqBody := BulkServiceActionRequest{
			Action:   "enable",
			Services: []string{"svc2"},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/services/bulk", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		s2, err := store.GetService(context.Background(), "svc2")
		assert.NoError(t, err)
		assert.False(t, s2.GetDisable())
	})

	t.Run("Bulk Disable", func(t *testing.T) {
		// svc2 is now enabled. Disable it.
		reqBody := BulkServiceActionRequest{
			Action:   "disable",
			Services: []string{"svc2"},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/services/bulk", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		s2, err := store.GetService(context.Background(), "svc2")
		assert.NoError(t, err)
		assert.True(t, s2.GetDisable())
	})

	t.Run("Bulk Restart", func(t *testing.T) {
		// Just checking it returns 200
		reqBody := BulkServiceActionRequest{
			Action:   "restart",
			Services: []string{"svc2"},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/services/bulk", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Invalid Action", func(t *testing.T) {
		reqBody := BulkServiceActionRequest{
			Action:   "invalid",
			Services: []string{"svc2"},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/services/bulk", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func boolPtr(b bool) *bool {
	return &b
}
