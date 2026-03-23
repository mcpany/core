// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// mockToolManager to simulate errors
type mockToolManager struct {
	tool.ManagerInterface
	tools    map[string]tool.Tool
	addError error
	getError bool
}

func newMockToolManager() *mockToolManager {
	return &mockToolManager{
		tools: make(map[string]tool.Tool),
	}
}

func (m *mockToolManager) AddTool(t tool.Tool) error {
	if m.addError != nil {
		return m.addError
	}
	sanitizedToolName, err := util.SanitizeToolName(t.Tool().GetName())
	if err != nil {
		return err
	}
	m.tools[t.Tool().GetServiceId()+"."+sanitizedToolName] = t
	return nil
}

func (m *mockToolManager) ExecuteTool(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	return nil, nil
}

func (m *mockToolManager) SetMCPServer(_ tool.MCPServerProvider) {
}

func (m *mockToolManager) AddMiddleware(_ tool.ExecutionMiddleware) {
}

func (m *mockToolManager) GetTool(name string) (tool.Tool, bool) {
	if m.getError {
		return nil, false
	}
	t, ok := m.tools[name]
	return t, ok
}

func (m *mockToolManager) ListTools() []tool.Tool {
	tools := make([]tool.Tool, 0, len(m.tools))
	for _, t := range m.tools {
		tools = append(tools, t)
	}
	return tools
}

func (m *mockToolManager) ListServices() []*tool.ServiceInfo {
	return nil
}

func (m *mockToolManager) AddServiceInfo(_ string, _ *tool.ServiceInfo) {}

func (m *mockToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}

