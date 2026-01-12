// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"
	"fmt"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/mcpany/core/server/pkg/appconsts"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/consts"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/metrics"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// AddReceivingMiddlewareHook is a hook for adding receiving middleware.
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
	reloadFunc      func(context.Context) error
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
	_ context.Context,
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

	s.router.Register(
		consts.MethodToolsCall,
		func(ctx context.Context, req mcp.Request) (mcp.Result, error) {
			if r, ok := req.(*mcp.CallToolRequest); ok {
				execReq := &tool.ExecutionRequest{
					ToolName:   r.Params.Name,
					ToolInputs: r.Params.Arguments,
				}

				session := req.GetSession()
				if serverSession, ok := session.(*mcp.ServerSession); ok {
					mcpSession := NewMCPSession(serverSession)
					ctx = tool.NewContextWithSession(ctx, mcpSession)
				}

				res, err := s.CallTool(ctx, execReq)
				if err != nil {
					return &mcp.CallToolResult{
						Content: []mcp.Content{
							&mcp.TextContent{
								Text: fmt.Sprintf("Tool execution failed: %v", err),
							},
						},
						IsError: true,
					}, nil
				}
				if result, ok := res.(mcp.Result); ok {
					return result, nil
				}

				// Fallback for other types (string, []byte, etc.)
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: util.ToString(res),
						},
					},
				}, nil
			}
			return nil, fmt.Errorf("invalid request type for %s", consts.MethodToolsCall)
		},
	)

	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    appconsts.Name,
		Version: appconsts.Version,
	}, &mcp.ServerOptions{
		HasPrompts:   true,
		HasTools:     true,
		HasResources: true,
	})
	s.server = mcpServer

	s.toolManager.SetMCPServer(s)
	s.promptManager.SetMCPServer(prompt.NewMCPServerProvider(s.Server()))

	// Register built-in tools
	if err := s.toolManager.AddTool(NewRootsTool()); err != nil {
		// Log error but don't fail startup if duplicate (e.g. reload)
		// Assuming logging is initialized
		logging.GetLogger().Error("Failed to register built-in tools", "error", err)
	}

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

	// Register DLP middleware
	// Note: config.GlobalSettings() returns *configv1.GlobalSettings
	s.server.AddReceivingMiddleware(middleware.DLPMiddleware(config.GlobalSettings().GetDlp(), logging.GetLogger()))

	s.server.AddReceivingMiddleware(s.routerMiddleware)
	s.server.AddReceivingMiddleware(s.toolListFilteringMiddleware)
	s.server.AddReceivingMiddleware(s.resourceListFilteringMiddleware)
	s.server.AddReceivingMiddleware(s.promptListFilteringMiddleware)

	return s, nil
}

func (s *Server) routerMiddleware(next mcp.MethodHandler) mcp.MethodHandler {
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

func (s *Server) toolListFilteringMiddleware(next mcp.MethodHandler) mcp.MethodHandler {
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


			profileID, _ := auth.ProfileIDFromContext(ctx)

			for _, toolInstance := range managedTools {
				// Profile-based filtering
				if profileID != "" {
					serviceID := toolInstance.Tool().GetServiceId()
					if !s.toolManager.IsServiceAllowed(serviceID, profileID) {
						continue
					}
				}

				mcpTool := toolInstance.MCPTool()
				if mcpTool != nil {
					refreshedTools = append(refreshedTools, mcpTool)
				} else {
					logging.GetLogger().
						Error("Failed to convert tool to MCP format", "toolName", toolInstance.Tool().GetName())
					// We continue instead of failing the whole request.
				}
			}
			return &mcp.ListToolsResult{Tools: refreshedTools}, nil
		}
		return next(ctx, method, req)
	}
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
	mcpPrompts := make([]*mcp.Prompt, 0, len(prompts))
	for _, p := range prompts {
		if prompt := p.Prompt(); prompt != nil {
			mcpPrompts = append(mcpPrompts, prompt)
		}
	}
	return &mcp.ListPromptsResult{
		Prompts: mcpPrompts,
	}, nil
}

