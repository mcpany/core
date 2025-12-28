// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package webhook_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/webhook"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockToolManager is a mock for tool.ManagerInterface.
type MockToolManager struct {
	mock.Mock
}

func (m *MockToolManager) AddTool(t tool.Tool) error {
	args := m.Called(t)
	return args.Error(0)
}
func (m *MockToolManager) GetTool(name string) (tool.Tool, bool) {
	args := m.Called(name)
	return args.Get(0).(tool.Tool), args.Bool(1)
}
func (m *MockToolManager) ListTools() []tool.Tool {
	args := m.Called()
	return args.Get(0).([]tool.Tool)
}
func (m *MockToolManager) ExecuteTool(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}
func (m *MockToolManager) AddMiddleware(middleware tool.ExecutionMiddleware) {
	m.Called(middleware)
}
func (m *MockToolManager) SetMCPServer(s tool.MCPServerProvider) {
	m.Called(s)
}
func (m *MockToolManager) SetProfiles(profiles []string, definitions []*configv1.ProfileDefinition) {
	m.Called(profiles, definitions)
}
func (m *MockToolManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	args := m.Called(serviceID)
	return args.Get(0).(*tool.ServiceInfo), args.Bool(1)
}
func (m *MockToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	m.Called(serviceID, info)
}
func (m *MockToolManager) ClearToolsForService(serviceID string) {
	m.Called(serviceID)
}
func (m *MockToolManager) ListServices() []*tool.ServiceInfo {
	args := m.Called()
	return args.Get(0).([]*tool.ServiceInfo)
}

func TestWebhookManager(t *testing.T) {
	mockToolManager := new(MockToolManager)
	mgr := webhook.NewManager(mockToolManager)

	configs := []*configv1.SystemWebhookConfig{
		{
			UrlPath:     "/webhooks/github",
			Secret:      "gh-secret",
			Action:      &configv1.SystemWebhookConfig_TriggerTool{TriggerTool: "github-tool"},
			Name:        "github",
			Description: "GitHub Webhook",
		},
		{
			UrlPath:     "/webhooks/stripe",
			Action:      &configv1.SystemWebhookConfig_TriggerTool{TriggerTool: "stripe-tool"},
			Name:        "stripe",
			Description: "Stripe Webhook",
		},
	}
	mgr.UpdateConfig(configs)

	handler := mgr.Handler()
	server := httptest.NewServer(handler)
	defer server.Close()

	t.Run("Valid Webhook with Secret", func(t *testing.T) {
		payload := map[string]string{"event": "push"}
		body, _ := json.Marshal(payload)

		mockToolManager.On("ExecuteTool", mock.Anything, mock.MatchedBy(func(req *tool.ExecutionRequest) bool {
			return req.ToolName == "github-tool" && string(req.ToolInputs) == string(body)
		})).Return("success", nil).Once()

		req, _ := http.NewRequest(http.MethodPost, server.URL+"/webhooks/github", bytes.NewReader(body))
		req.Header.Set("X-Webhook-Secret", "gh-secret")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		mockToolManager.AssertExpectations(t)
	})

	t.Run("Invalid Secret", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, server.URL+"/webhooks/github", nil)
		req.Header.Set("X-Webhook-Secret", "wrong-secret")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Valid Webhook without Secret", func(t *testing.T) {
		mockToolManager.On("ExecuteTool", mock.Anything, mock.MatchedBy(func(req *tool.ExecutionRequest) bool {
			return req.ToolName == "stripe-tool"
		})).Return("success", nil).Once()

		req, _ := http.NewRequest(http.MethodPost, server.URL+"/webhooks/stripe", bytes.NewBufferString("{}"))

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		mockToolManager.AssertExpectations(t)
	})

	t.Run("Unknown Webhook", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, server.URL+"/webhooks/unknown", nil)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