func TestNewStdioUpstream(t *testing.T) {
	u := NewUpstream()
	assert.NotNil(t, u)
	_, ok := u.(*Upstream)
	assert.True(t, ok)
	err := u.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestStdioUpstream_Shutdown(t *testing.T) {
	u := NewUpstream()
	// Shutdown without register (checker is nil)
	err := u.Shutdown(context.Background())
	assert.NoError(t, err)

	// Register then Shutdown
	tm := newMockToolManager()
	prm := prompt.NewManager()
	rm := resource.NewManager()
	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name:               proto.String("test-shutdown"),
		CommandLineService: configv1.CommandLineUpstreamService_builder{}.Build(),
	}.Build()

	_, _, _, err = u.Register(context.Background(), serviceConfig, tm, prm, rm, false)
	require.NoError(t, err)

	err = u.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestStdioUpstream_Register(t *testing.T) {
	tm := newMockToolManager()
	prm := prompt.NewManager()
	rm := resource.NewManager()
	u := NewUpstream()

	t.Run("successful registration", func(t *testing.T) {
		tm := newMockToolManager()
		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test-stdio-service"),
			CommandLineService: configv1.CommandLineUpstreamService_builder{
				Command: proto.String("/bin/echo"),
				Calls: map[string]*configv1.CommandLineCallDefinition{
					"echo-call": configv1.CommandLineCallDefinition_builder{
						Id: proto.String("echo-call"),
						Parameters: []*configv1.CommandLineParameterMapping{
							configv1.CommandLineParameterMapping_builder{
								Schema: configv1.ParameterSchema_builder{
									Name:        proto.String("args"),
									Type:        configv1.ParameterType_ARRAY.Enum(),
									Description: proto.String("Additional arguments"),
								}.Build(),
							}.Build(),
						},
					}.Build(),
				},
				Tools: []*configv1.ToolDefinition{
					configv1.ToolDefinition_builder{
						Name:   proto.String("echo"),
						CallId: proto.String("echo-call"),
					}.Build(),
				},
			}.Build(),
		}.Build()

		serviceID, discoveredTools, _, err := u.Register(
			context.Background(),
			serviceConfig,
			tm,
			prm,
			rm,
			false,
		)
		require.NoError(t, err)
		expectedKey, _ := util.SanitizeServiceName("test-stdio-service")
		assert.Equal(t, expectedKey, serviceID)
		assert.Len(t, discoveredTools, 1)
		assert.Equal(t, "echo", discoveredTools[0].GetName())
		assert.Len(t, tm.ListTools(), 1)

		// Execute the registered tool
		cmdTool := tm.ListTools()[0]
		assert.Equal(t, "echo", cmdTool.Tool().GetName())

		cmdTool.Tool().GetOutputSchema()
		outputSchema := cmdTool.Tool().GetOutputSchema()
		assert.NotNil(t, outputSchema)
		assert.Equal(t, "object", outputSchema.Fields["type"].GetStringValue())
		properties := outputSchema.Fields["properties"].GetStructValue().GetFields()
		assert.Contains(t, properties, "stdout")
		assert.Equal(t, "string", properties["stdout"].GetStructValue().GetFields()["type"].GetStringValue())

		inputData := map[string]interface{}{"args": []string{"hello from test"}}
		inputs, err := json.Marshal(inputData)
		require.NoError(t, err)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		result, err := cmdTool.Execute(context.Background(), req)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "hello from test\n", resultMap["stdout"])
		assert.Equal(t, "/bin/echo", resultMap["command"])
	})

	t.Run("successful prompt registration", func(t *testing.T) {
		tm := newMockToolManager()
		prm := prompt.NewManager()
		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test-prompt-service"),
			CommandLineService: configv1.CommandLineUpstreamService_builder{
				Prompts: []*configv1.PromptDefinition{
					configv1.PromptDefinition_builder{
						Name: proto.String("test-prompt"),
					}.Build(),
				},
			}.Build(),
		}.Build()

		serviceID, _, _, err := u.Register(
			context.Background(),
			serviceConfig,
			tm,
			prm,
			rm,
			false,
		)
		require.NoError(t, err)

		sanitizedName, _ := util.SanitizeServiceName("test-prompt-service")
		assert.Equal(t, sanitizedName, serviceID)

		p, ok := prm.GetPrompt(serviceID + ".test-prompt")
		require.True(t, ok)
		assert.Equal(t, serviceID+".test-prompt", p.Prompt().Name)
	})

	t.Run("successful dynamic resource registration", func(t *testing.T) {
		tm := newMockToolManager()
		rm := resource.NewManager()
		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test-dynamic-resource-service"),
			CommandLineService: configv1.CommandLineUpstreamService_builder{
				Calls: map[string]*configv1.CommandLineCallDefinition{
					"list-files-call": configv1.CommandLineCallDefinition_builder{
						Id: proto.String("list-files-call"),
					}.Build(),
				},
				Tools: []*configv1.ToolDefinition{
					configv1.ToolDefinition_builder{
						Name:   proto.String("list-files"),
						CallId: proto.String("list-files-call"),
					}.Build(),
				},
				Resources: []*configv1.ResourceDefinition{
					configv1.ResourceDefinition_builder{
						Name: proto.String("files"),
						Uri:  proto.String("test-dynamic-resource-service.files"),
						Dynamic: configv1.DynamicResource_builder{
							CommandLineCall: configv1.CommandLineCallDefinition_builder{
								Id: proto.String("list-files-call"),
							}.Build(),
						}.Build(),
					}.Build(),
				},
			}.Build(),
		}.Build()

		serviceID, _, _, err := u.Register(
			context.Background(),
			serviceConfig,
			tm,
			prm,
			rm,
			false,
		)
		require.NoError(t, err)
		assert.Len(t, rm.ListResources(), 1)
		dynResource, ok := rm.GetResource(serviceID + ".files")
		require.True(t, ok)
		assert.Equal(t, "files", dynResource.Resource().Name)
	})

	t.Run("missing call definition", func(t *testing.T) {
		tm := newMockToolManager()
		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test-missing-call-def"),
			CommandLineService: configv1.CommandLineUpstreamService_builder{
				Tools: []*configv1.ToolDefinition{
					configv1.ToolDefinition_builder{
						Name:   proto.String("echo"),
						CallId: proto.String("echo-call-missing"),
					}.Build(),
				},
			}.Build(),
		}.Build()

		_, _, _, err := u.Register(
			context.Background(),
			serviceConfig,
			tm,
			prm,
			rm,
			false,
		)
		require.NoError(t, err)
		assert.Len(t, tm.ListTools(), 0)
	})

	t.Run("nil command line service config", func(t *testing.T) {
		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test-nil-config"),
		}.Build()
		_, _, _, err := u.Register(context.Background(), serviceConfig, tm, prm, rm, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "command line service config is nil")
	})

	t.Run("invalid service name", func(t *testing.T) {
		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name: proto.String(""), // empty name
			CommandLineService: configv1.CommandLineUpstreamService_builder{}.Build(),
		}.Build()
		_, _, _, err := u.Register(context.Background(), serviceConfig, tm, prm, rm, false)
		require.Error(t, err)
	})

	t.Run("add tool error", func(t *testing.T) {
		tmWithError := newMockToolManager()
		tmWithError.addError = errors.New("failed to add tool")

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test-add-tool-error"),
			CommandLineService: configv1.CommandLineUpstreamService_builder{
				Calls: map[string]*configv1.CommandLineCallDefinition{
					"ls-call": configv1.CommandLineCallDefinition_builder{
						Id: proto.String("ls-call"),
					}.Build(),
				},
				Tools: []*configv1.ToolDefinition{
					configv1.ToolDefinition_builder{
						Name:   proto.String("ls"),
						CallId: proto.String("ls-call"),
					}.Build(),
				},
			}.Build(),
		}.Build()

		_, discoveredTools, _, err := u.Register(
			context.Background(),
			serviceConfig,
			tmWithError,
			prm,
			rm,
			false,
		)
		require.Error(t, err)
		assert.Nil(t, discoveredTools)
		assert.Empty(t, tmWithError.ListTools())
		assert.Contains(t, err.Error(), "failed to add tool")
	})
}

