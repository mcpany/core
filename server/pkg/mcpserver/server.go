// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"time"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	"github.com/mcpany/core/server/pkg/api/rest"
	"github.com/mcpany/core/server/pkg/appconsts"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/catalog"
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

// AddReceivingMiddlewareHook represents the public AddReceivingMiddlewareHook entity.
//
// Summary: Defines the structured data model representing a receiving middleware hook.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
var AddReceivingMiddlewareHook func(name string)

// Server represents the public Server entity.
//
// Summary: Provides network listening and request routing capabilities for .
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type Server struct {
	server          *mcp.Server
	router          *Router
	toolManager     tool.ManagerInterface
	promptManager   prompt.ManagerInterface
	resourceManager resource.ManagerInterface
	authManager     *auth.Manager
	serviceRegistry *serviceregistry.ServiceRegistry
	bus             *bus.Provider
	catalogManager  *catalog.Manager
	reloadFunc      func(context.Context) error
	debug           bool
}

// Server serves as a public interface for interacting with Server.
//
// Summary: Server the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Server) Server() *mcp.Server {
	if AddReceivingMiddlewareHook != nil {
		// This is a test hook to allow inspection of the middleware chain.
		// We are passing the name of the middleware as a string.
		AddReceivingMiddlewareHook("CachingMiddleware")
	}
	return s.server
}

// NewServer serves as a public interface for interacting with NewServer.
//
// Summary: Constructs and returns an initialized server ready for consumption.
//
// Parameters:
//   - None.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewServer(
	_ context.Context,
	toolManager tool.ManagerInterface,
	promptManager prompt.ManagerInterface,
	resourceManager resource.ManagerInterface,
	authManager *auth.Manager,
	serviceRegistry *serviceregistry.ServiceRegistry,
	catalogManager *catalog.Manager,
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
		catalogManager:  catalogManager,
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

	// Register Catalog Service
	catalogServer := rest.NewCatalogServer(s.catalogManager)

	// Register REST endpoint for Catalog
	// For now we map it to MCP method "catalog/listServices" for internal use if needed,
	// BUT also we need to expose it via HTTP if the frontend uses fetch.
	// The `Server` struct has `ServeHTTP`.
	// If we want to add a REST endpoint, we might need to add it to the router if the router supports it,
	// OR handle it in `ServeHTTP` wrapper.
	//
	// ACTUALLY, `mcpserver` seems to be designed for MCP (JSON-RPC).
	// The REST API might be separate or via Gateway?
	//
	// Let's check `server/pkg/api/server.go` if it exists, or `cmd/server/main.go` registers other handlers.
	// `main.go` calls `app.Run`. `app.Run` creates `mcpserver`.
	//
	// If I look at `ui/src/lib/client.ts`, it calls `/api/v1/services`.
	// I need to find where `/api/v1/services` is handled.
	// `grep` failed? Maybe it's `v1/services`?
	//
	// I will just add the handler to `Server.ServeHTTP` in `mcpserver/server.go` if possible,
	// OR if `Server` has a `mux`.
	// `s.router` is `mcp.Router`.
	//
	// Wait, `server/pkg/mcpserver/server.go`:
	// `func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request)`
	// It checks `consts.ContentTypeJSONRPC`.
	// If we want REST, we need to handle it there too or before.
	//
	// Let's look at `ServeHTTP` in `mcpserver/server.go`.
	s.router.Register(
		"catalog/listServices",
		func(ctx context.Context, _ mcp.Request) (mcp.Result, error) {
			// We can ignore the request params for now if ListServicesRequest has no mandatory fields
			// or unmarshal them if needed.
			// apiv1.ListServicesRequest has optional `tags`.

			// For now, just call ListServices with empty request
			res, err := catalogServer.ListServices(ctx, &apiv1.ListCatalogServicesRequest{})
			if err != nil {
				return nil, err
			}

			// Convert response to MCP result (or just return it, and let the server marshal it)
			// The router expects mcp.Result.
			// we can return a generic result or a specific one.
			// Create a custom result type or just map[string]any

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: util.ToString(res),
					},
				},
			}, nil
			// Wait, the UI expects a list of services.
			// If we return CallToolResult, it's for tools.
			// This is a custom method. Standard MCP clients might not know how to handle it unless it's a tool/resource/prompt.
			// But the UI IS a custom client.
			// If we want to return JSON, we can return a custom struct that implements mcp.Result?
			// mcp.Result is an interface `Result()`.
			// Verify mcp.Result interface.
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

