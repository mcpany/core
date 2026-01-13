// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	// Use json-iterator for faster JSON operations.
	json "github.com/json-iterator/go"

	"github.com/google/uuid"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	xsync "github.com/puzpuzpuz/xsync/v4"
)

// MCPServerProvider defines an interface for components that can provide an
// instance of an *mcp.Server. This is used to decouple the Manager from the
// concrete server implementation.
type MCPServerProvider interface {
	// Server returns the MCP server instance.
	Server() *mcp.Server
}

// ManagerInterface defines the interface for a tool manager.
type ManagerInterface interface {
	// AddTool registers a new tool.
	AddTool(tool Tool) error
	// GetTool retrieves a tool by name.
	GetTool(toolName string) (Tool, bool)
	// ListTools returns all registered tools.
	ListTools() []Tool
	// ClearToolsForService removes all tools for a given service.
	ClearToolsForService(serviceID string)
	// ExecuteTool executes a tool with the given request.
	ExecuteTool(ctx context.Context, req *ExecutionRequest) (any, error)
	// SetMCPServer sets the MCP server provider.
	SetMCPServer(mcpServer MCPServerProvider)
	// AddMiddleware adds a middleware to the tool execution chain.
	AddMiddleware(middleware ExecutionMiddleware)
	// AddServiceInfo adds metadata for a service.
	AddServiceInfo(serviceID string, info *ServiceInfo)
	// GetServiceInfo retrieves metadata for a service.
	GetServiceInfo(serviceID string) (*ServiceInfo, bool)
	// ListServices returns all registered services.
	ListServices() []*ServiceInfo
	// SetProfiles sets the enabled profiles and their definitions.
	SetProfiles(enabled []string, defs []*configv1.ProfileDefinition)
	// IsServiceAllowed checks if a service is allowed for a given profile.
	IsServiceAllowed(serviceID, profileID string) bool
	// GetAllowedServiceIDs returns a set of allowed service IDs for a given profile.
	GetAllowedServiceIDs(profileID string) map[string]bool
	// ToolMatchesProfile checks if a tool matches a given profile.
	ToolMatchesProfile(tool Tool, profileID string) bool
}

// ExecutionMiddleware defines the interface for tool execution middleware.
type ExecutionMiddleware interface {
	// Execute executes the middleware logic.
	Execute(ctx context.Context, req *ExecutionRequest, next ExecutionFunc) (any, error)
}

// Manager is a thread-safe manager for registering tooling.
type Manager struct {
	tools       *xsync.Map[string, Tool]
	serviceInfo *xsync.Map[string, *ServiceInfo]
	nameMap     *xsync.Map[string, string] // Maps client-facing tool name to internal tool ID
	mcpServer   MCPServerProvider
	bus         *bus.Provider
	mu          sync.RWMutex
	middlewares []ExecutionMiddleware
	cachedTools []Tool
	toolsMutex  sync.RWMutex

	enabledProfiles []string
	profileDefs     map[string]*configv1.ProfileDefinition
}

// NewManager creates and returns a new, empty Manager.
func NewManager(bus *bus.Provider) *Manager {
	return &Manager{
		bus:         bus,
		tools:       xsync.NewMap[string, Tool](),
		serviceInfo: xsync.NewMap[string, *ServiceInfo](),
		nameMap:     xsync.NewMap[string, string](),
		profileDefs: make(map[string]*configv1.ProfileDefinition),
	}
}

// SetProfiles sets the enabled profiles and their definitions for filtering.
func (tm *Manager) SetProfiles(enabled []string, defs []*configv1.ProfileDefinition) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.enabledProfiles = enabled
	tm.profileDefs = make(map[string]*configv1.ProfileDefinition)
	for _, d := range defs {
		tm.profileDefs[d.GetName()] = d
	}
}

// isToolAllowed checks if the tool is allowed based on the enabled profiles.
// isToolAllowed checks if the tool is allowed based on the enabled profiles.
func (tm *Manager) isToolAllowed(t *v1.Tool) bool {
	// If no profiles are enabled, allow everything (default behavior).
	if len(tm.enabledProfiles) == 0 {
		return true
	}

	for _, ep := range tm.enabledProfiles {
		if tm.toolMatchesProfile(t, ep) {
			return true
		}
	}

	return false
}

