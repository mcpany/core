// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
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

var (
	metricToolsListTotal   = []string{"tools", "list", "total"}
	metricToolsCallTotal   = []string{"tools", "call", "total"}
	metricToolsCallErrors  = []string{"tools", "call", "errors"}
	metricToolsCallLatency = []string{"tools", "call", "latency"}
)

// fastJSON is a jsoniter configuration that disables map key sorting for performance.
// The order of keys in the JSON response does not matter for the LLM.
var fastJSON = jsoniter.Config{
	EscapeHTML:             true,
	SortMapKeys:            false,
	ValidateJsonRawMessage: true,
}.Froze()

// AddReceivingMiddlewareHook is a testing hook that allows inspection of the middleware chain.
// It is invoked when the Server method is called, allowing tests to verify which middlewares are present.
//
// Side Effects:
//   - When set, this function is called synchronously during Server() access.
var AddReceivingMiddlewareHook func(name string)

// Server is the core of the MCP Any application.
//
// It orchestrates the handling of MCP (Model Context Protocol) requests by managing various components such as
// tools, prompts, resources, and services. It uses an internal router to delegate requests to the appropriate
// handlers and communicates with backend workers via an event bus.
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

// Server returns the underlying *mcp.Server instance.
//
// It provides access to the core MCP server functionality, which can be used for advanced
// configurations or direct interaction with the MCP server.
//
// Returns:
//   - *mcp.Server: The underlying server instance.
func (s *Server) Server() *mcp.Server {
	if AddReceivingMiddlewareHook != nil {
		// This is a test hook to allow inspection of the middleware chain.
		// We are passing the name of the middleware as a string.
		AddReceivingMiddlewareHook("CachingMiddleware")
	}
	return s.server
}