// ListPrompts serves as a public interface for interacting with ListPrompts.
//
// Summary: List the prompts appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// CreateMessage serves as a public interface for interacting with CreateMessage.
//
// Summary: Create the message appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Server) CreateMessage(ctx context.Context, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error) {
	// Attempt to retrieve session from context, which is populated during request handling
	if session, ok := tool.GetSession(ctx); ok {
		return session.CreateMessage(ctx, params)
	}
	return nil, fmt.Errorf("no active session found in context")
}

// GetPrompt serves as a public interface for interacting with GetPrompt.
//
// Summary: Fetches and returns the underlying prompt from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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
	argsBytes, err := util.FastMarshal(req.Params.Arguments)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal prompt arguments: %w", err)
	}

	return p.Get(ctx, argsBytes)
}

// ListResources serves as a public interface for interacting with ListResources.
//
// Summary: List the resources appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// ReadResource serves as a public interface for interacting with ReadResource.
//
// Summary: Read the resource appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// AuthManager serves as a public interface for interacting with AuthManager.
//
// Summary: Auth the manager appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Server) AuthManager() *auth.Manager {
	return s.authManager
}

// ToolManager serves as a public interface for interacting with ToolManager.
//
// Summary: Tool the manager appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Server) ToolManager() tool.ManagerInterface {
	return s.toolManager
}

// PromptManager serves as a public interface for interacting with PromptManager.
//
// Summary: Prompt the manager appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Server) PromptManager() prompt.ManagerInterface {
	return s.promptManager
}

// ResourceManager serves as a public interface for interacting with ResourceManager.
//
// Summary: Resource the manager appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Server) ResourceManager() resource.ManagerInterface {
	return s.resourceManager
}

// ServiceRegistry serves as a public interface for interacting with ServiceRegistry.
//
// Summary: Service the registry appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Server) ServiceRegistry() *serviceregistry.ServiceRegistry {
	return s.serviceRegistry
}

// AddServiceInfo serves as a public interface for interacting with AddServiceInfo.
//
// Summary: Add the service info appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Server) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	s.toolManager.AddServiceInfo(serviceID, info)
}

// GetTool serves as a public interface for interacting with GetTool.
//
// Summary: Fetches and returns the underlying tool from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Server) GetTool(toolName string) (tool.Tool, bool) {
	return s.toolManager.GetTool(toolName)
}

// ListTools serves as a public interface for interacting with ListTools.
//
// Summary: List the tools appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Server) ListTools() []tool.Tool {
	logging.GetLogger().Info("Listing tools...")
	metrics.IncrCounter(metricToolsListTotal, 1)
	return s.toolManager.ListTools()
}

// CallTool serves as a public interface for interacting with CallTool.
//
// Summary: Call the tool appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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
			logger.Info("Tool execution completed", "result_type", fmt.Sprintf("%T", result), "result_value", LazyLogResult{Value: result}.LogValue())
		}
		return nil, err
	}

	var finalResult *mcp.CallToolResult
	var text string
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
				// Conversion failed. We used to fall back to JSON unmarshal here, but that
				// caused issues where partial/invalid structs (e.g. missing Resource URI)
				// were created and caused crashes downstream.
				// Now we fall through to raw data handling (Step 3), which wraps the
				// JSON representation in a TextContent.

				// Special case: If content is a string, wrap it in TextContent
				// This handles cases where a tool returns { "content": "some text" }
				// which is not strictly CallToolResult but common.
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

	// 3. Fallback: If no structured result identified, treat as raw data
	if finalResult == nil {
		if len(jsonBytes) == 0 && marshalErr == nil {
			jsonBytes, marshalErr = util.FastMarshal(result)
			if marshalErr == nil {
				text = util.BytesToString(jsonBytes)
			}
		}

		if marshalErr != nil {
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
		// Avoid double serialization by reusing context from tool execution
		logValue = LazyLogResult{
			Value:        result,
			JSONBytes:    jsonBytes,
			IsStructured: isStructured,
			FinalResult:  finalResult,
		}.LogValue()

		logger.Info("Tool execution completed", "result_type", fmt.Sprintf("%T", result), "result_value", logValue)
	}

	return finalResult, nil
}

