// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	"github.com/mcpany/core/server/tests/framework"
	"github.com/mcpany/core/server/tests/integration"
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

	// Try to locate binary in multiple common locations
	binPath := filepath.Join(root, "../build/test/bin/prompt-server")
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		// Try local build dir (if server/Makefile built relative to CWD)
		localBinPath := filepath.Join(root, "build/test/bin/prompt-server")
		if _, err := os.Stat(localBinPath); err == nil {
			binPath = localBinPath
		} else {
			t.Logf("Warning: prompt-server binary not found at %s or %s. Test may fail.", binPath, localBinPath)
		}
	}

	proc := integration.NewManagedProcess(t, "prompt_server", binPath, []string{"--port", fmt.Sprintf("%d", port)}, nil)
	proc.Port = port
	return proc
}

func RegisterPromptService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_prompt_server"
	integration.RegisterStreamableMCPService(t, registrationClient, serviceID, upstreamEndpoint, true, nil)
}

func InvokeAIWithPrompt(t *testing.T, mcpanyEndpoint string) {
	// Not using the gemini CLI for this test, as it's not working as expected.
	// Instead, we'll use the mcpserver directly.
	// The prompt server has a "hello" prompt that returns "Hello, world!".
	// We'll execute this prompt and check the output.
	// The prompt is registered with the service "e2e_prompt_server".
	// The tool is named "hello".

	// We need to create a client to connect to the mcpserver.
	// We can use the mcp.NewClient function to create a new client.
	// We also need to create a transport to connect to the server.
	// We can use the mcp.NewStreamableHTTPTransport function to create a new transport.

	// Give the server a moment to start up.
	time.Sleep(5 * time.Second)

	// Create a new MCP client.
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)

	// Create a new streamable HTTP transport.
	transport := &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}

	// Connect to the server.
	session, err := client.Connect(context.Background(), transport, nil)
	require.NoError(t, err, "failed to connect to mcp server")
	defer func() { _ = session.Close() }()

	// Get the "hello" prompt.
	result, err := session.GetPrompt(context.Background(), &mcp.GetPromptParams{
		Name: "hello",
	})
	require.NoError(t, err, "failed to get prompt")

	// Check the output.
	require.Len(t, result.Messages, 1)
	assert.Equal(t, mcp.Role("user"), result.Messages[0].Role)
	assert.Equal(t, "Hello, world!", result.Messages[0].Content.(*mcp.TextContent).Text)
}
