// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
)

func TestIsUnsafeConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   *configv1.UpstreamServiceConfig
		isUnsafe bool
	}{
		{
			name: "Safe HTTP Service",
			config: &configv1.UpstreamServiceConfig{
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{},
				},
			},
			isUnsafe: false,
		},
		{
			name: "Unsafe Command Line Service",
			config: &configv1.UpstreamServiceConfig{
				ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
					CommandLineService: &configv1.CommandLineUpstreamService{},
				},
			},
			isUnsafe: true,
		},
		{
			name: "Unsafe MCP Stdio Service",
			config: &configv1.UpstreamServiceConfig{
				ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
					McpService: &configv1.McpUpstreamService{
						ConnectionType: &configv1.McpUpstreamService_StdioConnection{
							StdioConnection: &configv1.McpStdioConnection{},
						},
					},
				},
			},
			isUnsafe: true,
		},
		{
			name: "Unsafe MCP Bundle Service",
			config: &configv1.UpstreamServiceConfig{
				ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
					McpService: &configv1.McpUpstreamService{
						ConnectionType: &configv1.McpUpstreamService_BundleConnection{
							BundleConnection: &configv1.McpBundleConnection{},
						},
					},
				},
			},
			isUnsafe: true,
		},
		{
			name: "Safe MCP HTTP Service",
			config: &configv1.UpstreamServiceConfig{
				ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
					McpService: &configv1.McpUpstreamService{
						ConnectionType: &configv1.McpUpstreamService_HttpConnection{
							HttpConnection: &configv1.McpStreamableHttpConnection{},
						},
					},
				},
			},
			isUnsafe: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isUnsafe, isUnsafeConfig(tt.config))
		})
	}
}

func TestHandleServiceStatus_Mocked(t *testing.T) {
	store := memory.NewStore()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockToolManager := tool.NewMockManagerInterface(ctrl)
	app := &Application{
		ToolManager: mockToolManager,
	}

	// Setup: Add a service to the store
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
	}
	require.NoError(t, store.SaveService(context.Background(), svc))

	t.Run("Status Inactive", func(t *testing.T) {
		mockToolManager.EXPECT().ListServices().Return(nil)

		req := httptest.NewRequest(http.MethodGet, "/services/test-service/status", nil)
		w := httptest.NewRecorder()

		app.handleServiceStatus(w, req, "test-service", store)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "Inactive", resp["status"])
	})

	t.Run("Status Active", func(t *testing.T) {
		mockToolManager.EXPECT().ListServices().Return([]*tool.ServiceInfo{
			{Name: "test-service"},
		})

		req := httptest.NewRequest(http.MethodGet, "/services/test-service/status", nil)
		w := httptest.NewRecorder()

		app.handleServiceStatus(w, req, "test-service", store)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "Active", resp["status"])
	})

	t.Run("Service Not Found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/services/unknown-service/status", nil)
		w := httptest.NewRecorder()

		app.handleServiceStatus(w, req, "unknown-service", store)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