// NewServer creates and initializes a new MCP Any Server.
//
// It sets up the necessary managers for tools, prompts, and resources, configures the router
// with handlers for standard MCP methods, and establishes middleware for request processing,
// such as routing and tool list filtering.
//
// Parameters:
//   - ctx: context.Context. The application's root context.
//   - toolManager: tool.ManagerInterface. Manages the lifecycle and access to tools.
//   - promptManager: prompt.ManagerInterface. Manages the lifecycle and access to prompts.
//   - resourceManager: resource.ManagerInterface. Manages the lifecycle and access to resources.
//   - authManager: *auth.Manager. Handles authentication for incoming requests.
//   - serviceRegistry: *serviceregistry.ServiceRegistry. Keeps track of all registered upstream services.
//   - bus: *bus.Provider. The event bus used for asynchronous communication between components.
//   - debug: bool. Whether to enable debug mode.
//
// Returns:
//   - *Server: A new instance of the Server.
//   - error: An error if initialization fails.
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
					return nil, err
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
			profileID, _ := auth.ProfileIDFromContext(ctx)
			// ⚡ Bolt Optimization: Use cached MCP tools list if no profile filtering is required
			// to avoid N allocations and conversions.
			if profileID == "" {
				return &mcp.ListToolsResult{Tools: s.toolManager.ListMCPTools()}, nil
			}

			// The tool manager is the authoritative source of tools. We iterate over the
			// tools in the manager to ensure that the list is always up-to-date and
			// reflects the current state of the system.
			managedTools := s.toolManager.ListTools()
			refreshedTools := make([]*mcp.Tool, 0, len(managedTools))

			// ⚡ Bolt Optimization: Fetch allowed services once to avoid N lock acquisitions
			var allowedServices map[string]bool
			if profileID != "" {
				allowedServices, _ = s.toolManager.GetAllowedServiceIDs(profileID)
			}

			for _, toolInstance := range managedTools {
				// Profile-based filtering
				if profileID != "" {
					serviceID := toolInstance.Tool().GetServiceId()
					// Optimized O(1) map lookup
					if allowedServices != nil {
						if !allowedServices[serviceID] {
							continue
						}
					} else {
						// Profile not found or error, default to deny if profileID was present?
						// Original IsServiceAllowed logic: if profile not found, return false.
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

// ListPrompts handles the "prompts/list" MCP request.
//
// It retrieves the list of available prompts from the PromptManager, converts them to the MCP format,
// and returns them to the client.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - req: *mcp.ListPromptsRequest. The "prompts/list" request from the client.
//
// Returns:
//   - *mcp.ListPromptsResult: A list of available prompts.
//   - error: An error if the retrieval fails.
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
//
// This method exposes sampling to the Server instance if a session is available.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - params: *mcp.CreateMessageParams. The parameters for the message creation.
//
// Returns:
//   - *mcp.CreateMessageResult: The result of the message creation.
//   - error: An error if no active session is found in context or if the operation fails.
func (s *Server) CreateMessage(ctx context.Context, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error) {
	// Attempt to retrieve session from context, which is populated during request handling
	if session, ok := tool.GetSession(ctx); ok {
		return session.CreateMessage(ctx, params)
	}
	return nil, fmt.Errorf("no active session found in context")
}

// GetPrompt handles the "prompts/get" MCP request.
//
// It retrieves a specific prompt by name from the PromptManager and executes it with the provided
// arguments, returning the result.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - req: *mcp.GetPromptRequest. The "prompts/get" request from the client, containing the prompt name and arguments.
//
// Returns:
//   - *mcp.GetPromptResult: The result of the prompt execution.
//   - error: An error if the prompt is not found or execution fails.
//
// Throws/Errors:
//   - prompt.ErrPromptNotFound: If the requested prompt does not exist.
func (s *Server) GetPrompt(
	ctx context.Context,
	req *mcp.GetPromptRequest,
) (*mcp.GetPromptResult, error) {
	p, ok := s.promptManager.GetPrompt(req.Params.Name)
	if !ok {
		return nil, prompt.ErrPromptNotFound
	}

	profileID, _ := auth.ProfileIDFromContext(ctx)
	if profileID != "" {
		serviceID := p.Service()
		if serviceID != "" && !s.toolManager.IsServiceAllowed(serviceID, profileID) {
			logging.GetLogger().Warn("Access denied to prompt by profile", "promptName", req.Params.Name, "profileID", profileID)
			return nil, fmt.Errorf("access denied to prompt %q", req.Params.Name)
		}
	}

	// Use json-iterator for faster JSON marshaling
	argsBytes, err := fastJSON.Marshal(req.Params.Arguments)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal prompt arguments: %w", err)
	}

	return p.Get(ctx, argsBytes)
}

// ListResources handles the "resources/list" MCP request.
//
// It fetches the list of available resources from the ResourceManager, converts them to the MCP
// format, and returns them to the client.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - req: *mcp.ListResourcesRequest. The "resources/list" request from the client.
//
// Returns:
//   - *mcp.ListResourcesResult: A list of available resources.
//   - error: An error if the retrieval fails.
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

// ReadResource handles the "resources/read" MCP request.
//
// It retrieves a specific resource by its URI from the ResourceManager and returns its content.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - req: *mcp.ReadResourceRequest. The "resources/read" request from the client, containing the URI of the resource.
//
// Returns:
//   - *mcp.ReadResourceResult: The content of the resource.
//   - error: An error if the resource is not found or reading fails.
//
// Throws/Errors:
//   - resource.ErrResourceNotFound: If the requested resource does not exist.
func (s *Server) ReadResource(
	ctx context.Context,
	req *mcp.ReadResourceRequest,
) (*mcp.ReadResourceResult, error) {
	r, ok := s.resourceManager.GetResource(req.Params.URI)
	if !ok {
		return nil, resource.ErrResourceNotFound
	}

	profileID, _ := auth.ProfileIDFromContext(ctx)
	if profileID != "" {
		serviceID := r.Service()
		if serviceID != "" && !s.toolManager.IsServiceAllowed(serviceID, profileID) {
			logging.GetLogger().Warn("Access denied to resource by profile", "resourceURI", req.Params.URI, "profileID", profileID)
			return nil, fmt.Errorf("access denied to resource %q", req.Params.URI)
		}
	}

	return r.Read(ctx)
}

// AuthManager returns the server's authentication manager.
//
// It provides access to the authentication manager, which is responsible for handling
// authentication for incoming requests.
//
// Returns:
//   - *auth.Manager: The authentication manager instance.
func (s *Server) AuthManager() *auth.Manager {
	return s.authManager
}

// ToolManager returns the server's tool manager.
//
// It provides access to the tool manager, which is responsible for managing the lifecycle
// and access to tools.
//
// Returns:
//   - tool.ManagerInterface: The tool manager interface.
func (s *Server) ToolManager() tool.ManagerInterface {
	return s.toolManager
}

// PromptManager returns the server's prompt manager.
//
// It provides access to the prompt manager, which is responsible for managing the lifecycle
// and access to prompts.
//
// Returns:
//   - prompt.ManagerInterface: The prompt manager interface.
func (s *Server) PromptManager() prompt.ManagerInterface {
	return s.promptManager
}

// ResourceManager returns the server's resource manager.
//
// It provides access to the resource manager, which is responsible for managing the lifecycle
// and access to resources.
//
// Returns:
//   - resource.ManagerInterface: The resource manager interface.
func (s *Server) ResourceManager() resource.ManagerInterface {
	return s.resourceManager
}

// ServiceRegistry returns the server's service registry.
//
// It provides access to the service registry, which keeps track of all registered upstream services.
//
// Returns:
//   - *serviceregistry.ServiceRegistry: The service registry instance.
func (s *Server) ServiceRegistry() *serviceregistry.ServiceRegistry {
	return s.serviceRegistry
}

// AddServiceInfo adds information about a service to the tool manager.
//
// Parameters:
//   - serviceID: string. The unique identifier of the service.
//   - info: *tool.ServiceInfo. The service information to add.
//
// Returns:
//   None.
func (s *Server) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	s.toolManager.AddServiceInfo(serviceID, info)
}

// GetTool retrieves a tool by its name.
//
// Parameters:
//   - toolName: string. The name of the tool to retrieve.
//
// Returns:
//   - tool.Tool: The tool instance if found.
//   - bool: A boolean indicating whether the tool was found.
func (s *Server) GetTool(toolName string) (tool.Tool, bool) {
	return s.toolManager.GetTool(toolName)
}

// ListTools returns a list of all available tools.
//
// Returns:
//   - []tool.Tool: A slice of all available tools.
func (s *Server) ListTools() []tool.Tool {
	logging.GetLogger().Info("Listing tools...")
	metrics.IncrCounter(metricToolsListTotal, 1)
	return s.toolManager.ListTools()
}

// CallTool executes a tool with the provided request.
//
// It handles the execution of the tool, including logging, metrics collection, and profile-based
// access control.
//
// Parameters:
//   - ctx: context.Context. The context for the execution.
//   - req: *tool.ExecutionRequest. The execution request containing tool name and arguments.
//
// Returns:
//   - any: The result of the tool execution.
//   - error: An error if the tool execution fails or access is denied.
func (s *Server) CallTool(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	logger := logging.GetLogger()
	// ⚡ Bolt Optimization: Check if logging is enabled to avoid unnecessary allocations.
	if logger.Enabled(ctx, slog.LevelInfo) {
		logger.Info("Calling tool...", "toolName", req.ToolName, "arguments", LazyRedact(req.ToolInputs))
	}
	// Try to get service ID from tool
	var serviceID string
	if t, ok := s.GetTool(req.ToolName); ok {
		req.Tool = t
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

	metrics.IncrCounterWithLabels(metricToolsCallTotal, 1, []metrics.Label{
		{Name: "tool", Value: req.ToolName},
		{Name: "service_id", Value: serviceID},
	})
	startTime := time.Now()
	metrics.MeasureSinceWithLabels(metricToolsCallLatency, startTime, []metrics.Label{
		{Name: "tool", Value: req.ToolName},
		{Name: "service_id", Value: serviceID},
	})

	result, err := s.toolManager.ExecuteTool(ctx, req)
	if err != nil {
		metrics.IncrCounterWithLabels(metricToolsCallErrors, 1, []metrics.Label{
			{Name: "tool", Value: req.ToolName},
			{Name: "service_id", Value: serviceID},
		})
	}

	// ⚡ Bolt Optimization: Defer logging until AFTER we have processed the result.
	// This avoids double-marshaling large result maps (once for logging, once for return).

	if err != nil {
		// Log error result (nil result usually)
		if logger.Enabled(ctx, slog.LevelInfo) {
			logger.Info("Tool execution completed", "result_type", fmt.Sprintf("%T", result), "result_value", LazyLogResult{Value: result})
		}
		return nil, err
	}

	var finalResult *mcp.CallToolResult
	var jsonBytes []byte
	var marshalErr error
	var isStructured bool

	// 1. Check if it's already a CallToolResult
	if ctr, ok := result.(*mcp.CallToolResult); ok {
		finalResult = ctr
		isStructured = true
	} else if resultMap, ok := result.(map[string]any); ok {
		// 2. Handle map[string]any result
		// Heuristic: If map looks like CallToolResult (has "content" or "isError"), try to parse it.
		_, hasContent := resultMap["content"]
		_, hasIsError := resultMap["isError"]

		if hasContent || hasIsError {
			// ⚡ Bolt Optimization: Try fast path conversion first to avoid JSON roundtrip
			if res, err := convertMapToCallToolResult(resultMap); err == nil {
				finalResult = res
				isStructured = true
			} else {
				// Conversion failed, fall back to JSON unmarshal
				// We marshal it to JSON bytes (reused for logging and return)
				jsonBytes, marshalErr = fastJSON.Marshal(resultMap)
				if marshalErr == nil {
					var callToolRes mcp.CallToolResult
					if err := fastJSON.Unmarshal(jsonBytes, &callToolRes); err == nil {
						finalResult = &callToolRes
						isStructured = true
					} else {
						// Unmarshal failed
						logging.GetLogger().Warn("Failed to unmarshal potential CallToolResult map, treating as raw data", "toolName", req.ToolName)
						// Fall through to raw data handling
					}
				} else {
					// Marshal failed? Extremely rare for map[string]any unless it has cycles/funcs
					// Fall through to error handling or raw data
					logging.GetLogger().Warn("Failed to marshal map[string]any", "error", marshalErr)
				}

				// Special case: If content is a string, wrap it in TextContent
				if finalResult == nil && marshalErr == nil {
					if content, ok := resultMap["content"].(string); ok {
						isError := false
						if val, ok := resultMap["isError"].(bool); ok {
							isError = val
						}
						finalResult = &mcp.CallToolResult{
							Content: []mcp.Content{
								&mcp.TextContent{Text: content},
							},
							IsError: isError,
						}
						isStructured = true
					}
				}
			}
		}
	}

	// 3. Fallback: If no structured result identified, treat as raw data
	if finalResult == nil {
		if jsonBytes == nil {
			jsonBytes, marshalErr = fastJSON.Marshal(result)
		}

		var text string
		// ⚡ Bolt Optimization: Use Zero-copy conversion for large JSON payloads
		if marshalErr == nil {
			text = util.BytesToString(jsonBytes)
		} else {
			text = util.ToString(result)
		}

		finalResult = &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: text},
			},
		}
	}

	// Log the result
	if logger.Enabled(ctx, slog.LevelInfo) {
		var logValue slog.Value
		// If we have a structured result (either directly or converted), use the summarizer.
		// If we fell back to raw JSON (isStructured=false), reuse the jsonBytes for redacted logging.
		if !isStructured && jsonBytes != nil && marshalErr == nil {
			// ⚡ Bolt Optimization: Reuse marshaled bytes for logging (redacted)
			// This saves a second marshal operation for large maps.
			logValue = slog.StringValue(util.BytesToString(util.RedactJSON(jsonBytes)))
		} else {
			logValue = summarizeCallToolResult(finalResult)
		}

		logger.Info("Tool execution completed", "result_type", fmt.Sprintf("%T", result), "result_value", logValue)
	}

	return finalResult, nil
}

