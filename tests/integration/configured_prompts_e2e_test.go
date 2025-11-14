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

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfiguredPromptsEndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a temp config file with a prompt
	promptName := "test_configured_prompt"
	promptTitle := "Test Configured Prompt"
	promptDescription := "A prompt for testing configured prompts."
	promptArgName := "user_input"
	promptMessageText := "Hello, {{.user_input}}!"
	serviceName := "test-service-with-prompt"
	address := "http://localhost:12345"
	argDesc := "The user's input."

	configContent := `
upstream_services:
  - name: ` + serviceName + `
    http_service:
      address: ` + address + `
      prompts:
        - name: ` + promptName + `
          title: ` + promptTitle + `
          description: ` + promptDescription + `
          arguments:
            - name: ` + promptArgName + `
              description: ` + argDesc + `
              required: true
          messages:
            - role: USER
              text:
                text: "` + promptMessageText + `"
`
	tmpFile, err := os.CreateTemp(t.TempDir(), "mcpany-config-*.yaml")
	require.NoError(t, err)
	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)
	configFile := tmpFile.Name()

	// Start the MCPANY server with the config file
	serverInfo := StartMCPANYServer(t, "configured-prompts", "--config-path", configFile)
	defer serverInfo.CleanupFunc()

	// Client setup
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	httpTransport := &mcp.StreamableClientTransport{Endpoint: serverInfo.HTTPEndpoint}
	clientSession, err := client.Connect(ctx, httpTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	// Test prompts/list
	listResult, err := clientSession.ListPrompts(ctx, &mcp.ListPromptsParams{})
	require.NoError(t, err)
	require.Len(t, listResult.Prompts, 1)
	assert.Equal(t, promptName, listResult.Prompts[0].Name)
	assert.Equal(t, promptTitle, listResult.Prompts[0].Title)

	// Test prompts/get
	getResult, err := clientSession.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: promptName,
		Arguments: map[string]string{
			promptArgName: "world",
		},
	})
	require.NoError(t, err)
	require.Len(t, getResult.Messages, 1)
	textContent, ok := getResult.Messages[0].Content.(*mcp.TextContent)
	require.True(t, ok)
	assert.Equal(t, "Hello, world!", textContent.Text)
}