func (tm *Manager) toolMatchesProfile(t *v1.Tool, profileName string) bool {
	// 1. Check explicit profile assignment
	for _, toolProfile := range t.GetProfiles() {
		if toolProfile == profileName {
			return true
		}
	}

	// 2. Check tag-based and property-based selection via ProfileDefinition
	def, ok := tm.profileDefs[profileName]
	if !ok {
		return false
	}

	// 3. Check Service Config
	if sc, ok := def.GetServiceConfig()[t.GetServiceId()]; ok {
		if sc.GetEnabled() {
			// If selector has criteria, apply them.
			if def.GetSelector() != nil {
				hasCriteria := len(def.GetSelector().GetTags()) > 0 || len(def.GetSelector().GetToolProperties()) > 0
				if hasCriteria {
					return tm.matchesSelector(t, def.GetSelector())
				}
			}
			// If no criteria, enabled service implies all tools allowed.
			return true
		}
	}

	return tm.matchesSelector(t, def.GetSelector())
}

// IsServiceAllowed checks if a service is allowed for a given profile.
func (tm *Manager) IsServiceAllowed(serviceID, profileID string) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	def, ok := tm.profileDefs[profileID]
	if !ok {
		return false
	}

	if sc, ok := def.GetServiceConfig()[serviceID]; ok {
		return sc.GetEnabled()
	}

	// Default to false if not explicitly enabled in service_config?
	// Or check selector? Selectors are for tools, not services?
	// But if a profile selects tools via tags, the service is implicitly allowed?
	// Ideally, service should be explicitly enabled.
	return false
}

// GetAllowedServiceIDs returns a set of allowed service IDs for a given profile.
func (tm *Manager) GetAllowedServiceIDs(profileID string) map[string]bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	def, ok := tm.profileDefs[profileID]
	if !ok {
		return nil
	}

	allowed := make(map[string]bool)
	for sID, sc := range def.GetServiceConfig() {
		if sc.GetEnabled() {
			allowed[sID] = true
		}
	}
	return allowed
}

// ToolMatchesProfile checks if a tool matches a given profile.
func (tm *Manager) ToolMatchesProfile(tool Tool, profileID string) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.toolMatchesProfile(tool.Tool(), profileID)
}

func (tm *Manager) matchesSelector(t *v1.Tool, selector *configv1.ProfileSelector) bool {
	if selector == nil {
		return false
	}

	hasTags := len(selector.GetTags()) > 0
	if hasTags && !tm.matchesTags(t.GetTags(), selector.GetTags()) {
		return false
	}

	hasProps := len(selector.GetToolProperties()) > 0
	if hasProps && !tm.matchesProperties(t.GetAnnotations(), selector.GetToolProperties()) {
		return false
	}

	return hasTags || hasProps
}

func (tm *Manager) matchesTags(toolTags []string, selectorTags []string) bool {
	for _, sTag := range selectorTags {
		for _, tTag := range toolTags {
			if sTag == tTag {
				return true
			}
		}
	}
	return false
}

func (tm *Manager) matchesProperties(annotations *v1.ToolAnnotations, props map[string]string) bool {
	const falseVal = "false"
	for k, v := range props {
		var actual string
		switch k {
		case "read_only":
			// Handle nil annotations gracefully if needed, though proto accessors usually return defaults
			if annotations == nil {
				actual = falseVal
			} else {
				actual = fmt.Sprintf("%v", annotations.GetReadOnlyHint())
			}
		case "destructive":
			if annotations == nil {
				actual = falseVal
			} else {
				actual = fmt.Sprintf("%v", annotations.GetDestructiveHint())
			}
		case "idempotent":
			if annotations == nil {
				actual = falseVal
			} else {
				actual = fmt.Sprintf("%v", annotations.GetIdempotentHint())
			}
		case "open_world":
			if annotations == nil {
				actual = falseVal
			} else {
				actual = fmt.Sprintf("%v", annotations.GetOpenWorldHint())
			}
		default:
			return false
		}

		if actual != v {
			return false
		}
	}
	return true
}

// AddMiddleware adds a middleware to the tool manager.
func (tm *Manager) AddMiddleware(middleware ExecutionMiddleware) {
	tm.middlewares = append(tm.middlewares, middleware)
}

// SetMCPServer provides the Manager with a reference to the MCP server.
// This is necessary for registering tool handlers with the server.
func (tm *Manager) SetMCPServer(mcpServer MCPServerProvider) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.mcpServer = mcpServer
}