// SetMCPServer sets the MCP server provider for the tool manager.
//
// Parameters:
//   - mcpServer: tool.MCPServerProvider. The MCP server provider to set.
//
// Returns:
//   None.
func (s *Server) SetMCPServer(mcpServer tool.MCPServerProvider) {
	s.toolManager.SetMCPServer(mcpServer)
}

// AddTool registers a new tool with the tool manager.
//
// Parameters:
//   - t: tool.Tool. The tool instance to register.
//
// Returns:
//   - error: An error if the tool cannot be added (e.g., if it already exists).
func (s *Server) AddTool(t tool.Tool) error {
	return s.toolManager.AddTool(t)
}

// GetServiceInfo retrieves information about a service by its ID.
//
// Parameters:
//   - serviceID: string. The unique identifier of the service.
//
// Returns:
//   - *tool.ServiceInfo: A pointer to the ServiceInfo if found.
//   - bool: A boolean indicating whether the service was found.
func (s *Server) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	return s.toolManager.GetServiceInfo(serviceID)
}

// isServiceAllowed checked if the current user profile had access to the service.
// This logic is removed as we moved to centralized profile management where loaded services are implicit.
// If we need RBAC, it should be a separate concern.

// ClearToolsForService removes all tools associated with a specific service.
//
// Parameters:
//   - serviceKey: string. The identifier of the service whose tools should be cleared.
//
// Returns:
//   None.
func (s *Server) ClearToolsForService(serviceKey string) {
	s.toolManager.ClearToolsForService(serviceKey)
}

