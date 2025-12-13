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
	"time"

	"github.com/mcpany/core/pkg/appconsts"
	"github.com/mcpany/core/pkg/auth"

	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/consts"
	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/metrics"
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
	toolManager     tool.ManagerInterface
	promptManager   prompt.ManagerInterface
	resourceManager resource.ManagerInterface
	authManager     *auth.Manager
	serviceRegistry *serviceregistry.ServiceRegistry
	bus             *bus.Provider
	reloadFunc      func() error
	debug           bool
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
	toolManager tool.ManagerInterface,
	promptManager prompt.ManagerInterface,
	resourceManager resource.ManagerInterface,
	authManager *auth.Manager,
	serviceRegistry *serviceregistry.ServiceRegistry,
	bus *bus.Provider,
	debug bool,
) (*Server, error) {
	s := &Server{
		router:          NewRouter(),
		toolManager:     toolManager,
		promptManager:   promptManager,
		resourceManager: resourceManager,
		authManager:     authManager,
		serviceRegistry: serviceRegistry,
		bus:             bus,
		debug:           debug,
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

	s.resourceManager.OnListChanged(func() {
		if s.server != nil {
			// WORKAROUND: The Go SDK (v1.1.0) does not expose a way to manually trigger
			// notifications. However, adding a resource using AddResource triggers
			// "notifications/resources/list_changed". We add a dummy resource to
			// trigger the notification. The server intercepts "resources/list", so this
			// dummy resource will not be visible to clients.
			s.server.AddResource(
				&mcp.Resource{
					Name: "internal-notification-trigger",
					URI:  "internal://notification-trigger",
				},
				func(context.Context, *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
					return nil, mcp.ResourceNotFoundError("internal://notification-trigger")
				},
			)
		}
	})

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
			if method == consts.MethodToolsList {
				// The tool manager is the authoritative source of tools. We iterate over the
				// tools in the manager to ensure that the list is always up-to-date and
				// reflects the current state of the system.
				managedTools := s.toolManager.ListTools()
				refreshedTools := make([]*mcp.Tool, 0, len(managedTools))
				for _, toolInstance := range managedTools {
					mcpTool, err := tool.ConvertProtoToMCPTool(toolInstance.Tool())
					if err != nil {
						logging.GetLogger().
							Error("Failed to convert tool to MCP format", "toolName", toolInstance.Tool().GetName(), "error", err)
						return nil, fmt.Errorf("failed to convert tool %q to MCP format: %w", toolInstance.Tool().GetName(), err)
					}
					refreshedTools = append(refreshedTools, mcpTool)
				}
				return &mcp.ListToolsResult{Tools: refreshedTools}, nil
			}
			return next(ctx, method, req)
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
func (s *Server) AuthManager() *auth.Manager {
	return s.authManager
}

// ToolManager returns the server's tool manager, which is responsible for
// managing the lifecycle and access to tools.
func (s *Server) ToolManager() tool.ManagerInterface {
	return s.toolManager
}

// PromptManager returns the server's prompt manager, which is responsible for
// managing the lifecycle and access to prompts.
func (s *Server) PromptManager() prompt.ManagerInterface {
	return s.promptManager
}

// ResourceManager returns the server's resource manager, which is responsible
// for managing the lifecycle and access to resources.
func (s *Server) ResourceManager() resource.ManagerInterface {
	return s.resourceManager
}

// ServiceRegistry returns the server's service registry, which keeps track of
// all registered upstream services.
func (s *Server) ServiceRegistry() *serviceregistry.ServiceRegistry {
	return s.serviceRegistry
}

// AddServiceInfo adds information about a service to the tool manager.
func (s *Server) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	s.toolManager.AddServiceInfo(serviceID, info)
}

// GetTool retrieves a tool by its name.
func (s *Server) GetTool(toolName string) (tool.Tool, bool) {
	return s.toolManager.GetTool(toolName)
}

// ListTools returns a list of all available tools.
func (s *Server) ListTools() []tool.Tool {
	logging.GetLogger().Info("Listing tools...")
	metrics.IncrCounter([]string{"tools", "list", "total"}, 1)
	return s.toolManager.ListTools()
}

// CallTool executes a tool with the provided request.
func (s *Server) CallTool(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	logging.GetLogger().Info("Calling tool...", "toolName", req.ToolName)
	metrics.IncrCounter([]string{"tools", "call", "total"}, 1)
	metrics.IncrCounter([]string{"tool", req.ToolName, "call", "total"}, 1)
	startTime := time.Now()
	defer metrics.MeasureSince([]string{"tool", req.ToolName, "call", "latency"}, startTime)
	defer metrics.MeasureSince([]string{"tools", "call", "latency"}, startTime)

	result, err := s.toolManager.ExecuteTool(ctx, req)
	if err != nil {
		metrics.IncrCounter([]string{"tools", "call", "errors"}, 1)
		metrics.IncrCounter([]string{"tool", req.ToolName, "call", "errors"}, 1)
	}
	return result, err
}

// SetMCPServer sets the MCP server provider for the tool manager.
func (s *Server) SetMCPServer(mcpServer tool.MCPServerProvider) {
	s.toolManager.SetMCPServer(mcpServer)
}

// AddTool registers a new tool with the tool manager.
func (s *Server) AddTool(t tool.Tool) error {
	return s.toolManager.AddTool(t)
}

// GetServiceInfo retrieves information about a service by its ID.
func (s *Server) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	return s.toolManager.GetServiceInfo(serviceID)
}

// ClearToolsForService removes all tools associated with a specific service.
func (s *Server) ClearToolsForService(serviceKey string) {
	s.toolManager.ClearToolsForService(serviceKey)
}

// SetReloadFunc sets the function to be called when a configuration reload is triggered.
func (s *Server) SetReloadFunc(f func() error) {
	s.reloadFunc = f
}

// Reload reloads the server's configuration and updates its state.
func (s *Server) Reload() error {
	if s.reloadFunc != nil {
		return s.reloadFunc()
	}
	return nil
}
