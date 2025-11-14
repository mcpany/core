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
	"testing"
	"time"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/mcpserver"
	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/serviceregistry"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream/factory"
	bus_pb "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/require"
)

func NewTestServer(t *testing.T) *mcpserver.Server {
	t.Helper()

	ctx := context.Background()
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

	server, err := mcpserver.NewServer(
		ctx,
		toolManager,
		promptManager,
		resourceManager,
		authManager,
		serviceRegistry,
		busProvider,
	)
	require.NoError(t, err)

	return server
}

func TestPromptFieldsIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Server setup
	server := NewTestServer(t)

	// Create a test service config with a tool and a resource that have prompts
	promptDef := &configv1.PromptDefinition{}
	promptDef.SetName("test-prompt")

	toolDef := &configv1.ToolDefinition{}
	toolDef.SetName("test-tool")
	toolDef.SetPrompts([]*configv1.PromptDefinition{promptDef})

	httpService := &configv1.HttpUpstreamService{}
	httpService.SetAddress("http://localhost:8080")
	httpService.SetTools([]*configv1.ToolDefinition{toolDef})

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-service")
	serviceConfig.SetHttpService(httpService)

	_, _, _, err := server.ServiceRegistry().RegisterService(ctx, serviceConfig)
	require.NoError(t, err)
}
