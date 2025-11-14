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
	"path/filepath"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/config"
	"github.com/mcpany/core/pkg/mcpserver"
	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/serviceregistry"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/upstream/factory"
	bus_pb "github.com/mcpany/core/proto/bus"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromptConfigEndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a temporary config file
	configContent := `
upstream_services: {
	name: "service-with-prompt"
	http_service: {
		address: "http://api.example.com/v1"
	}
	prompts: {
		name: "test-prompt-from-config"
		title: "Test Prompt From Config"
		description: "This is a test prompt from a config file."
		arguments: {
			name: "arg1"
			description: "Argument 1"
			required: true
		}
		messages: {
			role: "user"
			text_content: {
				text: "Hello, {{arg1}}"
			}
		}
	}
}
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_config.textproto")
	err := os.WriteFile(filePath, []byte(configContent), 0o644)
	require.NoError(t, err, "Failed to write temp config file")

	// Load the config
	fs := afero.NewOsFs()
	fileStore := config.NewFileStore(fs, []string{filePath})
	cfg, err := config.LoadServices(fileStore, "server")
	require.NoError(t, err)

	// Server setup
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewBusProvider(messageBus)
	require.NoError(t, err)
	poolManager := pool.NewManager()
	toolManager := tool.NewToolManager(busProvider)
	promptManager := prompt.NewPromptManager()
	resourceManager := resource.NewResourceManager()
	authManager := auth.NewAuthManager()
	serviceRegistry := serviceregistry.New(
		factory.NewUpstreamServiceFactory(poolManager),
		toolManager,
		promptManager,
		resourceManager,
		authManager,
	)

	// Register services from config
	for _, serviceConfig := range cfg.GetUpstreamServices() {
		_, _, _, err := serviceRegistry.RegisterService(ctx, serviceConfig)
		require.NoError(t, err)
	}

	mcpServer, err := mcpserver.NewServer(
		ctx,
		toolManager,
		promptManager,
		resourceManager,
		authManager,
		serviceRegistry,
		busProvider,
	)
	require.NoError(t, err)

	// Client setup
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	// Connect server and client
	serverSession, err := mcpServer.Server().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer serverSession.Close()
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	// Test prompts/list
	listResult, err := clientSession.ListPrompts(ctx, &mcp.ListPromptsParams{})
	require.NoError(t, err)
	require.Len(t, listResult.Prompts, 1)
	assert.Equal(t, "test-prompt-from-config", listResult.Prompts[0].Name)
	assert.Equal(t, "Test Prompt From Config", listResult.Prompts[0].Title)

	// Test prompts/get
	getResult, err := clientSession.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "test-prompt-from-config",
		Arguments: map[string]string{
			"arg1": "world",
		},
	})
	require.NoError(t, err)
	require.Len(t, getResult.Messages, 1)
	textContent, ok := getResult.Messages[0].Content.(*mcp.TextContent)
	require.True(t, ok)
	assert.Equal(t, "Hello, world", textContent.Text)
}
