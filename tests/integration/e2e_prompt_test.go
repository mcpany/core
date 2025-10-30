package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	apiv1 "github.com/mcpxy/core/proto/api/v1"
	"github.com/mcpxy/core/tests/framework"
	"github.com/mcpxy/core/tests/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2EPrompt(t *testing.T) {
	t.Skip("Skipping TestE2EPrompt for now.")
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

func InvokeAIWithPrompt(t *testing.T, mcpxyEndpoint string) {
	gemini := framework.NewGeminiCLI(t)
	gemini.Install()

	// Configure the MCP server with the Gemini CLI
	serverName := "mcpxy_e2e_prompt_test"
	gemini.AddMCP(serverName, mcpxyEndpoint)
	defer gemini.RemoveMCP(serverName)

	// Run a prompt and validate the output
	prompt := fmt.Sprintf("@%s hello", serverName)
	output, err := gemini.Run(os.Getenv("GEMINI_API_KEY"), prompt)
	require.NoError(t, err, "gemini-cli failed to run")

	assert.Contains(t, output, "Hello, world!", "The output should contain 'Hello, world!'")
}