// CreateMessage requests a message creation from the client (sampling).
// This method exposes sampling to the Server instance if a session is available.
// Note: In a stateless request context without a persistent session, this might fail.
func (s *Server) CreateMessage(ctx context.Context, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error) {
	// Attempt to retrieve session from context, which is populated during request handling
	if session, ok := tool.GetSession(ctx); ok {
		return session.CreateMessage(ctx, params)
	}
	return nil, fmt.Errorf("no active session found in context")
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

	// Use json-iterator for faster JSON marshaling
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
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
	mcpResources := make([]*mcp.Resource, 0, len(resources))
	for _, r := range resources {
		if resource := r.Resource(); resource != nil {
			mcpResources = append(mcpResources, resource)
		}
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
//
// Returns:
//   - The authentication manager instance.
func (s *Server) AuthManager() *auth.Manager {
	return s.authManager
}

// ToolManager returns the server's tool manager, which is responsible for
// managing the lifecycle and access to tools.
//
// Returns:
//   - The tool manager interface.
func (s *Server) ToolManager() tool.ManagerInterface {
	return s.toolManager
}

// PromptManager returns the server's prompt manager, which is responsible for
// managing the lifecycle and access to prompts.
//
// Returns:
//   - The prompt manager interface.
func (s *Server) PromptManager() prompt.ManagerInterface {
	return s.promptManager
}

// ResourceManager returns the server's resource manager, which is responsible
// for managing the lifecycle and access to resources.
//
// Returns:
//   - The resource manager interface.
func (s *Server) ResourceManager() resource.ManagerInterface {
	return s.resourceManager
}

// ServiceRegistry returns the server's service registry, which keeps track of
// all registered upstream services.
//
// Returns:
//   - The service registry instance.
func (s *Server) ServiceRegistry() *serviceregistry.ServiceRegistry {
	return s.serviceRegistry
}

// AddServiceInfo adds information about a service to the tool manager.
//
// Parameters:
//   - serviceID: The unique identifier of the service.
//   - info: The service information to add.
func (s *Server) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	s.toolManager.AddServiceInfo(serviceID, info)
}

// GetTool retrieves a tool by its name.
//
// Parameters:
//   - toolName: The name of the tool to retrieve.
//
// Returns:
//   - The tool instance if found.
//   - A boolean indicating whether the tool was found.
func (s *Server) GetTool(toolName string) (tool.Tool, bool) {
	return s.toolManager.GetTool(toolName)
}

// ListTools returns a list of all available tools.
//
// Returns:
//   - A slice of all available tools.
func (s *Server) ListTools() []tool.Tool {
	logging.GetLogger().Info("Listing tools...")
	metrics.IncrCounter([]string{"tools", "list", "total"}, 1)
	return s.toolManager.ListTools()
}

// CallTool executes a tool with the provided request.
//
// Parameters:
//   - ctx: The context for the execution.
//   - req: The execution request containing tool name and arguments.
//
// Returns:
//   - The result of the tool execution.
//   - An error if the tool execution fails or access is denied.
func (s *Server) CallTool(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	logging.GetLogger().Info("Calling tool...", "toolName", req.ToolName, "arguments", string(util.RedactJSON(req.ToolInputs)))
	// Try to get service ID from tool
	var serviceID string
	if t, ok := s.GetTool(req.ToolName); ok {
		if t.Tool() != nil {
			serviceID = t.Tool().GetServiceId()
		}
	}

	profileID, _ := auth.ProfileIDFromContext(ctx)
	if profileID != "" && serviceID != "" {
		if !s.toolManager.IsServiceAllowed(serviceID, profileID) {
			logging.GetLogger().Warn("Access denied to tool by profile", "toolName", req.ToolName, "profileID", profileID)
			return nil, fmt.Errorf("access denied to tool %q", req.ToolName)
		}
	}

	metrics.IncrCounterWithLabels([]string{"tools", "call", "total"}, 1, []metrics.Label{
		{Name: "tool", Value: req.ToolName},
		{Name: "service_id", Value: serviceID},
	})
	startTime := time.Now()
	defer func() {
		// Use AddSampleWithLabels directly to avoid emitting an unlabelled metric (which MeasureSinceWithLabels does).
		// MeasureSince emits in milliseconds.
		duration := float32(time.Since(startTime).Seconds() * 1000)
		metrics.AddSampleWithLabels([]string{"tools", "call", "latency"}, duration, []metrics.Label{
			{Name: "tool", Value: req.ToolName},
			{Name: "service_id", Value: serviceID},
		})
	}()

	result, err := s.toolManager.ExecuteTool(ctx, req)
	if err != nil {
		metrics.IncrCounterWithLabels([]string{"tools", "call", "errors"}, 1, []metrics.Label{
			{Name: "tool", Value: req.ToolName},
			{Name: "service_id", Value: serviceID},
		})
	}
	logging.GetLogger().Info("Tool execution completed", "result_type", fmt.Sprintf("%T", result), "result_value", result)

	if err != nil {
		return nil, err
	}

	if ctr, ok := result.(*mcp.CallToolResult); ok {
		return ctr, nil
	}

	// Handle map[string]any result (e.g. from HTTP tools)
	var jsonBytes []byte
	var marshalErr error

	// Use json-iterator for faster JSON operations
	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	if resultMap, ok := result.(map[string]any); ok {
		// Heuristic: If map looks like CallToolResult (has "content" or "isError"),
		// try to parse it as such.
		_, hasContent := resultMap["content"]
		_, hasIsError := resultMap["isError"]

		if hasContent || hasIsError {
			// Convert map to CallToolResult via JSON
			jsonBytes, marshalErr = json.Marshal(resultMap)
			if marshalErr != nil {
				return nil, fmt.Errorf("failed to marshal tool result map: %w", marshalErr)
			}

			var callToolRes mcp.CallToolResult
			if err := json.Unmarshal(jsonBytes, &callToolRes); err == nil {
				return &callToolRes, nil
			}
			// If unmarshal fails (e.g. content is string instead of array), fall through to default behavior
			// and treat it as raw data.
			logging.GetLogger().Warn("Failed to unmarshal potential CallToolResult map, treating as raw data", "toolName", req.ToolName)

			// Heuristic: If content is a string, wrap it in a TextContent
			if content, ok := resultMap["content"].(string); ok {
				isError := callToolRes.IsError
				// If IsError is false, it might be because Unmarshal failed (defaulting to false).
				// Check the raw map to see if it was explicitly set.
				if val, ok := resultMap["isError"].(bool); ok {
					isError = val
				}

				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: content,
						},
					},
					IsError: isError,
				}, nil
			}
		}
	}

	// Default to JSON encoding for the result
	if jsonBytes == nil {
		jsonBytes, marshalErr = json.Marshal(result)
	}
	text := string(jsonBytes)
	if marshalErr != nil {
		text = util.ToString(result)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: text,
			},
		},
	}, nil
}