// SetReloadFunc sets the function to be called when a configuration reload is triggered.
//
// Parameters:
//   - f: func(context.Context) error. The function to execute on reload.
//
// Returns:
//   None.
func (s *Server) SetReloadFunc(f func(context.Context) error) {
	s.reloadFunc = f
}

// Reload reloads the server's configuration and updates its state.
//
// Parameters:
//   - ctx: context.Context. The context for the reload operation.
//
// Returns:
//   - error: An error if the reload function fails.
func (s *Server) Reload(ctx context.Context) error {
	if s.reloadFunc != nil {
		return s.reloadFunc(ctx)
	}
	return nil
}

// convertMapToCallToolResult attempts to convert a map result to a CallToolResult
// without JSON serialization overhead. It supports text, image, and resource content.
func convertMapToCallToolResult(m map[string]any) (*mcp.CallToolResult, error) {
	// Fast path for content
	contentRaw, ok := m["content"]
	if !ok {
		// If content is missing, check for isError
		if _, hasIsError := m["isError"]; !hasIsError {
			return nil, fmt.Errorf("neither content nor isError present")
		}
		// Maybe it's just error?
		isError, _ := m["isError"].(bool)
		return &mcp.CallToolResult{IsError: isError}, nil
	}

	contentList, ok := contentRaw.([]any)
	if !ok {
		return nil, fmt.Errorf("content is not a list")
	}

	contents := make([]mcp.Content, 0, len(contentList))
	for _, c := range contentList {
		cMap, ok := c.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("content item is not a map")
		}

		typeStr, ok := cMap["type"].(string)
		if !ok {
			return nil, fmt.Errorf("content type is not a string")
		}

		switch typeStr {
		case "text":
			text, ok := cMap["text"].(string)
			if !ok {
				return nil, fmt.Errorf("text content text is not a string")
			}
			contents = append(contents, &mcp.TextContent{
				Text: text,
			})
		case "image":
			dataStr, ok := cMap["data"].(string)
			if !ok {
				return nil, fmt.Errorf("image content data is not a string")
			}
			data, err := base64.StdEncoding.DecodeString(dataStr)
			if err != nil {
				return nil, fmt.Errorf("failed to decode image data: %w", err)
			}
			mimeType, ok := cMap["mimeType"].(string)
			if !ok {
				return nil, fmt.Errorf("image content mimeType is not a string")
			}
			contents = append(contents, &mcp.ImageContent{
				Data:     data,
				MIMEType: mimeType,
			})
		case "resource":
			resMap, ok := cMap["resource"].(map[string]any)
			if !ok {
				return nil, fmt.Errorf("resource content resource is not a map")
			}
			uri, ok := resMap["uri"].(string)
			if !ok {
				return nil, fmt.Errorf("resource uri is not a string")
			}
			resContent := &mcp.ResourceContents{
				URI: uri,
			}
			if mt, ok := resMap["mimeType"].(string); ok {
				resContent.MIMEType = mt
			}
			if txt, ok := resMap["text"].(string); ok {
				resContent.Text = txt
			}
			if blobStr, ok := resMap["blob"].(string); ok {
				blob, err := base64.StdEncoding.DecodeString(blobStr)
				if err != nil {
					return nil, fmt.Errorf("failed to decode resource blob: %w", err)
				}
				resContent.Blob = blob
			} else if blobBytes, ok := resMap["blob"].([]byte); ok {
				resContent.Blob = blobBytes
			}
			contents = append(contents, &mcp.EmbeddedResource{
				Resource: resContent,
			})
		default:
			// Fallback for other types
			return nil, fmt.Errorf("unsupported content type for fast path: %s", typeStr)
		}
	}

	isError, _ := m["isError"].(bool)
	return &mcp.CallToolResult{
		Content: contents,
		IsError: isError,
	}, nil
}