// ExecuteTool finds a tool by its name and executes it with the provided
// request context and inputs.
//
// ctx is the context for the tool execution.
// req contains the name of the tool and its inputs.
// It returns the result of the execution or an error if the tool is not found
// or if the execution fails.
func (tm *Manager) ExecuteTool(ctx context.Context, req *ExecutionRequest) (any, error) {
	log := logging.GetLogger().With("toolName", req.ToolName)
	log.Debug("Executing tool")

	// 1. Resolve Tool and Service Info
	t, ok := tm.GetTool(req.ToolName)
	if !ok {
		log.Error("Tool not found")
		return nil, ErrToolNotFound
	}
	serviceID := t.Tool().GetServiceId()
	serviceInfo, ok := tm.GetServiceInfo(serviceID)

	var preHooks []PreCallHook
	var postHooks []PostCallHook
	if ok {
		preHooks = serviceInfo.PreHooks
		postHooks = serviceInfo.PostHooks
	}

	// 2. Initialize Context with Tool and CacheControl
	ctx = NewContextWithTool(ctx, t)
	ctx = NewContextWithCacheControl(ctx, &CacheControl{Action: ActionAllow})

	// 3. Run Pre-execution Hooks (modifies ctx/req)
	for _, h := range preHooks {
		action, modifiedReq, err := h.ExecutePre(ctx, req)
		if err != nil {
			log.Warn("Tool execution denied by pre-hook error", "error", err)
			return nil, err
		}
		if action == ActionDeny {
			log.Warn("Tool execution denied by pre-hook")
			return nil, fmt.Errorf("tool execution denied by hook")
		}
		if action == ActionSaveCache || action == ActionDeleteCache {
			if cc, ok := GetCacheControl(ctx); ok {
				cc.Action = action
			}
		}
		if modifiedReq != nil {
			req = modifiedReq
		}
	}

	// 4. Define Core Execution (Execute + PostHooks)
	executeCore := func(ctx context.Context, req *ExecutionRequest) (any, error) {
		result, err := t.Execute(ctx, req)

		// Execute Post Hooks
		for _, h := range postHooks {
			newResult, hkErr := h.ExecutePost(ctx, req, result)
			if hkErr != nil {
				log.Warn("Post-hook execution failed", "error", hkErr)
				return nil, hkErr
			}
			result = newResult
		}

		if err != nil {
			log.Error("Tool execution failed", "error", err)
		} else {
			log.Debug("Tool execution successful")
		}
		return result, err
	}

	// 5. Build and Run Middleware Chain
	chain := executeCore
	for i := len(tm.middlewares) - 1; i >= 0; i-- {
		m := tm.middlewares[i]
		chain = func(next ExecutionFunc) ExecutionFunc {
			return func(ctx context.Context, req *ExecutionRequest) (any, error) {
				log.Debug("Executing middleware", "middleware", i)
				return m.Execute(ctx, req, next)
			}
		}(chain)
	}

	result, err := chain(ctx, req)
	if err != nil {
		log.Error("Tool execution chain failed", "error", err)
	} else {
		log.Debug("Tool execution chain successful")
	}
	return result, err
}

// AddServiceInfo stores metadata about a service, indexed by its ID.
//
// serviceID is the unique identifier for the service.
// info is the ServiceInfo struct containing the service's metadata.
func (tm *Manager) AddServiceInfo(serviceID string, info *ServiceInfo) {
	if info.Config != nil {
		var preHooks []PreCallHook
		var postHooks []PostCallHook

		// 1. New Call Policies -> converted to PreHook
		for _, policy := range info.Config.GetCallPolicies() {
			preHooks = append(preHooks, NewPolicyHook(policy))
		}

		// Compile call policies for the middleware
		if compiled, err := CompileCallPolicies(info.Config.GetCallPolicies()); err == nil {
			info.CompiledPolicies = compiled
		} else {
			logging.GetLogger().Error("Failed to compile call policies", "error", err)
		}
		// 2. PreCallHooks
		for _, hCfg := range info.Config.GetPreCallHooks() {
			if p := hCfg.GetCallPolicy(); p != nil {
				preHooks = append(preHooks, NewPolicyHook(p))
			}
			if w := hCfg.GetWebhook(); w != nil {
				preHooks = append(preHooks, NewWebhookHook(w))
			}
		}
		// 3. PostCallHooks
		for _, hCfg := range info.Config.GetPostCallHooks() {
			if w := hCfg.GetWebhook(); w != nil {
				postHooks = append(postHooks, NewWebhookHook(w))
			}
		}
		info.PreHooks = preHooks
		info.PostHooks = postHooks
	}
	tm.serviceInfo.Store(serviceID, info)
}

