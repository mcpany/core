// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver_test

import (
	"context"
	"encoding/json"
	"testing"

	bus_pb "github.com/mcpany/core/proto/bus"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/mcpserver"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// --- Local Mocks for Hermetic Testing ---

type authTestPrompt struct {
	p       *mcp.Prompt
	service string
}

func (m *authTestPrompt) Prompt() *mcp.Prompt {
	return m.p
}

func (m *authTestPrompt) Service() string {
	return m.service
}

func (m *authTestPrompt) Get(ctx context.Context, args json.RawMessage) (*mcp.GetPromptResult, error) {
	return &mcp.GetPromptResult{
		Description: "Allowed Prompt Result",
	}, nil
}

type authTestResource struct {
	res     *mcp.Resource
	service string
}

func (m *authTestResource) Resource() *mcp.Resource {
	return m.res
}

func (m *authTestResource) Service() string {
	return m.service
}

func (m *authTestResource) Read(ctx context.Context) (*mcp.ReadResourceResult, error) {
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:  m.res.URI,
				Text: "Allowed Resource Content",
			},
		},
	}, nil
}

func (m *authTestResource) Subscribe(_ context.Context) error {
	return nil
}

// mockSession implements tool.Session for testing
type mockSession struct {
	createMessageFunc func(ctx context.Context, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error)
	listRootsFunc     func(ctx context.Context) (*mcp.ListRootsResult, error)
}

func (m *mockSession) CreateMessage(ctx context.Context, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error) {
	if m.createMessageFunc != nil {
		return m.createMessageFunc(ctx, params)
	}
	return nil, nil
}

func (m *mockSession) ListRoots(ctx context.Context) (*mcp.ListRootsResult, error) {
	if m.listRootsFunc != nil {
		return m.listRootsFunc(ctx)
	}
	return nil, nil
}

func TestServer_Auth_GetPrompt(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Mock Managers
	mockToolManager := tool.NewMockManagerInterface(ctrl)
	mockPromptManager := prompt.NewMockManagerInterface(ctrl)
	mockResourceManager := resource.NewMockManagerInterface(ctrl)
	authManager := auth.NewManager()

	// Expectations for NewServer initialization
	mockToolManager.EXPECT().SetMCPServer(gomock.Any())
	mockToolManager.EXPECT().AddTool(gomock.Any()).AnyTimes()
	mockPromptManager.EXPECT().SetMCPServer(gomock.Any())
	mockResourceManager.EXPECT().OnListChanged(gomock.Any())

	// Setup Server
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, _ := bus.NewProvider(messageBus)

	// We need a ServiceRegistry, but we can pass nil factory since we won't use it for this test
	serviceRegistry := serviceregistry.New(nil, mockToolManager, mockPromptManager, mockResourceManager, authManager)

	server, err := mcpserver.NewServer(
		context.Background(),
		mockToolManager,
		mockPromptManager,
		mockResourceManager,
		authManager,
		serviceRegistry,
		busProvider,
		false,
	)
	require.NoError(t, err)

	// Mock Prompt
	promptName := "test-prompt"
	serviceID := "test-service"
	testPrompt := &authTestPrompt{
		p:       &mcp.Prompt{Name: promptName},
		service: serviceID,
	}

	t.Run("Allowed", func(t *testing.T) {
		ctx := auth.ContextWithProfileID(context.Background(), "user-profile")

		mockPromptManager.EXPECT().GetPrompt(promptName).Return(testPrompt, true)
		mockToolManager.EXPECT().IsServiceAllowed(serviceID, "user-profile").Return(true)

		res, err := server.GetPrompt(ctx, &mcp.GetPromptRequest{
			Params: &mcp.GetPromptParams{
				Name: promptName,
			},
		})

		require.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "Allowed Prompt Result", res.Description)
	})

	t.Run("Denied", func(t *testing.T) {
		ctx := auth.ContextWithProfileID(context.Background(), "user-profile")

		mockPromptManager.EXPECT().GetPrompt(promptName).Return(testPrompt, true)
		mockToolManager.EXPECT().IsServiceAllowed(serviceID, "user-profile").Return(false)

		_, err := server.GetPrompt(ctx, &mcp.GetPromptRequest{
			Params: &mcp.GetPromptParams{
				Name: promptName,
			},
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "access denied")
	})

	t.Run("NoProfile_Allowed", func(t *testing.T) {
		ctx := context.Background() // No profile

		mockPromptManager.EXPECT().GetPrompt(promptName).Return(testPrompt, true)
		// IsServiceAllowed should NOT be called

		res, err := server.GetPrompt(ctx, &mcp.GetPromptRequest{
			Params: &mcp.GetPromptParams{
				Name: promptName,
			},
		})

		require.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "Allowed Prompt Result", res.Description)
	})
}