// SetMCPServer sets the MCP server provider for the tool manager.
//
// Parameters:
//   - mcpServer: The MCP server provider to set.
func (s *Server) SetMCPServer(mcpServer tool.MCPServerProvider) {
	s.toolManager.SetMCPServer(mcpServer)
}

// AddTool registers a new tool with the tool manager.
//
// Parameters:
//   - t: The tool instance to register.
//
// Returns:
//   - An error if the tool cannot be added (e.g., if it already exists).
func (s *Server) AddTool(t tool.Tool) error {
	return s.toolManager.AddTool(t)
}

// GetServiceInfo retrieves information about a service by its ID.
//
// Parameters:
//   - serviceID: The unique identifier of the service.
//
// Returns:
//   - A pointer to the ServiceInfo if found.
//   - A boolean indicating whether the service was found.
func (s *Server) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	return s.toolManager.GetServiceInfo(serviceID)
}

// isServiceAllowed checked if the current user profile had access to the service.
// This logic is removed as we moved to centralized profile management where loaded services are implicit.
// If we need RBAC, it should be a separate concern.

// ClearToolsForService removes all tools associated with a specific service.
//
// Parameters:
//   - serviceKey: The identifier of the service whose tools should be cleared.
func (s *Server) ClearToolsForService(serviceKey string) {
	s.toolManager.ClearToolsForService(serviceKey)
}

// SetReloadFunc sets the function to be called when a configuration reload is triggered.
//
// Parameters:
//   - f: The function to execute on reload.
func (s *Server) SetReloadFunc(f func(context.Context) error) {
	s.reloadFunc = f
}

// Reload reloads the server's configuration and updates its state.
//
// Parameters:
//   - ctx: The context for the reload operation.
//
// Returns:
//   - An error if the reload function fails.
func (s *Server) Reload(ctx context.Context) error {
	if s.reloadFunc != nil {
		return s.reloadFunc(ctx)
	}
	return nil
}

func (s *Server) resourceListFilteringMiddleware(next mcp.MethodHandler) mcp.MethodHandler {
	return func(
		ctx context.Context,
		method string,
		req mcp.Request,
	) (mcp.Result, error) {
		if method == consts.MethodResourcesList {
			managedResources := s.resourceManager.ListResources()
			refreshedResources := make([]*mcp.Resource, 0, len(managedResources))


			profileID, _ := auth.ProfileIDFromContext(ctx)

			for _, resourceInstance := range managedResources {
				// Profile filtering
				if profileID != "" {
					if !s.toolManager.IsServiceAllowed(resourceInstance.Service(), profileID) {
						continue
					}
				}

				if res := resourceInstance.Resource(); res != nil {
					refreshedResources = append(refreshedResources, res)
				}
			}
			return &mcp.ListResourcesResult{Resources: refreshedResources}, nil
		}
		return next(ctx, method, req)
	}
}

func (s *Server) promptListFilteringMiddleware(next mcp.MethodHandler) mcp.MethodHandler {
	return func(
		ctx context.Context,
		method string,
		req mcp.Request,
	) (mcp.Result, error) {
		if method == consts.MethodPromptsList {
			managedPrompts := s.promptManager.ListPrompts()
			refreshedPrompts := make([]*mcp.Prompt, 0, len(managedPrompts))


			profileID, _ := auth.ProfileIDFromContext(ctx)

			for _, promptInstance := range managedPrompts {
				// Profile filtering
				if profileID != "" {
					if !s.toolManager.IsServiceAllowed(promptInstance.Service(), profileID) {
						continue
					}
				}

				if prompt := promptInstance.Prompt(); prompt != nil {
					refreshedPrompts = append(refreshedPrompts, prompt)
				}
			}
			return &mcp.ListPromptsResult{Prompts: refreshedPrompts}, nil
		}
		return next(ctx, method, req)
	}
}