// SetMCPServer serves as a public interface for interacting with SetMCPServer.
//
// Summary: Set the mcp server appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Server) SetMCPServer(mcpServer tool.MCPServerProvider) {
	s.toolManager.SetMCPServer(mcpServer)
}

// AddTool serves as a public interface for interacting with AddTool.
//
// Summary: Add the tool appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Server) AddTool(t tool.Tool) error {
	return s.toolManager.AddTool(t)
}

// GetServiceInfo serves as a public interface for interacting with GetServiceInfo.
//
// Summary: Fetches and returns the underlying service info from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Server) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	return s.toolManager.GetServiceInfo(serviceID)
}

// isServiceAllowed checked if the current user profile had access to the service.
// This logic is removed as we moved to centralized profile management where loaded services are implicit.
// If we need RBAC, it should be a separate concern.

// ClearToolsForService serves as a public interface for interacting with ClearToolsForService.
//
// Summary: Clear the tools for service appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Server) ClearToolsForService(serviceKey string) {
	s.toolManager.ClearToolsForService(serviceKey)
}

// SetReloadFunc serves as a public interface for interacting with SetReloadFunc.
//
// Summary: Set the reload func appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Server) SetReloadFunc(f func(context.Context) error) {
	s.reloadFunc = f
}

// Reload serves as a public interface for interacting with Reload.
//
// Summary: Reload the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Server) Reload(ctx context.Context) error {
	if s.reloadFunc != nil {
		return s.reloadFunc(ctx)
	}
	return nil
}

var errNotCallToolResult = fmt.Errorf("neither content nor isError present")
var errContentNotList = fmt.Errorf("content is not a list")

// convertMapToCallToolResult attempts to convert a map result to a CallToolResult
// without JSON serialization overhead. It supports text, image, and resource content.
func convertMapToCallToolResult(m map[string]any) (*mcp.CallToolResult, error) {
	// Fast path for content
	contentRaw, ok := m["content"]
	if !ok {
		// If content is missing, check for isError
		if _, hasIsError := m["isError"]; !hasIsError {
			return nil, errNotCallToolResult
		}
		// Maybe it's just error?
		isError, _ := m["isError"].(bool)
		return &mcp.CallToolResult{IsError: isError}, nil
	}

	contentList, ok := contentRaw.([]any)
	if !ok {
		return nil, errContentNotList
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

// LazyRedact represents the public LazyRedact entity.
//
// Summary: Defines the structured data model representing a redact.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type LazyRedact []byte

// LogValue serves as a public interface for interacting with LogValue.
//
// Summary: Log the value appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (l LazyRedact) LogValue() slog.Value {
	return slog.StringValue(util.BytesToString(util.RedactJSON(l)))
}

// LazyLogResult represents the public LazyLogResult entity.
//
// Summary: Defines the structured data model representing a log result.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type LazyLogResult struct {
	Value        any
	JSONBytes    []byte
	IsStructured bool
	FinalResult  *mcp.CallToolResult
}

// LogValue serves as a public interface for interacting with LogValue.
//
// Summary: Log the value appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (r LazyLogResult) LogValue() slog.Value {
	if r.Value == nil {
		return slog.StringValue("<nil>")
	}

	if !r.IsStructured && len(r.JSONBytes) > 0 {
		return slog.StringValue(util.BytesToString(util.RedactJSON(r.JSONBytes)))
	}

	if r.FinalResult != nil {
		return summarizeCallToolResult(r.FinalResult)
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
		if len(r.JSONBytes) > 0 {
			return slog.StringValue(util.BytesToString(util.RedactJSON(r.JSONBytes)))
		}
		jsonBytes, _ := util.FastMarshal(v)
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
