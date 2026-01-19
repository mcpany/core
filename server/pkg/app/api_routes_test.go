// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestAPIRouteVersioning(t *testing.T) {
	store := memory.NewStore()
	app := &Application{
		SettingsManager: NewGlobalSettingsManager("", nil, nil),
	}
	// We rely on fallback to store in handleListServices if Registry is nil

	// Create a service in store so we can list it.
	svc := &configv1.UpstreamServiceConfig{
		Id:   proto.String(uuid.New().String()),
		Name: proto.String("test-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String("http://localhost:8080"),
			},
		},
	}
	require.NoError(t, store.SaveService(context.Background(), svc))

	handler := app.createAPIHandler(store)

	tests := []struct {
		name         string
		path         string
		method       string
		expectedCode int
	}{
		{
			name:         "Legacy Services List",
			path:         "/services",
			method:       http.MethodGet,
			expectedCode: http.StatusOK,
		},
		{
			name:         "V1 Services List",
			path:         "/api/v1/services",
			method:       http.MethodGet,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Legacy Service Detail",
			path:         "/services/test-service",
			method:       http.MethodGet,
			expectedCode: http.StatusOK,
		},
		{
			name:         "V1 Service Detail",
			path:         "/api/v1/services/test-service",
			method:       http.MethodGet,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Health Check",
			path:         "/health",
			method:       http.MethodGet,
			expectedCode: http.StatusOK,
		},
		{
			name:         "V1 System Status",
			path:         "/api/v1/system/status",
			method:       http.MethodGet,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Legacy System Status",
			path:         "/system/status",
			method:       http.MethodGet,
			expectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedCode, w.Code, "path: %s", tt.path)
		})
	}
}
