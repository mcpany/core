/*
 * Copyright 2025 Author(s) of MCP-XY
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

	"github.com/mcpxy/core/pkg/appconsts"
	"github.com/mcpxy/core/pkg/auth"
	"github.com/mcpxy/core/pkg/bus"
	"github.com/mcpxy/core/pkg/consts"
	"github.com/mcpxy/core/pkg/prompt"
	"github.com/mcpxy/core/pkg/resource"
	"github.com/mcpxy/core/pkg/serviceregistry"
	"github.com/mcpxy/core/pkg/tool"
	"github.com/mcpxy/core/pkg/util"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server is the core of the MCP-X application. It orchestrates the handling of
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

// Server returns the underlying *mcp.Server instance, providing access to the
// core MCP server functionality.
func (s *Server) Server() *mcp.Server {
	return s.server
}

// NewServer creates and initializes a new MCP-X Server. It sets up the
// necessary managers for tools, prompts, and resources, configures the router
// with handlers for standard MCP methods, and establishes middleware for
// request processing and validation.
//
// ctx is the application's root context.
// toolManager manages the lifecycle and access to tools.
// promptManager manages the lifecycle and access to prompts.
// resourceManager manages the lifecycle and access to resources.
// authManager handles authentication for incoming requests.
// serviceRegistry keeps track of all registered upstream services.
// bus is the event bus used for asynchronous communication between components.
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

	s.router.Register(consts.MethodPromptsList, func(ctx context.Context, req mcp.Request) (mcp.Result, error) {
		if r, ok := req.(*mcp.ListPromptsRequest); ok {
			return s.ListPrompts(ctx, r)
		}
		return nil, fmt.Errorf("invalid request type for %s", consts.MethodPromptsList)
	})

	s.router.Register(consts.MethodPromptsGet, func(ctx context.Context, req mcp.Request) (mcp.Result, error) {
		if r, ok := req.(*mcp.GetPromptRequest); ok {
			return s.GetPrompt(ctx, r)
		}
		return nil, fmt.Errorf("invalid request type for %s", consts.MethodPromptsGet)
	})

	s.router.Register(consts.MethodResourcesList, func(ctx context.Context, req mcp.Request) (mcp.Result, error) {
		if r, ok := req.(*mcp.ListResourcesRequest); ok {
			return s.ListResources(ctx, r)
		}
		return nil, fmt.Errorf("invalid request type for %s", consts.MethodResourcesList)
	})

	s.router.Register(consts.MethodResourcesRead, func(ctx context.Context, req mcp.Request) (mcp.Result, error) {
		if r, ok := req.(*mcp.ReadResourceRequest); ok {
			return s.ReadResource(ctx, r)
		}
		return nil, fmt.Errorf("invalid request type for %s", consts.MethodResourcesRead)
	})

	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    appconsts.Name,
		Version: appconsts.Version,
	}, &mcp.ServerOptions{})
	s.server = mcpServer

	s.toolManager.SetMCPServer(s)

	// TODO: Re-enable notifications when the go-sdk supports them
	// s.promptManager.OnListChanged(func() {
	// 	if s.server != nil {
	// 		s.server.Notify(consts.NotificationPromptsListChanged, nil)
	// 	}
	// })
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
					if listResult, ok := result.(*mcp.ListToolsResult); ok {
						activeTools := s.toolManager.ListTools()
						activeToolsMap := make(map[string]bool)
						for _, t := range activeTools {
							toolID, _ := util.GenerateToolID(t.Tool().GetServiceId(), t.Tool().GetName())
							activeToolsMap[toolID] = true
						}

						filteredTools := []*mcp.Tool{}
						for _, t := range listResult.Tools {
							if activeToolsMap[t.Name] {
								filteredTools = append(filteredTools, t)
							}
						}
						listResult.Tools = filteredTools
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
// available prompts from the PromptManager and returns them to the client.
func (s *Server) ListPrompts(ctx context.Context, req *mcp.ListPromptsRequest) (*mcp.ListPromptsResult, error) {
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
// arguments, returning the result.
func (s *Server) GetPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
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

// ListResources handles the "resources/list" MCP request. It fetches the list of
// available resources from the ResourceManager and returns them to the client.
func (s *Server) ListResources(ctx context.Context, req *mcp.ListResourcesRequest) (*mcp.ListResourcesResult, error) {
	resources := s.resourceManager.ListResources()
	mcpResources := make([]*mcp.Resource, len(resources))
	for i, r := range resources {
		mcpResources[i] = r.Resource()
	}
	return &mcp.ListResourcesResult{
		Resources: mcpResources,
	}, nil
}

// ReadResource handles the "resources/read" MCP request. It retrieves a specific
// resource by its URI from the ResourceManager and returns its content.
func (s *Server) ReadResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	r, ok := s.resourceManager.GetResource(req.Params.URI)
	if !ok {
		return nil, resource.ErrResourceNotFound
	}
	return r.Read(ctx)
}

// AuthManager returns the server's authentication manager.
func (s *Server) AuthManager() *auth.AuthManager {
	return s.authManager
}

// ToolManager returns the server's tool manager.
func (s *Server) ToolManager() tool.ToolManagerInterface {
	return s.toolManager
}

// PromptManager returns the server's prompt manager.
func (s *Server) PromptManager() prompt.PromptManagerInterface {
	return s.promptManager
}

// ResourceManager returns the server's resource manager.
func (s *Server) ResourceManager() resource.ResourceManagerInterface {
	return s.resourceManager
}

// ServiceRegistry returns the server's service registry.
func (s *Server) ServiceRegistry() *serviceregistry.ServiceRegistry {
	return s.serviceRegistry
}
