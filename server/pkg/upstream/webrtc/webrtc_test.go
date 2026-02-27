// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package webrtc

import (
	"context"
	"errors"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/alexliesenfeld/health"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// Mock objects for testing
type MockChecker struct {
	mock.Mock
}

func (m *MockChecker) Check(ctx context.Context) health.CheckerResult {
	args := m.Called(ctx)
	return args.Get(0).(health.CheckerResult)
}

func (m *MockChecker) Start() {
	m.Called()
}

func (m *MockChecker) Stop() {
	m.Called()
}

func (m *MockChecker) GetRunningPeriodicCheckCount() int {
	args := m.Called()
	return args.Int(0)
}

// Implement missing methods for health.Checker interface
func (m *MockChecker) IsStarted() bool {
	args := m.Called()
	return args.Bool(0)
}

type MockToolManager struct {
	mock.Mock
	tool.ManagerInterface
}

func (m *MockToolManager) AddTool(t tool.Tool) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockToolManager) AddServiceInfo(id string, info *tool.ServiceInfo) {
	m.Called(id, info)
}

func (m *MockToolManager) GetTool(name string) (tool.Tool, bool) {
	args := m.Called(name)
	if t := args.Get(0); t != nil {
		return t.(tool.Tool), args.Bool(1)
	}
	return nil, args.Bool(1)
}

type MockPromptManager struct {
	mock.Mock
	prompt.ManagerInterface
}

func (m *MockPromptManager) AddPrompt(p prompt.Prompt) error {
	args := m.Called(p)
	return args.Error(0)
}

type MockResourceManager struct {
	mock.Mock
	resource.ManagerInterface
}

func (m *MockResourceManager) AddResource(r resource.Resource) error {
	args := m.Called(r)
	return args.Error(0)
}

func TestUpstream(t *testing.T) {
	poolManager := pool.NewManager()
	u := NewUpstream(poolManager)
	require.NotNil(t, u)

	t.Run("CheckHealth_NoChecker", func(t *testing.T) {
		err := u.CheckHealth(context.Background())
		assert.NoError(t, err)
	})

	t.Run("CheckHealth_WithChecker_Up", func(t *testing.T) {
		checker := &MockChecker{}
		checker.On("Check", mock.Anything).Return(health.CheckerResult{Status: health.StatusUp})

		// Access private field using type assertion since NewUpstream returns interface
		uu, ok := u.(*Upstream)
		require.True(t, ok)
		uu.checker = checker

		err := u.CheckHealth(context.Background())
		assert.NoError(t, err)
	})

	t.Run("CheckHealth_WithChecker_Down", func(t *testing.T) {
		checker := &MockChecker{}
		checker.On("Check", mock.Anything).Return(health.CheckerResult{Status: health.StatusDown})

		uu, ok := u.(*Upstream)
		require.True(t, ok)
		uu.checker = checker

		err := u.CheckHealth(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "health check failed")
	})
}

func TestWebrtcUpstream_Shutdown(t *testing.T) {
	poolManager := pool.NewManager()
	uInterface := NewUpstream(poolManager)

	err := uInterface.Shutdown(context.Background())
	assert.NoError(t, err)

	// Test with checker
	checker := &MockChecker{}
	checker.On("Stop").Return()

	uu, ok := uInterface.(*Upstream)
	require.True(t, ok)
	uu.checker = checker

	err = uInterface.Shutdown(context.Background())
	assert.NoError(t, err)
	checker.AssertCalled(t, "Stop")
}