// GetServiceInfo retrieves the metadata for a service by its ID.
//
// serviceID is the unique identifier for the service.
// It returns the ServiceInfo and a boolean indicating whether the service was found.
func (tm *Manager) GetServiceInfo(serviceID string) (*ServiceInfo, bool) {
	info, ok := tm.serviceInfo.Load(serviceID)
	if !ok {
		return nil, false
	}
	return info, true
}

// ListServices returns a slice containing all the services currently registered with
// the manager.
func (tm *Manager) ListServices() []*ServiceInfo {
	var services []*ServiceInfo
	tm.serviceInfo.Range(func(_ string, value *ServiceInfo) bool {
		services = append(services, value)
		return true
	})
	return services
}

// AddTool registers a new tool with the manager. It generates a unique tool ID
// and, if an MCP server is configured, registers a handler for the tool with
// the server.
//
// tool is the tool to be added.
// It returns an error if the tool ID cannot be generated.
func (tm *Manager) AddTool(tool Tool) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Filter tool based on profiles
	if !tm.isToolAllowed(tool.Tool()) {
		return nil
	}

	if tool.Tool().GetServiceId() == "" {
		return fmt.Errorf("tool service ID cannot be empty")
	}

	sanitizedToolName, err := util.SanitizeToolName(tool.Tool().GetName())
	if err != nil {
		logging.GetLogger().
			Error("Failed to sanitize tool name", "serviceID", tool.Tool().GetServiceId(), "toolName", tool.Tool().GetName(), "error", err)
		return fmt.Errorf("failed to sanitize tool name: %w", err)
	}
	toolID := tool.Tool().GetServiceId() + "." + sanitizedToolName
	log := logging.GetLogger().With("toolID", toolID)
	log.Debug("Adding tool to Manager")
	tm.tools.Store(toolID, tool)

	// Map client-facing name to internal ID
	// Use tool.Tool().GetName() which is the raw name (e.g. "mcp:list_roots")
	// If the name is scoped (e.g. "service.tool"), this works.
	// If there are collisions, last one wins.
	tm.nameMap.Store(tool.Tool().GetName(), toolID)
	// Also map the fully qualified name that the MCP server exposes to clients
	if tool.Tool().GetServiceId() != "" {
		fullExposedName := tool.Tool().GetServiceId() + "." + tool.Tool().GetName()
		tm.nameMap.Store(fullExposedName, toolID)
	}

	tm.toolsMutex.Lock()
	tm.cachedTools = nil
	tm.toolsMutex.Unlock()

	if tm.mcpServer != nil {
		mcpTool, err := ConvertProtoToMCPTool(tool.Tool())
		if err != nil {
			return fmt.Errorf("failed to convert proto tool to mcp tool: %w", err)
		}

		if tool.Tool().GetInputSchema() != nil {
			b, err := tool.Tool().GetInputSchema().MarshalJSON()
			if err != nil {
				return fmt.Errorf("failed to marshal input schema: %w", err)
			}
			if err := json.Unmarshal(b, &mcpTool.InputSchema); err != nil {
				return fmt.Errorf("failed to unmarshal input schema: %w", err)
			}
		}

		if tool.Tool().GetOutputSchema() != nil {
			b, err := tool.Tool().GetOutputSchema().MarshalJSON()
			if err != nil {
				return fmt.Errorf("failed to marshal output schema: %w", err)
			}
			if err := json.Unmarshal(b, &mcpTool.OutputSchema); err != nil {
				return fmt.Errorf("failed to unmarshal output schema: %w", err)
			}
		}

		log.Info(
			"Registering tool with MCP server",
			"toolName",
			mcpTool.Name,
			"tool",
			mcpTool,
		)

		handler := func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			logging.GetLogger().Info("Queueing tool execution", "toolName", req.Params.Name)

			correlationID := uuid.New().String()
			resultChan := make(chan *bus.ToolExecutionResult, 1)

			resultBus, err := bus.GetBus[*bus.ToolExecutionResult](tm.bus, "tool_execution_results")
			if err != nil {
				return nil, fmt.Errorf("failed to get result bus: %w", err)
			}
			unsubscribe := resultBus.SubscribeOnce(
				ctx,
				correlationID,
				func(result *bus.ToolExecutionResult) {
					resultChan <- result
				},
			)
			defer unsubscribe()

			requestBus, err := bus.GetBus[*bus.ToolExecutionRequest](tm.bus, "tool_execution_requests")
			if err != nil {
				return nil, fmt.Errorf("failed to get request bus: %w", err)
			}
			execReq := &bus.ToolExecutionRequest{
				Context:    ctx,
				ToolName:   req.Params.Name,
				ToolInputs: req.Params.Arguments,
			}
			execReq.SetCorrelationID(correlationID)
			if err := requestBus.Publish(ctx, "request", execReq); err != nil {
				return nil, fmt.Errorf("failed to publish request: %w", err)
			}

			select {
			case result := <-resultChan:
				if result.Error != nil {
					return nil, fmt.Errorf(
						"error executing tool %s: %w",
						req.Params.Name,
						result.Error,
					)
				}

				jsonResult, err := json.Marshal(result.Result)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal tool result: %w", err)
				}

				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: string(jsonResult),
						},
					},
				}, nil
			case <-ctx.Done():
				return nil, fmt.Errorf("context deadline exceeded while waiting for tool execution")
			case <-time.After(60 * time.Second): // Safety timeout
				return nil, fmt.Errorf(
					"timed out waiting for tool execution result for tool %s",
					req.Params.Name,
				)
			}
		}
		tm.mcpServer.Server().AddTool(mcpTool, handler)
	}
	return nil
}