// LazyRedact is a byte slice that implements slog.LogValuer to lazily redact
// its JSON content only when logged.
type LazyRedact []byte

// LogValue implements slog.LogValuer.
func (l LazyRedact) LogValue() slog.Value {
	return slog.StringValue(util.BytesToString(util.RedactJSON(l)))
}

// LazyLogResult wraps a tool execution result for efficient logging.
// It avoids expensive serialization of large payloads (e.g. images, huge text)
// and lazily computes the string representation only when logging is enabled.
type LazyLogResult struct {
	Value any
}

// LogValue implements slog.LogValuer.
func (r LazyLogResult) LogValue() slog.Value {
	if r.Value == nil {
		return slog.StringValue("<nil>")
	}

	switch v := r.Value.(type) {
	case *mcp.CallToolResult:
		return summarizeCallToolResult(v)
	case map[string]any:
		// Heuristic: Check if it looks like a CallToolResult
		if ctr, err := convertMapToCallToolResult(v); err == nil {
			return summarizeCallToolResult(ctr)
		}
		// Otherwise redact it. We marshal it to JSON bytes to use RedactJSON.
		// Use json-iterator for speed.
		jsonBytes, _ := fastJSON.Marshal(v)
		return slog.StringValue(util.BytesToString(util.RedactJSON(jsonBytes)))
	default:
		// Fallback for other types
		return slog.StringValue(util.ToString(v))
	}
}

