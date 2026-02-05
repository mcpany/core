// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/mcpserver"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	bus_pb "github.com/mcpany/core/proto/bus"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestResourceListNotification(t *testing.T) {
	poolManager := pool.NewManager()
	factory := factory.NewUpstreamServiceFactory(poolManager, nil)
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewProvider(messageBus)
	require.NoError(t, err)
	toolManager := tool.NewManager(busProvider)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	authManager := auth.NewManager()
	serviceRegistry := serviceregistry.New(factory, toolManager, promptManager, resourceManager, authManager)
	ctx := context.Background()

	server, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider, false)
	require.NoError(t, err)

	// Create client
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	// Capture notifications using middleware
	var notificationReceived bool
	var mu sync.Mutex
	notificationChan := make(chan struct{}, 1)

	client.AddReceivingMiddleware(func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			if method == "notifications/resources/list_changed" {
				mu.Lock()
				// Only signal if we were waiting for one (channel open)
				select {
				case <-notificationChan:
					// Already closed/signaled
				default:
					if !notificationReceived {
						notificationReceived = true
						close(notificationChan)
					}
				}
				mu.Unlock()
			}
			return next(ctx, method, req)
		}
	})

	// Connect server and client
	serverSession, err := server.Server().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	// Trigger multiple resource list changes to ensure:
	// 1. Notifications are fired each time.
	// 2. The workaround handles duplicate additions gracefully (idempotency).

	// First update
	resourceManager.AddResource(&testResource{
		URIValue: "test-resource-1",
	})

	// Wait for first notification
	select {
	case <-notificationChan:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for first resource list changed notification")
	}

	// Reset notification flag
	mu.Lock()
	notificationReceived = false
	notificationChan = make(chan struct{}, 1)
	mu.Unlock()

	// Second update - this triggers the workaround again with the same dummy resource
	resourceManager.AddResource(&testResource{
		URIValue: "test-resource-2",
	})

	// Wait for second notification
	select {
	case <-notificationChan:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for second resource list changed notification")
	}
}