// GetTool retrieves a tool from the manager by its fully qualified name.
//
// toolName is the name of the tool to retrieve.
// It returns the tool and a boolean indicating whether the tool was found.
func (tm *Manager) GetTool(toolName string) (Tool, bool) {
	// Try direct lookup (if client sends ID)
	tool, ok := tm.tools.Load(toolName)
	if ok {
		return tool, true
	}

	// Try lookup by name
	if id, ok := tm.nameMap.Load(toolName); ok {
		if tool, ok := tm.tools.Load(id); ok {
			return tool, true
		}
	}

	return nil, false
}

// ListTools returns a slice containing all the tools currently registered with
// the manager.
func (tm *Manager) ListTools() []Tool {
	tm.toolsMutex.RLock()
	if tm.cachedTools != nil {
		defer tm.toolsMutex.RUnlock()
		return tm.cachedTools
	}
	tm.toolsMutex.RUnlock()

	tm.toolsMutex.Lock()
	defer tm.toolsMutex.Unlock()
	// After acquiring the write lock, we need to check if the cache is still nil.
	// This is because another goroutine might have populated it between the RUnlock
	// and Lock calls.
	if tm.cachedTools != nil {
		return tm.cachedTools
	}

	var tools []Tool
	tm.tools.Range(func(_ string, value Tool) bool {
		tools = append(tools, value)
		return true
	})
	tm.cachedTools = tools
	return tools
}

// ClearToolsForService removes all tools associated with a given service key from
// the manager. This is useful when a service is being re-registered or
// unregistered.
//
// serviceID is the unique identifier for the service whose tools should be
// cleared.
func (tm *Manager) ClearToolsForService(serviceID string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	log := logging.GetLogger().With("serviceID", serviceID)
	log.Debug("Clearing existing tools for serviceID before reload/overwrite.")
	deletedCount := 0
	tm.tools.Range(func(key string, value Tool) bool {
		if value.Tool().GetServiceId() == serviceID {
			tm.tools.Delete(key)
			// Also remove from nameMap.
			// This is inefficient as we have to scan nameMap or store reverse mapping.
			// Since ClearToolsForService is rare (only on reload), scanning nameMap is acceptable?
			// OR we assume nameMap is consistent.
			// Let's scan nameMap to clean up by Value? No Xsync Map doesn't support Range easily for Delete?
			// It does supports Range.
			// But deleting while ranging?
			deletedCount++
		}
		return true
	})

	// Cleanup NameMap
	tm.nameMap.Range(func(key, value string) bool {
		// If the ID (value) belongs to the service?
		// We don't verify serviceID from key easily without parsing.
		// Value is toolID: "serviceID.sanitizedName".
		if strings.HasPrefix(value, serviceID+".") {
			tm.nameMap.Delete(key)
		}
		return true
	})

	if deletedCount > 0 {
		tm.toolsMutex.Lock()
		tm.cachedTools = nil
		tm.toolsMutex.Unlock()
	}
	log.Debug("Cleared tools for serviceID", "count", deletedCount)
}