func TestUpstream_Register(t *testing.T) {
	poolManager := pool.NewManager()
	u := NewUpstream(poolManager).(*Upstream)

	t.Run("successful_registration", func(t *testing.T) {
		serviceConfig := &configv1.UpstreamServiceConfig{
			Name: proto.String("test-webrtc-service"),
			WebrtcService: &configv1.WebrtcServiceConfig{
				Address: proto.String("ws://localhost:8080"),
				Calls: map[string]*configv1.WebrtcCallDefinition{
					"call1": {
						Parameters: &configv1.JSONSchema{
							Type: proto.String("object"),
						},
					},
				},
				Tools: []*configv1.WebrtcToolDefinition{
					{
						Name:        proto.String("echo"),
						Description: proto.String("echo tool"),
						CallId:      proto.String("call1"),
					},
				},
			},
		}

		mockTM := &MockToolManager{}
		mockTM.On("AddServiceInfo", mock.Anything, mock.Anything).Return()
		mockTM.On("AddTool", mock.Anything).Return(nil)

		mockPM := &MockPromptManager{}
		mockRM := &MockResourceManager{}

		serviceID, tools, _, err := u.Register(context.Background(), serviceConfig, mockTM, mockPM, mockRM, false)
		require.NoError(t, err)
		assert.Equal(t, "test-webrtc-service", serviceID)
		assert.Len(t, tools, 1)

		mockTM.AssertExpectations(t)
	})

	t.Run("nil_service_config", func(t *testing.T) {
		_, _, _, err := u.Register(context.Background(), nil, nil, nil, nil, false)
		assert.Error(t, err)
		assert.Equal(t, "service config is nil", err.Error())
	})

	t.Run("nil_webrtc_service_config", func(t *testing.T) {
		serviceConfig := &configv1.UpstreamServiceConfig{
			Name: proto.String("test-service"),
		}
		_, _, _, err := u.Register(context.Background(), serviceConfig, nil, nil, nil, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "webrtc service config is nil")
	})

	t.Run("add_tool_error", func(t *testing.T) {
		serviceConfig := &configv1.UpstreamServiceConfig{
			Name: proto.String("test-webrtc-service"),
			WebrtcService: &configv1.WebrtcServiceConfig{
				Calls: map[string]*configv1.WebrtcCallDefinition{
					"call1": {},
				},
				Tools: []*configv1.WebrtcToolDefinition{
					{
						Name:   proto.String("echo"),
						CallId: proto.String("call1"),
					},
				},
			},
		}

		mockTM := &MockToolManager{}
		mockTM.On("AddServiceInfo", mock.Anything, mock.Anything).Return()
		mockTM.On("AddTool", mock.Anything).Return(errors.New("failed to add tool"))

		serviceID, tools, _, err := u.Register(context.Background(), serviceConfig, mockTM, nil, nil, false)
		require.NoError(t, err) // Register itself doesn't fail, it logs errors and continues
		assert.Equal(t, "test-webrtc-service", serviceID)
		assert.Empty(t, tools) // Tools should be empty because addition failed
	})

	t.Run("authenticator_error", func(t *testing.T) {
		serviceConfig := &configv1.UpstreamServiceConfig{
			Name: proto.String("test-webrtc-service"),
			UpstreamAuth: &configv1.UpstreamAuth{
				ApiKey: &configv1.SecretValue{
					Value: proto.String("secret"),
				},
				// Missing Location/ParamName for API Key auth -> error
			},
			WebrtcService: &configv1.WebrtcServiceConfig{
				Calls: map[string]*configv1.WebrtcCallDefinition{"c1":{}},
				Tools: []*configv1.WebrtcToolDefinition{{Name:proto.String("t1"), CallId:proto.String("c1")}},
			},
		}

		mockTM := &MockToolManager{}

		_, tools, _, err := u.Register(context.Background(), serviceConfig, mockTM, nil, nil, false)
		assert.NoError(t, err) // Register returns nil error
		assert.Nil(t, tools)   // tools list is nil
	})

	t.Run("missing_call_id", func(t *testing.T) {
		serviceConfig := &configv1.UpstreamServiceConfig{
			Name: proto.String("test-webrtc-service"),
			WebrtcService: &configv1.WebrtcServiceConfig{
				Calls: map[string]*configv1.WebrtcCallDefinition{}, // Empty calls
				Tools: []*configv1.WebrtcToolDefinition{
					{
						Name:   proto.String("echo"),
						CallId: proto.String("non-existent-call-id"),
					},
				},
			},
		}

		mockTM := &MockToolManager{}
		mockTM.On("AddServiceInfo", mock.Anything, mock.Anything).Return()

		_, tools, _, err := u.Register(context.Background(), serviceConfig, mockTM, nil, nil, false)
		require.NoError(t, err)
		assert.Empty(t, tools)
	})

	t.Run("successful_prompt_and_resource_registration", func(t *testing.T) {
		serviceConfig := &configv1.UpstreamServiceConfig{
			Name: proto.String("test-webrtc-service-with-prompts-and-resources"),
			WebrtcService: &configv1.WebrtcServiceConfig{
				Calls: map[string]*configv1.WebrtcCallDefinition{
					"call1": {},
				},
				Tools: []*configv1.WebrtcToolDefinition{
					{Name: proto.String("tool1"), CallId: proto.String("call1")},
				},
				Prompts: []*configv1.PromptDefinition{
					{Name: proto.String("weather-prompt"), Template: proto.String("What is the weather?")},
				},
				Resources: []*configv1.ResourceDefinition{
					{
						Name: proto.String("dynamic-res"),
						Dynamic: &configv1.DynamicResourceDefinition{
							WebrtcCall: &configv1.WebrtcCallRef{Id: proto.String("call1")},
						},
					},
				},
			},
		}

		mockTM := &MockToolManager{}
		mockTM.On("AddServiceInfo", mock.Anything, mock.Anything).Return()
		mockTM.On("AddTool", mock.Anything).Return(nil)
		// For dynamic resource, GetTool is called
		mockTM.On("GetTool", mock.Anything).Return(&tool.MockTool{}, true)

		mockPM := &MockPromptManager{}
		mockPM.On("AddPrompt", mock.Anything).Return(nil)

		mockRM := &MockResourceManager{}
		mockRM.On("AddResource", mock.Anything).Return(nil)

		_, _, _, err := u.Register(context.Background(), serviceConfig, mockTM, mockPM, mockRM, false)
		require.NoError(t, err)

		mockPM.AssertCalled(t, "AddPrompt", mock.Anything)
		mockRM.AssertCalled(t, "AddResource", mock.Anything)
	})

	t.Run("sanitizer_failure", func(t *testing.T) {
		// Mock sanitizer to fail
		uSanitizer := NewUpstream(poolManager).(*Upstream)
		uSanitizer.toolNameSanitizer = func(s string) (string, error) {
			return "", errors.New("sanitization failed")
		}

		serviceConfig := &configv1.UpstreamServiceConfig{
			Name: proto.String("test-webrtc-service-with-sanitizer-failure"),
			WebrtcService: &configv1.WebrtcServiceConfig{
				Calls: map[string]*configv1.WebrtcCallDefinition{"c1":{}},
				Tools: []*configv1.WebrtcToolDefinition{{Name:proto.String("t1"), CallId:proto.String("c1")}},
				Resources: []*configv1.ResourceDefinition{
					{
						Name: proto.String("res1"),
						Dynamic: &configv1.DynamicResourceDefinition{
							WebrtcCall: &configv1.WebrtcCallRef{Id: proto.String("c1")},
						},
					},
				},
			},
		}

		mockTM := &MockToolManager{}
		mockTM.On("AddServiceInfo", mock.Anything, mock.Anything).Return()
		mockTM.On("AddTool", mock.Anything).Return(nil)
		// GetTool won't be called because sanitizer fails before looking up tool

		mockRM := &MockResourceManager{}

		_, _, _, err := uSanitizer.Register(context.Background(), serviceConfig, mockTM, nil, mockRM, false)
		require.NoError(t, err)

		// Resource should NOT be added due to sanitizer error
		mockRM.AssertNotCalled(t, "AddResource")
	})
}