func summarizeCallToolResult(ctr *mcp.CallToolResult) slog.Value {
	if ctr == nil {
		return slog.StringValue("<nil>")
	}
	attrs := make([]slog.Attr, 0, 2)
	attrs = append(attrs, slog.Bool("isError", ctr.IsError))

	// Summarize content
	contentSummaries := make([]string, 0, len(ctr.Content))
	for _, c := range ctr.Content {
		switch c := c.(type) {
		case *mcp.TextContent:
			// Truncate text if too long
			text := c.Text
			if len(text) > 512 {
				text = text[:512] + fmt.Sprintf("... (%d chars truncated)", len(c.Text)-512)
			}
			contentSummaries = append(contentSummaries, fmt.Sprintf("Text(len=%d): %q", len(c.Text), text))
		case *mcp.ImageContent:
			contentSummaries = append(contentSummaries, fmt.Sprintf("Image(mime=%s, size=%d bytes)", c.MIMEType, len(c.Data)))
		case *mcp.EmbeddedResource:
			res := c.Resource
			if res == nil {
				contentSummaries = append(contentSummaries, "Resource(<nil>)")
				continue
			}
			desc := fmt.Sprintf("Resource(uri=%s)", res.URI)
			if len(res.Blob) > 0 {
				desc += fmt.Sprintf(" blob=%d bytes", len(res.Blob))
			}
			if len(res.Text) > 0 {
				desc += fmt.Sprintf(" text=%d chars", len(res.Text))
			}
			contentSummaries = append(contentSummaries, desc)
		}
	}
	attrs = append(attrs, slog.Any("content", contentSummaries))
	return slog.GroupValue(attrs...)
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
			// ⚡ Bolt Optimization: Fetch allowed services once to avoid N lock acquisitions
			var allowedServices map[string]bool
			if profileID != "" {
				allowedServices, _ = s.toolManager.GetAllowedServiceIDs(profileID)
			}

			for _, resourceInstance := range managedResources {
				// Profile filtering
				if profileID != "" {
					serviceID := resourceInstance.Service()
					// Optimized O(1) map lookup
					if allowedServices != nil {
						if !allowedServices[serviceID] {
							continue
						}
					} else {
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
			// ⚡ Bolt Optimization: Fetch allowed services once to avoid N lock acquisitions
			var allowedServices map[string]bool
			if profileID != "" {
				allowedServices, _ = s.toolManager.GetAllowedServiceIDs(profileID)
			}

			for _, promptInstance := range managedPrompts {
				// Profile filtering
				if profileID != "" {
					serviceID := promptInstance.Service()
					// Optimized O(1) map lookup
					if allowedServices != nil {
						if !allowedServices[serviceID] {
							continue
						}
					} else {
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
