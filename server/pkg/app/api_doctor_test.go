// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHandleServiceDiagnose(t *testing.T) {
	app, store := setupApiTestApp()
	handler := app.handleServiceDiagnose(store)

	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String("http://127.0.0.1:8080"),
			},
		},
		Id: proto.String(uuid.New().String()),
	}
	_ = store.SaveService(context.Background(), svc)

	t.Run("Diagnose Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/services/test-service/diagnose", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var result map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "test-service", result["ServiceName"])
		// Status might be ERROR because 127.0.0.1:8080 likely isn't running, but we check if it ran at all
		assert.NotEmpty(t, result["Status"])
	})

	t.Run("Diagnose NotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/services/non-existent/diagnose", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/services/test-service/diagnose", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}