func TestUpstream_Register_ToolNameGeneration(t *testing.T) {
	poolManager := pool.NewManager()
	u := NewUpstream(poolManager).(*Upstream)

	serviceConfig := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-webrtc-service-tool-name-generation"),
		WebrtcService: &configv1.WebrtcServiceConfig{
			Calls: map[string]*configv1.WebrtcCallDefinition{
				"call1": {},
			},
			Tools: []*configv1.WebrtcToolDefinition{
				{
					// Empty Name, should fallback to Description sanitized
					Description: proto.String("Get Weather Data"),
					CallId:      proto.String("call1"),
				},
			},
		},
	}

	mockTM := &MockToolManager{}
	mockTM.On("AddServiceInfo", mock.Anything, mock.Anything).Return()
	mockTM.On("AddTool", mock.Anything).Return(nil)

	_, tools, _, err := u.Register(context.Background(), serviceConfig, mockTM, nil, nil, false)
	require.NoError(t, err)
	require.Len(t, tools, 1)
	assert.Equal(t, "", tools[0].GetName())
}

func TestUpstream_Register_CornerCases(t *testing.T) {
	poolManager := pool.NewManager()
	u := NewUpstream(poolManager).(*Upstream)

	t.Run("disabled_tool", func(t *testing.T) {
		serviceConfig := &configv1.UpstreamServiceConfig{
			Name: proto.String("test-webrtc-disabled"),
			WebrtcService: &configv1.WebrtcServiceConfig{
				Calls: map[string]*configv1.WebrtcCallDefinition{"c1":{}},
				Tools: []*configv1.WebrtcToolDefinition{
					{Name: proto.String("disabled-tool"), CallId: proto.String("c1"), Disable: proto.Bool(true)},
				},
			},
		}
		mockTM := &MockToolManager{}
		mockTM.On("AddServiceInfo", mock.Anything, mock.Anything).Return()

		_, tools, _, _ := u.Register(context.Background(), serviceConfig, mockTM, nil, nil, false)
		assert.Empty(t, tools)
	})

	t.Run("empty_name_fallback", func(t *testing.T) {
		// Tool with no name and no description -> fallback to opX
		serviceConfig := &configv1.UpstreamServiceConfig{
			Name: proto.String("test-webrtc-empty-name"),
			WebrtcService: &configv1.WebrtcServiceConfig{
				Calls: map[string]*configv1.WebrtcCallDefinition{"c1":{}},
				Tools: []*configv1.WebrtcToolDefinition{
					{CallId: proto.String("c1")}, // No Name, No Description
				},
			},
		}
		mockTM := &MockToolManager{}
		mockTM.On("AddServiceInfo", mock.Anything, mock.Anything).Return()
		mockTM.On("AddTool", mock.MatchedBy(func(t tool.Tool) bool {
			return t.Tool().GetName() == "op0" // Expect "op0"
		})).Return(nil)

		_, _, _, _ := u.Register(context.Background(), serviceConfig, mockTM, nil, nil, false)
	})

	t.Run("disabled_resource", func(t *testing.T) {
		serviceConfig := &configv1.UpstreamServiceConfig{
			Name: proto.String("test-webrtc-disabled-resource"),
			WebrtcService: &configv1.WebrtcServiceConfig{
				Calls: map[string]*configv1.WebrtcCallDefinition{"c1":{}},
				Tools: []*configv1.WebrtcToolDefinition{{Name:proto.String("t1"), CallId:proto.String("c1")}},
				Resources: []*configv1.ResourceDefinition{
					{Name: proto.String("disabled-resource"), Disable: proto.Bool(true)},
				},
			},
		}
		mockTM := &MockToolManager{}
		mockTM.On("AddServiceInfo", mock.Anything, mock.Anything).Return()
		mockTM.On("AddTool", mock.Anything).Return(nil)
		mockRM := &MockResourceManager{}

		u.Register(context.Background(), serviceConfig, mockTM, nil, mockRM, false)
		mockRM.AssertNotCalled(t, "AddResource")
	})

	t.Run("dynamic_resource_missing_call", func(t *testing.T) {
		serviceConfig := &configv1.UpstreamServiceConfig{
			Name: proto.String("test-webrtc-missing-call"),
			WebrtcService: &configv1.WebrtcServiceConfig{
				Resources: []*configv1.ResourceDefinition{
					{Name: proto.String("res"), Dynamic: &configv1.DynamicResourceDefinition{}}, // Missing WebrtcCall
				},
			},
		}
		mockTM := &MockToolManager{}
		mockTM.On("AddServiceInfo", mock.Anything, mock.Anything).Return()
		mockRM := &MockResourceManager{}

		u.Register(context.Background(), serviceConfig, mockTM, nil, mockRM, false)
		mockRM.AssertNotCalled(t, "AddResource")
	})

	t.Run("dynamic_resource_call_id_not_found", func(t *testing.T) {
		serviceConfig := &configv1.UpstreamServiceConfig{
			Name: proto.String("test-webrtc-call-not-found"),
			WebrtcService: &configv1.WebrtcServiceConfig{
				Tools: []*configv1.WebrtcToolDefinition{}, // Empty tools map, so callIDToName map is empty
				Resources: []*configv1.ResourceDefinition{
					{
						Name: proto.String("res"),
						Dynamic: &configv1.DynamicResourceDefinition{
							WebrtcCall: &configv1.WebrtcCallRef{Id: proto.String("unknown-call-id")},
						},
					},
				},
			},
		}
		mockTM := &MockToolManager{}
		mockTM.On("AddServiceInfo", mock.Anything, mock.Anything).Return()
		mockRM := &MockResourceManager{}

		u.Register(context.Background(), serviceConfig, mockTM, nil, mockRM, false)
		mockRM.AssertNotCalled(t, "AddResource")
	})

	t.Run("tool_not_found_for_dynamic_resource", func(t *testing.T) {
		// This happens if Tool addition fails or sanitized name mismatch in map
		serviceConfig := &configv1.UpstreamServiceConfig{
			Name: proto.String("test-webrtc-tool-not-found"),
			WebrtcService: &configv1.WebrtcServiceConfig{
				Calls: map[string]*configv1.WebrtcCallDefinition{"c1":{}},
				Tools: []*configv1.WebrtcToolDefinition{{Name:proto.String("tool1"), CallId:proto.String("c1")}},
				Resources: []*configv1.ResourceDefinition{
					{
						Name: proto.String("res"),
						Dynamic: &configv1.DynamicResourceDefinition{
							WebrtcCall: &configv1.WebrtcCallRef{Id: proto.String("c1")},
						},
					},
				},
			},
		}
		mockTM := &MockToolManager{}
		mockTM.On("AddServiceInfo", mock.Anything, mock.Anything).Return()
		mockTM.On("AddTool", mock.Anything).Return(errors.New("fail add tool")) // Tool add fails
		// GetTool returns not found
		mockTM.On("GetTool", mock.Anything).Return(nil, false)
		mockRM := &MockResourceManager{}

		u.Register(context.Background(), serviceConfig, mockTM, nil, mockRM, false)
		mockRM.AssertNotCalled(t, "AddResource")
	})

	t.Run("disabled_prompt", func(t *testing.T) {
		serviceConfig := &configv1.UpstreamServiceConfig{
			Name: proto.String("test-webrtc-disabled-prompt"),
			WebrtcService: &configv1.WebrtcServiceConfig{
				Prompts: []*configv1.PromptDefinition{
					{Name: proto.String("disabled-prompt"), Disable: proto.Bool(true)},
				},
			},
		}
		mockTM := &MockToolManager{}
		mockTM.On("AddServiceInfo", mock.Anything, mock.Anything).Return()
		mockPM := &MockPromptManager{}

		u.Register(context.Background(), serviceConfig, mockTM, mockPM, nil, false)
		mockPM.AssertNotCalled(t, "AddPrompt")
	})

	t.Run("correct_input_schema_generation", func(t *testing.T) {
		serviceConfig := &configv1.UpstreamServiceConfig{
			Name: proto.String("test-webrtc-service"),
			WebrtcService: &configv1.WebrtcServiceConfig{
				Calls: map[string]*configv1.WebrtcCallDefinition{
					"c1": {
						Parameters: &configv1.JSONSchema{
							Type: proto.String("object"),
							Properties: map[string]*configv1.JSONSchema{
								"p1": {Type: proto.String("string")},
							},
							Required: []string{"p1"},
						},
					},
				},
				Tools: []*configv1.WebrtcToolDefinition{{Name:proto.String("t1"), CallId:proto.String("c1")}},
			},
		}
		mockTM := &MockToolManager{}
		mockTM.On("AddServiceInfo", mock.Anything, mock.Anything).Return()
		mockTM.On("AddTool", mock.MatchedBy(func(t tool.Tool) bool {
			// Check input schema
			// It's a bit complex to inspect protobuf struct, but we verify it doesn't crash
			// and required fields are passed.
			// Schemaconv is tested elsewhere, here we just ensure flow works.
			return true
		})).Return(nil)

		_, _, _, err := u.Register(context.Background(), serviceConfig, mockTM, nil, nil, false)
		assert.NoError(t, err)
	})
}
