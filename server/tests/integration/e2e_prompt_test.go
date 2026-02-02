// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration_test

import (
	"os"
	"os/exec"
	"context"
	"fmt"
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
	binaryPath := filepath.Join(root, "../build/test/bin/prompt-server")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Logf("Binary not found at %s, attempting to build...", binaryPath)
		// Assuming the source is at cmd/mocks/prompt-server
		sourcePath := filepath.Join(root, "tests/integration/cmd/mocks/prompt-server/main.go")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		// Ensure output directory exists
		err := os.MkdirAll(filepath.Dir(binaryPath), 0755)
		require.NoError(t, err, "Failed to create bin directory")
		cmd := exec.CommandContext(ctx, "go", "build", "-o", binaryPath, sourcePath) //nolint:gosec
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = root // Run from module root
		err = cmd.Run()
		require.NoError(t, err, "Failed to build prompt-server")
	}
	proc := integration.NewManagedProcess(t, "prompt_server", binaryPath, []string{"--port", fmt.Sprintf("%d", port)}, nil)
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
