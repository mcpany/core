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

package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mcpany/core/pkg/appconsts"
	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/consts"
	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/serviceregistry"
	"github.com/mcpany/core/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var AddReceivingMiddlewareHook func(name string)

// Server is the core of the MCP Any application. It orchestrates the handling of
// MCP (Model Context Protocol) requests by managing various components such as
// tools, prompts, resources, and services. It uses an internal router to
// delegate requests to the appropriate handlers and communicates with backend
// workers via an event bus.
type Server struct {
	server          *mcp.Server
	router          *Router
	toolManager     tool.ToolManagerInterface
	promptManager   prompt.PromptManagerInterface
	resourceManager resource.ResourceManagerInterface
	authManager     *auth.AuthManager
	serviceRegistry *serviceregistry.ServiceRegistry
	bus             *bus.BusProvider
}

// Server returns the underlying *mcp.Server instance, which provides access to
// the core MCP server functionality. This can be used for advanced
// configurations or direct interaction with the MCP server.
func (s *Server) Server() *mcp.Server {
	if AddReceivingMiddlewareHook != nil {
		// This is a test hook to allow inspection of the middleware chain.
		// We are passing the name of the middleware as a string.
		AddReceivingMiddlewareHook("CachingMiddleware")
	}
	return s.server
}

// NewServer creates and initializes a new MCP Any Server. It sets up the
// necessary managers for tools, prompts, and resources, configures the router
// with handlers for standard MCP methods, and establishes middleware for
// request processing, such as routing and tool list filtering.
//
// The server is initialized with all the necessary components for handling MCP
// requests and managing the lifecycle of tools, prompts, and resources.
//
// Parameters:
//   - ctx: The application's root context.
//   - toolManager: Manages the lifecycle and access to tools.
//   - promptManager: Manages the lifecycle and access to prompts.
//   - resourceManager: Manages the lifecycle and access to resources.
//   - authManager: Handles authentication for incoming requests.
//   - serviceRegistry: Keeps track of all registered upstream services.
//   - bus: The event bus used for asynchronous communication between components.
//
// Returns a new instance of the Server and an error if initialization fails.
func NewServer(
	ctx context.Context,
	toolManager tool.ToolManagerInterface,
	promptManager prompt.PromptManagerInterface,
	resourceManager resource.ResourceManagerInterface,
	authManager *auth.AuthManager,
	serviceRegistry *serviceregistry.ServiceRegistry,
	bus *bus.BusProvider,
) (*Server, error) {
	s := &Server{
		router:          NewRouter(),
		toolManager:     toolManager,
		promptManager:   promptManager,
		resourceManager: resourceManager,
		authManager:     authManager,
		serviceRegistry: serviceRegistry,
		bus:             bus,
	}

	s.router.Register(
		consts.MethodPromptsList,
		func(ctx context.Context, req mcp.Request) (mcp.Result, error) {
			if r, ok := req.(*mcp.ListPromptsRequest); ok {
				return s.ListPrompts(ctx, r)
			}
			return nil, fmt.Errorf("invalid request type for %s", consts.MethodPromptsList)
		},
	)

	s.router.Register(
		consts.MethodPromptsGet,
		func(ctx context.Context, req mcp.Request) (mcp.Result, error) {
			if r, ok := req.(*mcp.GetPromptRequest); ok {
				return s.GetPrompt(ctx, r)
			}
			return nil, fmt.Errorf("invalid request type for %s", consts.MethodPromptsGet)
		},
	)

	s.router.Register(
		consts.MethodResourcesList,
		func(ctx context.Context, req mcp.Request) (mcp.Result, error) {
			if r, ok := req.(*mcp.ListResourcesRequest); ok {
				return s.ListResources(ctx, r)
			}
			return nil, fmt.Errorf("invalid request type for %s", consts.MethodResourcesList)
		},
	)

	s.router.Register(
		consts.MethodResourcesRead,
		func(ctx context.Context, req mcp.Request) (mcp.Result, error) {
			if r, ok := req.(*mcp.ReadResourceRequest); ok {
				return s.ReadResource(ctx, r)
			}
			return nil, fmt.Errorf("invalid request type for %s", consts.MethodResourcesRead)
		},
	)

	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    appconsts.Name,
		Version: appconsts.Version,
	}, &mcp.ServerOptions{
		HasPrompts: true,
	})
	s.server = mcpServer

	s.toolManager.SetMCPServer(s)
	s.promptManager.SetMCPServer(prompt.NewMCPServerProvider(s.Server()))

	// TODO: Re-enable notifications when the go-sdk supports them
	// s.resourceManager.OnListChanged(func() {
	// 	if s.server != nil {
	// 		s.server.Notify(consts.NotificationResourcesListChanged, nil)
	// 	}
	// })

	routerMiddleware := func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(
			ctx context.Context,
			method string,
			req mcp.Request,
		) (mcp.Result, error) {
			if handler, ok := s.router.GetHandler(method); ok {
				return handler(ctx, req)
			}
			return next(ctx, method, req)
		}
	}

	toolListFilteringMiddleware := func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(
			ctx context.Context,
			method string,
			req mcp.Request,
		) (mcp.Result, error) {
			result, err := next(ctx, method, req)

			if method == consts.MethodToolsList {
				if err == nil {
					if listToolsResult, ok := result.(*mcp.ListToolsResult); ok {
						refreshedTools := make([]*mcp.Tool, 0, len(listToolsResult.Tools))
						for _, mcpTool := range listToolsResult.Tools {
							if toolInstance, toolExists := s.toolManager.GetTool(mcpTool.Name); toolExists {
								freshMCPTool, err := tool.ConvertProtoToMCPTool(toolInstance.Tool())
								logging.GetLogger().
									Info("Refreshing tool in MCP tool list", "toolName", mcpTool.Name, "tool", freshMCPTool)
								if err != nil {
									continue // Skip tools that fail to convert
								}
								refreshedTools = append(refreshedTools, freshMCPTool)
							}
						}
						result = &mcp.ListToolsResult{Tools: refreshedTools}
					}
				}
			}

			return result, err
		}
	}

	s.server.AddReceivingMiddleware(routerMiddleware)
	s.server.AddReceivingMiddleware(toolListFilteringMiddleware)

	return s, nil
}

