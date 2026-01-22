// Copyright 2026 Author(s) of MCP Any
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
	"github.com/mcpany/core/server/pkg/doctor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHandleServiceDiagnose(t *testing.T) {
	app, store := setupApiTestApp()

	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String("http://127.0.0.1:54321"), // Random port likely closed
			},
		},
		Id: proto.String(uuid.New().String()),
	}
	require.NoError(t, store.SaveService(context.Background(), svc))

	t.Run("Diagnose Service", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/services/test-service/diagnose", nil)
		w := httptest.NewRecorder()

		// Route via handleServiceDetail which dispatches to handleServiceDiagnose
		handler := app.handleServiceDetail(store)
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var result map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &result)
		require.NoError(t, err)

		// We expect a failure because the port is closed
		assert.Equal(t, "test-service", result["service_name"])
		status := doctor.Status(result["status"].(string))
		assert.Contains(t, []doctor.Status{doctor.StatusError, doctor.StatusWarning}, status)
		if status != doctor.StatusOk {
			// Error might be present
			if errStr, ok := result["error"].(string); ok {
				assert.NotEmpty(t, errStr)
			}
		}
	})

	t.Run("Diagnose Service Not Found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/services/unknown-service/diagnose", nil)
		w := httptest.NewRecorder()

		handler := app.handleServiceDetail(store)
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/services/test-service/diagnose", nil)
		w := httptest.NewRecorder()

		handler := app.handleServiceDetail(store)
		handler.ServeHTTP(w, req)

		// handleServiceDetail might fall through or 405 depending on implementation?
		// handleServiceDiagnose checks method.
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}