func TestStdioUpstream_Register_RequiredParams(t *testing.T) {
	tm := newMockToolManager()
	prm := prompt.NewManager()
	rm := resource.NewManager()
	u := NewUpstream()

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-required-params-service"),
		CommandLineService: configv1.CommandLineUpstreamService_builder{
			Command: proto.String("/bin/echo"),
			Calls: map[string]*configv1.CommandLineCallDefinition{
				"echo-call": configv1.CommandLineCallDefinition_builder{
					Id: proto.String("echo-call"),
					Parameters: []*configv1.CommandLineParameterMapping{
						configv1.CommandLineParameterMapping_builder{
							Schema: configv1.ParameterSchema_builder{
								Name:       proto.String("required_arg"),
								Type:       configv1.ParameterType_STRING.Enum(),
								IsRequired: proto.Bool(true),
							}.Build(),
						}.Build(),
					},
				}.Build(),
			},
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{
					Name:   proto.String("echo"),
					CallId: proto.String("echo-call"),
				}.Build(),
			},
		}.Build(),
	}.Build()

	_, _, _, err := u.Register(
		context.Background(),
		serviceConfig,
		tm,
		prm,
		rm,
		false,
	)
	require.NoError(t, err)

	tools := tm.ListTools()
	require.Len(t, tools, 1)

	tool := tools[0]
	inputSchema := tool.Tool().GetInputSchema()
	require.NotNil(t, inputSchema)

	// Check if 'required' field is present and contains 'required_arg'
	requiredVal, ok := inputSchema.Fields["required"]
	require.True(t, ok, "required field should be present")

	listVal := requiredVal.GetListValue()
	require.NotNil(t, listVal)

	paramName := "required_arg"
	found := false
	for _, v := range listVal.Values {
		if v.GetStringValue() == paramName {
			found = true
			break
		}
	}
	assert.True(t, found, "required_arg should be in the required list")
}

