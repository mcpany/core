// Copyright 2025 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package integration_test

import (
	"context"
	"testing"

	"github.com/mcpany/core/tests/framework"
	"github.com/mcpany/core/tests/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

func BuildNoOpUpstream(t *testing.T) *integration.ManagedProcess {
	return integration.NewManagedProcess(t, "noop_upstream", "/bin/true", []string{}, nil)
}

func NoOpInvokeAIClient(t *testing.T, mcpanyEndpoint string) {
	// Do nothing, verification is done in the registration step.
}

func TestE2EPromptField(t *testing.T) {
	framework.RunE2ETest(t, &framework.E2ETestCase{
		Name:                "prompt_field",
		UpstreamServiceType: "grpc",
		BuildUpstream:       BuildNoOpUpstream,
		RegisterUpstream:    RegisterAndVerifyServiceWithPrompts,
		InvokeAIClient:      NoOpInvokeAIClient,
		RegistrationMethods: []framework.RegistrationMethod{framework.GRPCRegistration},
	})
}

func RegisterAndVerifyServiceWithPrompts(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceName = "service_with_prompts"
	const callID = "test_call"

	// 1. Define all the config pieces
	toolPrompt := &configv1.PromptDefinition{}
	toolPrompt.SetId("prompt1")
	toolPrompt.SetTemplate("This is a test prompt.")

	tool := &configv1.ToolDefinition{}
	tool.SetName("tool_with_prompt")
	tool.SetPrompts([]*configv1.PromptDefinition{toolPrompt})
	tool.SetCallId(callID)

	resourcePrompt := &configv1.PromptDefinition{}
	resourcePrompt.SetId("prompt2")
	resourcePrompt.SetTemplate("This is another test prompt.")

	resource := &configv1.ResourceDefinition{}
	resource.SetName("resource_with_prompt")
	resource.SetPrompts([]*configv1.PromptDefinition{resourcePrompt})

	httpCall := &configv1.HttpCallDefinition{}
	httpCall.SetId(callID)
	httpCall.SetEndpointPath("/dummy")
	httpCall.SetMethod(configv1.HttpCallDefinition_HTTP_METHOD_GET)

	httpService := &configv1.HttpUpstreamService{}
	httpService.SetAddress("http://localhost:8080")
	httpService.SetTools([]*configv1.ToolDefinition{tool})
	httpService.SetResources([]*configv1.ResourceDefinition{resource})
	httpService.SetCalls(map[string]*configv1.HttpCallDefinition{callID: httpCall})

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName(serviceName)
	serviceConfig.SetHttpService(httpService)

	req := &apiv1.RegisterServiceRequest{}
	req.SetConfig(serviceConfig)

	// 2. Register the service
	_, err := registrationClient.RegisterService(context.Background(), req)
	require.NoError(t, err)

	// 3. Verify the service was registered correctly
	resp, err := registrationClient.ListServices(context.Background(), &apiv1.ListServicesRequest{})
	require.NoError(t, err)

	var registeredService *configv1.UpstreamServiceConfig
	for _, s := range resp.GetServices() {
		if s.GetName() == serviceName {
			registeredService = s
			break
		}
	}
	require.NotNil(t, registeredService, "Failed to find registered service '%s'", serviceName)

	registeredHttpService := registeredService.GetHttpService()
	require.NotNil(t, registeredHttpService)

	// Verify tool prompts
	require.Len(t, registeredHttpService.GetTools(), 1)
	registeredTool := registeredHttpService.GetTools()[0]
	assert.Equal(t, "tool_with_prompt", registeredTool.GetName())
	require.Len(t, registeredTool.GetPrompts(), 1)
	assert.Equal(t, "prompt1", registeredTool.GetPrompts()[0].GetId())
	assert.Equal(t, "This is a test prompt.", registeredTool.GetPrompts()[0].GetTemplate())

	// Verify resource prompts
	require.Len(t, registeredHttpService.GetResources(), 1)
	registeredResource := registeredHttpService.GetResources()[0]
	assert.Equal(t, "resource_with_prompt", registeredResource.GetName())
	require.Len(t, registeredResource.GetPrompts(), 1)
	assert.Equal(t, "prompt2", registeredResource.GetPrompts()[0].GetId())
	assert.Equal(t, "This is another test prompt.", registeredResource.GetPrompts()[0].GetTemplate())
}