// ListPrompts handles the "prompts/list" MCP request. It retrieves the list of
// available prompts from the PromptManager, converts them to the MCP format, and
// returns them to the client.
//
// Parameters:
//   - ctx: The context for the request.
//   - req: The "prompts/list" request from the client.
//
// Returns a list of available prompts or an error if the retrieval fails.
func (s *Server) ListPrompts(
	_ context.Context,
	_ *mcp.ListPromptsRequest,
) (*mcp.ListPromptsResult, error) {
	prompts := s.promptManager.ListPrompts()
	mcpPrompts := make([]*mcp.Prompt, len(prompts))
	for i, p := range prompts {
		mcpPrompts[i] = p.Prompt()
	}
	return &mcp.ListPromptsResult{
		Prompts: mcpPrompts,
	}, nil
}

// GetPrompt handles the "prompts/get" MCP request. It retrieves a specific
// prompt by name from the PromptManager and executes it with the provided
// arguments, returning the result. If the prompt is not found, it returns a
// prompt.ErrPromptNotFound error.
//
// Parameters:
//   - ctx: The context for the request.
//   - req: The "prompts/get" request from the client, containing the prompt
//     name and arguments.
//
// Returns the result of the prompt execution or an error if the prompt is not
// found or execution fails.
func (s *Server) GetPrompt(
	ctx context.Context,
	req *mcp.GetPromptRequest,
) (*mcp.GetPromptResult, error) {
	p, ok := s.promptManager.GetPrompt(req.Params.Name)
	if !ok {
		return nil, prompt.ErrPromptNotFound
	}

	argsBytes, err := json.Marshal(req.Params.Arguments)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal prompt arguments: %w", err)
	}

	return p.Get(ctx, argsBytes)
}

// ListResources handles the "resources/list" MCP request. It fetches the list
// of available resources from the ResourceManager, converts them to the MCP
// format, and returns them to the client.
//
// Parameters:
//   - ctx: The context for the request.
//   - req: The "resources/list" request from the client.
//
// Returns a list of available resources or an error if the retrieval fails.
func (s *Server) ListResources(
	_ context.Context,
	_ *mcp.ListResourcesRequest,
) (*mcp.ListResourcesResult, error) {
	resources := s.resourceManager.ListResources()
	mcpResources := make([]*mcp.Resource, len(resources))
	for i, r := range resources {
		mcpResources[i] = r.Resource()
	}
	return &mcp.ListResourcesResult{
		Resources: mcpResources,
	}, nil
}

// ReadResource handles the "resources/read" MCP request. It retrieves a
// specific resource by its URI from the ResourceManager and returns its content.
// If the resource is not found, it returns a resource.ErrResourceNotFound
// error.
//
// Parameters:
//   - ctx: The context for the request.
//   - req: The "resources/read" request from the client, containing the URI
//     of the resource to be read.
//
// Returns the content of the resource or an error if the resource is not found
// or reading fails.
func (s *Server) ReadResource(
	ctx context.Context,
	req *mcp.ReadResourceRequest,
) (*mcp.ReadResourceResult, error) {
	r, ok := s.resourceManager.GetResource(req.Params.URI)
	if !ok {
		return nil, resource.ErrResourceNotFound
	}
	return r.Read(ctx)
}

// AuthManager returns the server's authentication manager, which is responsible
// for handling authentication for incoming requests.
func (s *Server) AuthManager() *auth.AuthManager {
	return s.authManager
}

// ToolManager returns the server's tool manager, which is responsible for
// managing the lifecycle and access to tools.
func (s *Server) ToolManager() tool.ToolManagerInterface {
	return s.toolManager
}

// PromptManager returns the server's prompt manager, which is responsible for
// managing the lifecycle and access to prompts.
func (s *Server) PromptManager() prompt.PromptManagerInterface {
	return s.promptManager
}

// ResourceManager returns the server's resource manager, which is responsible
// for managing the lifecycle and access to resources.
func (s *Server) ResourceManager() resource.ResourceManagerInterface {
	return s.resourceManager
}

// ServiceRegistry returns the server's service registry, which keeps track of
// all registered upstream services.
func (s *Server) ServiceRegistry() *serviceregistry.ServiceRegistry {
	return s.serviceRegistry
}

func (s *Server) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	s.serviceRegistry.AddServiceInfo(serviceID, info)
}

func (s *Server) GetTool(toolName string) (tool.Tool, bool) {
	return s.toolManager.GetTool(toolName)
}

func (s *Server) ListTools() []tool.Tool {
	return s.toolManager.ListTools()
}

func (s *Server) ExecuteTool(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	return s.toolManager.ExecuteTool(ctx, req)
}

func (s *Server) SetMCPServer(mcpServer tool.MCPServerProvider) {
	s.toolManager.SetMCPServer(mcpServer)
}

func (s *Server) AddTool(t tool.Tool) error {
	return s.toolManager.AddTool(t)
}

func (s *Server) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	return s.serviceRegistry.GetServiceInfo(serviceID)
}

func (s *Server) ClearToolsForService(serviceKey string) {
	s.toolManager.ClearToolsForService(serviceKey)
}