func TestStdioUpstream_Register_DynamicResourceErrors(t *testing.T) {
	u := NewUpstream()

	t.Run("dynamic resource tool not found", func(t *testing.T) {
		tm := newMockToolManager()
		tm.getError = true // Simulate tool not found
		rm := resource.NewManager()
		prm := prompt.NewManager()

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test-dynamic-resource-missing-tool"),
			CommandLineService: configv1.CommandLineUpstreamService_builder{
				Calls: map[string]*configv1.CommandLineCallDefinition{
					"list-files-call": configv1.CommandLineCallDefinition_builder{
						Id: proto.String("list-files-call"),
					}.Build(),
				},
				Tools: []*configv1.ToolDefinition{
					configv1.ToolDefinition_builder{
						Name:   proto.String("list-files"),
						CallId: proto.String("list-files-call"),
					}.Build(),
				},
				Resources: []*configv1.ResourceDefinition{
					configv1.ResourceDefinition_builder{
						Name: proto.String("files"),
						Dynamic: configv1.DynamicResource_builder{
							CommandLineCall: configv1.CommandLineCallDefinition_builder{
								Id: proto.String("list-files-call"),
							}.Build(),
						}.Build(),
					}.Build(),
				},
			}.Build(),
		}.Build()

		_, _, _, err := u.Register(
			context.Background(),
			serviceConfig,
			tm,
			prm,
			rm,
			false,
		)
		require.NoError(t, err)

		// Resource should NOT be added because tool was not found
		assert.Len(t, rm.ListResources(), 0)
	})

	t.Run("dynamic resource call id not found", func(t *testing.T) {
		tm := newMockToolManager()
		rm := resource.NewManager()
		prm := prompt.NewManager()

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test-dynamic-resource-missing-call"),
			CommandLineService: configv1.CommandLineUpstreamService_builder{
				Calls: make(map[string]*configv1.CommandLineCallDefinition),
				Resources: []*configv1.ResourceDefinition{
					configv1.ResourceDefinition_builder{
						Name: proto.String("files"),
						Dynamic: configv1.DynamicResource_builder{
							CommandLineCall: configv1.CommandLineCallDefinition_builder{
								Id: proto.String("list-files-call-missing"),
							}.Build(),
						}.Build(),
					}.Build(),
				},
			}.Build(),
		}.Build()

		_, _, _, err := u.Register(
			context.Background(),
			serviceConfig,
			tm,
			prm,
			rm,
			false,
		)
		require.NoError(t, err)

		// Resource should NOT be added because tool call was not found
		assert.Len(t, rm.ListResources(), 0)
	})

	t.Run("dynamic resource sanitization error", func(t *testing.T) {
		tm := newMockToolManager()
		rm := resource.NewManager()
		prm := prompt.NewManager()

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test-dynamic-resource-sanitization-error"),
			CommandLineService: configv1.CommandLineUpstreamService_builder{
				Calls: map[string]*configv1.CommandLineCallDefinition{
					"empty-name-call": configv1.CommandLineCallDefinition_builder{
						Id: proto.String("empty-name-call"),
					}.Build(),
				},
				Tools: []*configv1.ToolDefinition{
					configv1.ToolDefinition_builder{
						Name:    proto.String(""), // Empty name
						CallId:  proto.String("empty-name-call"),
						Disable: proto.Bool(true),
					}.Build(),
				},
				Resources: []*configv1.ResourceDefinition{
					configv1.ResourceDefinition_builder{
						Name: proto.String("files"),
						Dynamic: configv1.DynamicResource_builder{
							CommandLineCall: configv1.CommandLineCallDefinition_builder{
								Id: proto.String("empty-name-call"),
							}.Build(),
						}.Build(),
					}.Build(),
				},
			}.Build(),
		}.Build()

		_, _, _, err := u.Register(
			context.Background(),
			serviceConfig,
			tm,
			prm,
			rm,
			false,
		)
		require.NoError(t, err)

		// Resource should NOT be added
		assert.Len(t, rm.ListResources(), 0)
	})

	t.Run("dynamic resource call is nil", func(t *testing.T) {
		tm := newMockToolManager()
		rm := resource.NewManager()
		prm := prompt.NewManager()

		serviceConfig := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test-dynamic-resource-call-nil"),
			CommandLineService: configv1.CommandLineUpstreamService_builder{
				Resources: []*configv1.ResourceDefinition{
					configv1.ResourceDefinition_builder{
						Name: proto.String("files"),
						Dynamic: configv1.DynamicResource_builder{
							// Not setting CommandLineCall, so it is nil
						}.Build(),
					}.Build(),
				},
			}.Build(),
		}.Build()

		_, _, _, err := u.Register(
			context.Background(),
			serviceConfig,
			tm,
			prm,
			rm,
			false,
		)
		require.NoError(t, err)

		// Resource should NOT be added
		assert.Len(t, rm.ListResources(), 0)
	})
}