func TestServer_Auth_ReadResource(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockToolManager := tool.NewMockManagerInterface(ctrl)
	mockPromptManager := prompt.NewMockManagerInterface(ctrl)
	mockResourceManager := resource.NewMockManagerInterface(ctrl)
	authManager := auth.NewManager()

	// Expectations for NewServer initialization
	mockToolManager.EXPECT().SetMCPServer(gomock.Any())
	mockToolManager.EXPECT().AddTool(gomock.Any()).AnyTimes()
	mockPromptManager.EXPECT().SetMCPServer(gomock.Any())
	mockResourceManager.EXPECT().OnListChanged(gomock.Any())

	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, _ := bus.NewProvider(messageBus)
	serviceRegistry := serviceregistry.New(nil, mockToolManager, mockPromptManager, mockResourceManager, authManager)

	server, err := mcpserver.NewServer(
		context.Background(),
		mockToolManager,
		mockPromptManager,
		mockResourceManager,
		authManager,
		serviceRegistry,
		busProvider,
		false,
	)
	require.NoError(t, err)

	resourceURI := "test://resource"
	serviceID := "test-service"
	testResource := &authTestResource{
		res:     &mcp.Resource{URI: resourceURI},
		service: serviceID,
	}

	t.Run("Allowed", func(t *testing.T) {
		ctx := auth.ContextWithProfileID(context.Background(), "user-profile")

		mockResourceManager.EXPECT().GetResource(resourceURI).Return(testResource, true)
		mockToolManager.EXPECT().IsServiceAllowed(serviceID, "user-profile").Return(true)

		res, err := server.ReadResource(ctx, &mcp.ReadResourceRequest{
			Params: &mcp.ReadResourceParams{
				URI: resourceURI,
			},
		})

		require.NoError(t, err)
		assert.NotNil(t, res)
		require.Len(t, res.Contents, 1)
		assert.Equal(t, "Allowed Resource Content", res.Contents[0].Text)
	})

	t.Run("Denied", func(t *testing.T) {
		ctx := auth.ContextWithProfileID(context.Background(), "user-profile")

		mockResourceManager.EXPECT().GetResource(resourceURI).Return(testResource, true)
		mockToolManager.EXPECT().IsServiceAllowed(serviceID, "user-profile").Return(false)

		_, err := server.ReadResource(ctx, &mcp.ReadResourceRequest{
			Params: &mcp.ReadResourceParams{
				URI: resourceURI,
			},
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "access denied")
	})
}

func TestServer_CreateMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockToolManager := tool.NewMockManagerInterface(ctrl)
	// We need to set expectation for SetMCPServer because NewServer calls it
	mockToolManager.EXPECT().SetMCPServer(gomock.Any())
	// And AddTool for built-in tools (roots)
	mockToolManager.EXPECT().AddTool(gomock.Any()).AnyTimes()

	mockPromptManager := prompt.NewMockManagerInterface(ctrl)
	mockPromptManager.EXPECT().SetMCPServer(gomock.Any())

	mockResourceManager := resource.NewMockManagerInterface(ctrl)
	// NewServer sets OnListChanged
	mockResourceManager.EXPECT().OnListChanged(gomock.Any())

	authManager := auth.NewManager()
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, _ := bus.NewProvider(messageBus)
	serviceRegistry := serviceregistry.New(nil, mockToolManager, mockPromptManager, mockResourceManager, authManager)

	server, err := mcpserver.NewServer(
		context.Background(),
		mockToolManager,
		mockPromptManager,
		mockResourceManager,
		authManager,
		serviceRegistry,
		busProvider,
		false,
	)
	require.NoError(t, err)

	t.Run("NoSession", func(t *testing.T) {
		ctx := context.Background()
		_, err := server.CreateMessage(ctx, &mcp.CreateMessageParams{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no active session")
	})

	t.Run("WithSession", func(t *testing.T) {
		expectedResult := &mcp.CreateMessageResult{
			Role: mcp.Role("assistant"),
			Content: &mcp.TextContent{
				Text: "hello",
			},
		}

		session := &mockSession{
			createMessageFunc: func(ctx context.Context, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error) {
				return expectedResult, nil
			},
		}

		ctx := tool.NewContextWithSession(context.Background(), session)
		res, err := server.CreateMessage(ctx, &mcp.CreateMessageParams{})
		require.NoError(t, err)
		assert.Equal(t, expectedResult, res)
	})
}
