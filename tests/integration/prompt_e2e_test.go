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
	"fmt"
	"path/filepath"
	"testing"
	"time"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	"github.com/mcpany/core/tests/framework"
	"github.com/mcpany/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2EPrompt(t *testing.T) {
	framework.RunE2ETest(t, &framework.E2ETestCase{
		Name:                "prompt",
		UpstreamServiceType: "http",
		BuildUpstream:       BuildPromptServer,
		RegisterUpstream:    RegisterPromptService,
		InvokeAIClient:      InvokeAIWithPrompt,
		RegistrationMethods: []framework.RegistrationMethod{framework.GRPCRegistration},
	})
}

func BuildPromptServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)
	proc := integration.NewManagedProcess(t, "prompt_server", filepath.Join(root, "build/test/bin/prompt-server"), []string{"--port", fmt.Sprintf("%d", port)}, nil)
	proc.Port = port
	return proc
}

func RegisterPromptService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_prompt_server"
	integration.RegisterStreamableMCPService(t, registrationClient, serviceID, upstreamEndpoint, true, nil)
}

func InvokeAIWithPrompt(t *testing.T, mcpanyEndpoint string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Client setup
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	transport := &mcp.StreamableClientTransport{
		Endpoint: mcpanyEndpoint,
	}

	// Connect server and client
	clientSession, err := client.Connect(ctx, transport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	// Test prompts/list
	listResult, err := clientSession.ListPrompts(ctx, &mcp.ListPromptsParams{})
	require.NoError(t, err)
	require.Len(t, listResult.Prompts, 1)
	assert.Equal(t, "hello", listResult.Prompts[0].Name)
	assert.Equal(t, "", listResult.Prompts[0].Title)

	// Test prompts/get
	getResult, err := clientSession.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "hello",
		Arguments: map[string]string{
			"name": "World",
		},
	})
	require.NoError(t, err)
	require.Len(t, getResult.Messages, 1)
	textContent, ok := getResult.Messages[0].Content.(*mcp.TextContent)
	require.True(t, ok)
	assert.Equal(t, "Hello, world!", textContent.Text)
}