func TestStdioUpstream_Register_Coverage(t *testing.T) {
	tm := newMockToolManager()
	prm := prompt.NewManager()
	rm := resource.NewManager()
	u := NewUpstream()

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-coverage-service"),
		CommandLineService: configv1.CommandLineUpstreamService_builder{
			Local: proto.Bool(true), // Cover Local: true path
			Calls: map[string]*configv1.CommandLineCallDefinition{
				"disabled-call": configv1.CommandLineCallDefinition_builder{Id: proto.String("disabled-call")}.Build(),
				"enabled-call":  configv1.CommandLineCallDefinition_builder{Id: proto.String("enabled-call")}.Build(),
			},
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{
					Name:    proto.String("disabled-tool"),
					CallId:  proto.String("disabled-call"),
					Disable: proto.Bool(true),
				}.Build(),
				configv1.ToolDefinition_builder{
					Name:   proto.String("enabled-tool"),
					CallId: proto.String("enabled-call"),
				}.Build(),
			},
			Resources: []*configv1.ResourceDefinition{
				configv1.ResourceDefinition_builder{
					Name:    proto.String("disabled-resource"),
					Disable: proto.Bool(true),
					Static: configv1.StaticResource_builder{
						TextContent: proto.String("content"),
					}.Build(),
				}.Build(),
			},
			Prompts: []*configv1.PromptDefinition{
				configv1.PromptDefinition_builder{
					Name:    proto.String("disabled-prompt"),
					Disable: proto.Bool(true),
				}.Build(),
			},
		}.Build(),
	}.Build()

	_, discoveredTools, _, err := u.Register(
		context.Background(),
		serviceConfig,
		tm,
		prm,
		rm,
		false,
	)
	require.NoError(t, err)

	// Check results
	assert.Len(t, discoveredTools, 1)
	assert.Equal(t, "enabled-tool", discoveredTools[0].GetName())

	assert.Len(t, rm.ListResources(), 0) // Only disabled resource was added

	// Disabled prompt shouldn't be registered, but prompt manager mock or real one?
	// We use real prompt manager.
	// We can check if prompt exists.
	// ServiceID will be hashed name.
	sanitizedName, _ := util.SanitizeServiceName("test-coverage-service")
	_, ok := prm.GetPrompt(sanitizedName + ".disabled-prompt")
	assert.False(t, ok)
}

func TestStdioUpstream_Register_AnnotationsAndHints(t *testing.T) {
	tm := newMockToolManager()
	prm := prompt.NewManager()
	rm := resource.NewManager()
	u := NewUpstream()

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-annotations-service"),
		CommandLineService: configv1.CommandLineUpstreamService_builder{
			Command: proto.String("/bin/true"),
			Calls: map[string]*configv1.CommandLineCallDefinition{
				"true-call": configv1.CommandLineCallDefinition_builder{
					Id: proto.String("true-call"),
				}.Build(),
			},
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{
					Name:            proto.String("true-tool"),
					Title:           proto.String("True Tool"),
					Description:     proto.String("Returns true"),
					CallId:          proto.String("true-call"),
					ReadOnlyHint:    proto.Bool(true),
					DestructiveHint: proto.Bool(false),
					IdempotentHint:  proto.Bool(true),
					OpenWorldHint:   proto.Bool(false),
				}.Build(),
			},
		}.Build(),
	}.Build()

	_, _, _, err := u.Register(
		context.Background(),
		serviceConfig,
		tm,
		prm,
		rm,
		false,
	)
	require.NoError(t, err)

	// Verify the tool properties
	tools := tm.ListTools()
	require.Len(t, tools, 1)
	cmdTool := tools[0]
	toolProto := cmdTool.Tool()

	// Check Title in Annotations
	annotations := toolProto.GetAnnotations()
	require.NotNil(t, annotations)
	assert.Equal(t, "True Tool", annotations.GetTitle())

	// Check Hints in Annotations
	assert.True(t, annotations.GetReadOnlyHint())
	assert.False(t, annotations.GetDestructiveHint())
	assert.True(t, annotations.GetIdempotentHint())
	assert.False(t, annotations.GetOpenWorldHint())

	// Check DisplayName
	assert.Equal(t, "True Tool", toolProto.GetDisplayName())
}
