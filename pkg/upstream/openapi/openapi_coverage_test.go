/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package openapi

import (
	"context"
	"testing"

	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type MockPromptManager struct {
	mock.Mock
}

func (m *MockPromptManager) AddPrompt(p prompt.Prompt) {
	m.Called(p)
}

func (m *MockPromptManager) GetPrompt(name string) (prompt.Prompt, bool) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(prompt.Prompt), args.Bool(1)
}

func (m *MockPromptManager) ClearPromptsForService(serviceID string) {
	m.Called(serviceID)
}

func (m *MockPromptManager) ListPrompts() []prompt.Prompt {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]prompt.Prompt)
}

func (m *MockPromptManager) SetMCPServer(provider prompt.MCPServerProvider) {
	m.Called(provider)
}

func TestOpenAPIUpstream_Coverage(t *testing.T) {
	ctx := context.Background()
	mockToolManager := new(MockToolManager)
	mockPromptManager := new(MockPromptManager)
	u := NewOpenAPIUpstream()

	spec := `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      operationId: getTest
      responses:
        '200':
          description: OK
`

	t.Run("Disabled Tool and Prompt", func(t *testing.T) {
		config := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test-service-disabled"),
			OpenapiService: configv1.OpenapiUpstreamService_builder{
				OpenapiSpec: proto.String(spec),
				Tools: []*configv1.ToolDefinition{
					configv1.ToolDefinition_builder{
						Name:    proto.String("getTest"),
						CallId:  proto.String("getTest-call"),
						Disable: proto.Bool(true),
					}.Build(),
				},
				Calls: map[string]*configv1.OpenAPICallDefinition{
					"getTest-call": configv1.OpenAPICallDefinition_builder{
						Id: proto.String("getTest-call"),
					}.Build(),
				},
				Prompts: []*configv1.PromptDefinition{
					configv1.PromptDefinition_builder{
						Name:    proto.String("test-prompt"),
						Disable: proto.Bool(true),
					}.Build(),
				},
			}.Build(),
		}.Build()

		expectedKey, _ := util.SanitizeServiceName("test-service-disabled")
		mockToolManager.On("AddServiceInfo", expectedKey, mock.Anything).Return().Once()
		mockToolManager.On("GetTool", mock.Anything).Return(nil, false)
		// AddTool should NOT be called for disabled tool
		// AddPrompt should NOT be called for disabled prompt

		_, discoveredTools, _, err := u.Register(ctx, config, mockToolManager, mockPromptManager, nil, false)
		require.NoError(t, err)
		assert.Len(t, discoveredTools, 0)

		mockToolManager.AssertExpectations(t)
		mockPromptManager.AssertNotCalled(t, "AddPrompt", mock.Anything)
	})

	t.Run("Enabled Prompt", func(t *testing.T) {
		config := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test-service-prompt"),
			OpenapiService: configv1.OpenapiUpstreamService_builder{
				OpenapiSpec: proto.String(spec),
				Prompts: []*configv1.PromptDefinition{
					configv1.PromptDefinition_builder{
						Name: proto.String("test-prompt-enabled"),
						Messages: []*configv1.PromptMessage{
							configv1.PromptMessage_builder{
								Role: configv1.PromptMessage_USER.Enum(),
								Text: configv1.TextContent_builder{Text: proto.String("hi")}.Build(),
							}.Build(),
						},
					}.Build(),
				},
			}.Build(),
		}.Build()

		expectedKey, _ := util.SanitizeServiceName("test-service-prompt")
		mockToolManager.On("AddServiceInfo", expectedKey, mock.Anything).Return().Once()
		// We expect AddTool might be called for auto-discovered tool "getTest" if not disabled
		mockToolManager.On("AddTool", mock.Anything).Return(nil)

		mockPromptManager.On("AddPrompt", mock.Anything).Return()

		_, _, _, err := u.Register(ctx, config, mockToolManager, mockPromptManager, nil, false)
		require.NoError(t, err)

		mockPromptManager.AssertCalled(t, "AddPrompt", mock.Anything)
	})
}
