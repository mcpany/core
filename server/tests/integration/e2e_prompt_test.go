// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration_test

import (
	"context"
	"path/filepath"
	"regexp"
	"strconv"
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
		WaitForUpstreamPort: WaitForPromptServerPort,
	})
}

func BuildPromptServer(t *testing.T) *integration.ManagedProcess {
	// Use port 0 for dynamic allocation
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)
	proc := integration.NewManagedProcess(t, "prompt_server", filepath.Join(root, "../build/test/bin/prompt-server"), []string{"--port", "0"}, nil)
	proc.Port = 0
	return proc
}

func WaitForPromptServerPort(t *testing.T, proc *integration.ManagedProcess) int {
	t.Helper()
	var port int
	timeout := 10 * time.Second
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// Regex for "server listening at 127.0.0.1:54321"
	re := regexp.MustCompile(`server listening at .*?:(\d+)`)

	for time.Now().Before(deadline) {
		out := proc.StderrString() // prompt-server uses log which goes to stderr
		matches := re.FindStringSubmatch(out)
		if len(matches) >= 2 {
			if p, err := strconv.Atoi(matches[1]); err == nil {
				port = p
				t.Logf("WaitForPromptServerPort: Found port %d", port)
				return port
			}
		}
		select {
		case <-ticker.C:
			continue
		default:
		}
	}
	t.Fatalf("Prompt server did not output a port within %v. Stderr: %q", timeout, proc.StderrString())
	return 0
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
