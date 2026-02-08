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
	"google.golang.org/protobuf/proto"
)

// MCPServerProvider defines an interface for components that can provide an.
//
// Summary: defines an interface for components that can provide an.
type MCPServerProvider interface {
	// Server returns the MCP server instance.
	//
	// Summary: returns the MCP server instance.
	//
	// Parameters:
	//   None.
	//
	// Returns:
	//   - *mcp.Server: The *mcp.Server.
	Server() *mcp.Server
}

// ManagerInterface defines the interface for a tool manager.
//
// Summary: defines the interface for a tool manager.
type ManagerInterface interface {
	// AddTool registers a new tool.
	//
	// Summary: registers a new tool.
	//
	// Parameters:
	//   - tool: Tool. The tool.
	//
	// Returns:
	//   - error: An error if the operation fails.
	//
	// Throws/Errors:
	//   Returns an error if the operation fails.
	AddTool(tool Tool) error
	// GetTool retrieves a tool by name.
	//
	// Summary: retrieves a tool by name.
	//
	// Parameters:
	//   - toolName: string. The string.
	//
	// Returns:
	//   - Tool: The Tool.
	//   - bool: The bool.
	GetTool(toolName string) (Tool, bool)
	// ListTools returns all registered tools.
	//
	// Summary: returns all registered tools.
	//
	// Parameters:
	//   None.
	//
	// Returns:
	//   - []Tool: The []Tool.
	ListTools() []Tool
	// ListMCPTools returns all registered tools in MCP format.
	//
	// Summary: returns all registered tools in MCP format.
	//
	// Parameters:
	//   None.
	//
	// Returns:
	//   - []*mcp.Tool: The []*mcp.Tool.
	ListMCPTools() []*mcp.Tool
	// ClearToolsForService removes all tools for a given service.
	//
	// Summary: removes all tools for a given service.
	//
	// Parameters:
	//   - serviceID: string. The string.
	//
	// Returns:
	//   None.
	ClearToolsForService(serviceID string)
	// ExecuteTool executes a tool with the given request.
	//
	// Summary: executes a tool with the given request.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the operation.
	//   - req: *ExecutionRequest. The request object.
	//
	// Returns:
	//   - any: The any.
	//   - error: An error if the operation fails.
	//
	// Throws/Errors:
	//   Returns an error if the operation fails.
	ExecuteTool(ctx context.Context, req *ExecutionRequest) (any, error)
	// SetMCPServer sets the MCP server provider.
	//
	// Summary: sets the MCP server provider.
	//
	// Parameters:
	//   - mcpServer: MCPServerProvider. The mcp server provider.
	//
	// Returns:
	//   None.
	SetMCPServer(mcpServer MCPServerProvider)
	// AddMiddleware adds a middleware to the tool execution chain.
	//
	// Summary: adds a middleware to the tool execution chain.
	//
	// Parameters:
	//   - middleware: ExecutionMiddleware. The execution middleware.
	//
	// Returns:
	//   None.
	AddMiddleware(middleware ExecutionMiddleware)
	// AddServiceInfo adds metadata for a service.
	//
	// Summary: adds metadata for a service.
	//
	// Parameters:
	//   - serviceID: string. The string.
	//   - info: *ServiceInfo. The service info.
	//
	// Returns:
	//   None.
	AddServiceInfo(serviceID string, info *ServiceInfo)
	// GetServiceInfo retrieves metadata for a service.
	//
	// Summary: retrieves metadata for a service.
	//
	// Parameters:
	//   - serviceID: string. The string.
	//
	// Returns:
	//   - *ServiceInfo: The *ServiceInfo.
	//   - bool: The bool.
	GetServiceInfo(serviceID string) (*ServiceInfo, bool)
	// ListServices returns all registered services.
	//
	// Summary: returns all registered services.
	//
	// Parameters:
	//   None.
	//
	// Returns:
	//   - []*ServiceInfo: The []*ServiceInfo.
	ListServices() []*ServiceInfo
	// SetProfiles sets the enabled profiles and their definitions.
	//
	// Summary: sets the enabled profiles and their definitions.
	//
	// Parameters:
	//   - enabled: []string. The []string.
	//   - defs: []*configv1.ProfileDefinition. The profile definition.
	//
	// Returns:
	//   None.
	SetProfiles(enabled []string, defs []*configv1.ProfileDefinition)
	// IsServiceAllowed checks if a service is allowed for a given profile.
	//
	// Summary: checks if a service is allowed for a given profile.
	//
	// Parameters:
	//   - serviceID: string. The string.
	//   - profileID: string. The string.
	//
	// Returns:
	//   - bool: The bool.
	IsServiceAllowed(serviceID, profileID string) bool
	// ToolMatchesProfile checks if a tool matches a given profile.
	//
	// Summary: checks if a tool matches a given profile.
	//
	// Parameters:
	//   - tool: Tool. The tool.
	//   - profileID: string. The string.
	//
	// Returns:
	//   - bool: The bool.
	ToolMatchesProfile(tool Tool, profileID string) bool
	// GetAllowedServiceIDs returns a map of allowed service IDs for a given profile.
	//
	// Summary: returns a map of allowed service IDs for a given profile.
	//
	// Parameters:
	//   - profileID: string. The string.
	//
	// Returns:
	//   - map[string]bool: The map[string]bool.
	//   - bool: The bool.
	GetAllowedServiceIDs(profileID string) (map[string]bool, bool)
	// GetToolCountForService returns the number of tools for a given service.
	//
	// Summary: returns the number of tools for a given service.
	//
	// Parameters:
	//   - serviceID: string. The string.
	//
	// Returns:
	//   - int: The int.
	GetToolCountForService(serviceID string) int
}

// ExecutionMiddleware defines the interface for tool execution middleware.
//
// Summary: defines the interface for tool execution middleware.
type ExecutionMiddleware interface {
	// Execute executes the middleware logic.
	//
	// Summary: executes the middleware logic.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the operation.
	//   - req: *ExecutionRequest. The request object.
	//   - next: ExecutionFunc. The execution func.
	//
	// Returns:
	//   - any: The any.
	//   - error: An error if the operation fails.
	//
	// Throws/Errors:
	//   Returns an error if the operation fails.
	Execute(ctx context.Context, req *ExecutionRequest, next ExecutionFunc) (any, error)
}

// Manager is a thread-safe manager for registering tooling.
//
// Summary: is a thread-safe manager for registering tooling.
type Manager struct {
	tools       *xsync.Map[string, Tool]
	serviceInfo *xsync.Map[string, *ServiceInfo]
	nameMap     *xsync.Map[string, string] // Maps client-facing tool name to internal tool ID
	mcpServer   MCPServerProvider
	bus         *bus.Provider
	mu          sync.RWMutex
	middlewares []ExecutionMiddleware
	cachedTools []Tool
	// cachedMCPTools caches the list of tools in MCP format to avoid
	// re-allocating and re-converting them on every request.
	cachedMCPTools []*mcp.Tool
	toolsMutex     sync.RWMutex

	// Indices for O(1) cleanup
	serviceToolIDs   map[string]map[string]struct{}
	serviceToolNames map[string]map[string]struct{}

	enabledProfiles      []string
	profileDefs          map[string]*configv1.ProfileDefinition
	allowedServicesCache map[string]map[string]bool
}

// NewManager creates and returns a new, empty Manager.
//
// Summary: creates and returns a new, empty Manager.
//
// Parameters:
//   - bus: *bus.Provider. The bus.
//
// Returns:
//   - *Manager: The *Manager.
func NewManager(bus *bus.Provider) *Manager {
	return &Manager{
		bus:                  bus,
		tools:                xsync.NewMap[string, Tool](),
		serviceInfo:          xsync.NewMap[string, *ServiceInfo](),
		nameMap:              xsync.NewMap[string, string](),
		serviceToolIDs:       make(map[string]map[string]struct{}),
		serviceToolNames:     make(map[string]map[string]struct{}),
		profileDefs:          make(map[string]*configv1.ProfileDefinition),
		allowedServicesCache: make(map[string]map[string]bool),
	}
}

// SetProfiles sets the enabled profiles and their definitions for filtering.
//
// Summary: sets the enabled profiles and their definitions for filtering.
//
// Parameters:
//   - enabled: []string. The enabled.
//   - defs: []*configv1.ProfileDefinition. The defs.
//
// Returns:
//   None.
func (tm *Manager) SetProfiles(enabled []string, defs []*configv1.ProfileDefinition) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.enabledProfiles = enabled
	tm.profileDefs = make(map[string]*configv1.ProfileDefinition)
	tm.allowedServicesCache = make(map[string]map[string]bool)

	for _, d := range defs {
		tm.profileDefs[d.GetName()] = d

		// Pre-compute allowed services for this profile
		allowed := make(map[string]bool)
		for serviceID, sc := range d.GetServiceConfig() {
			if sc.GetEnabled() {
				allowed[serviceID] = true
			}
		}
		tm.allowedServicesCache[d.GetName()] = allowed
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
			// Check if specific tool is disabled
			if tc, ok := sc.GetTools()[t.GetName()]; ok && tc.GetDisabled() {
				return false
			}

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
	} else {
		// Fallback: Check Service Config by Name (retrieved via ServiceInfo)
		// This handles cases where the Service ID might have been hashed (e.g. if empty in config)
		// but the profile uses the human-readable Service Name.
		if info, found := tm.GetServiceInfo(t.GetServiceId()); found && info.Config != nil {
			if sc, ok := def.GetServiceConfig()[info.Config.GetName()]; ok {
				if sc.GetEnabled() {
					// Check if specific tool is disabled
					if tc, ok := sc.GetTools()[t.GetName()]; ok && tc.GetDisabled() {
						return false
					}

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
		}
	}

	return tm.matchesSelector(t, def.GetSelector())
}

// IsServiceAllowed checks if a service is allowed for a given profile.
//
// Summary: checks if a service is allowed for a given profile.
//
// Parameters:
//   - serviceID: string. The serviceID.
//   - profileID: string. The profileID.
//
// Returns:
//   - bool: The bool.
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

// ToolMatchesProfile checks if a tool matches a given profile.
//
// Summary: checks if a tool matches a given profile.
//
// Parameters:
//   - tool: Tool. The tool.
//   - profileID: string. The profileID.
//
// Returns:
//   - bool: The bool.
func (tm *Manager) ToolMatchesProfile(tool Tool, profileID string) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.toolMatchesProfile(tool.Tool(), profileID)
}

// GetAllowedServiceIDs returns a map of allowed service IDs for a given profile.
//
// Summary: returns a map of allowed service IDs for a given profile.
//
// Parameters:
//   - profileID: string. The profileID.
//
// Returns:
//   - map[string]bool: The map[string]bool.
//   - bool: The bool.
func (tm *Manager) GetAllowedServiceIDs(profileID string) (map[string]bool, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	allowed, ok := tm.allowedServicesCache[profileID]
	return allowed, ok
}

// GetToolCountForService returns the number of tools for a given service.
//
// Summary: returns the number of tools for a given service.
//
// Parameters:
//   - serviceID: string. The serviceID.
//
// Returns:
//   - int: The int.
func (tm *Manager) GetToolCountForService(serviceID string) int {
	// ⚡ Bolt Optimization: Check health status first without locking the main mutex.
	// Randomized Selection from Top 5 High-Impact Targets
	if info, ok := tm.serviceInfo.Load(serviceID); ok {
		if info.HealthStatus == HealthStatusUnhealthy {
			return 0
		}
	}

	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if tools, ok := tm.serviceToolIDs[serviceID]; ok {
		return len(tools)
	}
	return 0
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
//
// Summary: adds a middleware to the tool manager.
//
// Parameters:
//   - middleware: ExecutionMiddleware. The middleware.
//
// Returns:
//   None.
func (tm *Manager) AddMiddleware(middleware ExecutionMiddleware) {
	tm.middlewares = append(tm.middlewares, middleware)
}

// SetMCPServer provides the Manager with a reference to the MCP server.
//
// Summary: provides the Manager with a reference to the MCP server.
//
// Parameters:
//   - mcpServer: MCPServerProvider. The mcpServer.
//
// Returns:
//   None.
func (tm *Manager) SetMCPServer(mcpServer MCPServerProvider) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.mcpServer = mcpServer
}

// ExecuteTool finds a tool by its name and executes it with the provided.
//
// Summary: finds a tool by its name and executes it with the provided.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - req: *ExecutionRequest. The req.
//
// Returns:
//   - any: The any.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (tm *Manager) ExecuteTool(ctx context.Context, req *ExecutionRequest) (any, error) {
	log := logging.GetLogger().With("toolName", req.ToolName)
	log.Debug("Executing tool")

	// 1. Resolve Tool and Service Info
	var t Tool
	var ok bool
	// ⚡ Bolt Optimization: Use resolved tool if provided to avoid map lookup
	if req.Tool != nil {
		t = req.Tool
		ok = true
	} else {
		t, ok = tm.GetTool(req.ToolName)
	}

	if !ok {
		log.Error("Tool not found")

		// Friction Fighter: Fuzzy Matching for better error messages
		var bestMatch string
		minDist := 1000
		var suffixMatches []string

		// ⚡ Bolt Optimization: Use bounded Levenshtein search
		// We only care about matches with distance <= 3.
		currentLimit := 3

		// Iterate over exposed tool names (keys in nameMap)
		tm.nameMap.Range(func(name string, _ string) bool {
			// Check for missing namespace (suffix match)
			// e.g. req="get_weather", existing="weather.get_weather"
			if strings.HasSuffix(name, "."+req.ToolName) {
				suffixMatches = append(suffixMatches, name)
			}

			// Levenshtein check with limit
			dist := util.LevenshteinDistanceWithLimit(req.ToolName, name, currentLimit)
			if dist <= currentLimit {
				if dist < minDist {
					minDist = dist
					bestMatch = name
					// Tighten the limit: we only care about finding something BETTER than what we have.
					// If we have distance 2, next time we only care about distance <= 1.
					currentLimit = dist - 1
				}
			}
			return true
		})

		if len(suffixMatches) > 0 {
			if len(suffixMatches) == 1 {
				return nil, fmt.Errorf("%w: %q (did you mean %q?)", ErrToolNotFound, req.ToolName, suffixMatches[0])
			}
			return nil, fmt.Errorf("%w: %q (did you mean one of: %s?)", ErrToolNotFound, req.ToolName, strings.Join(suffixMatches, ", "))
		}

		if minDist <= 3 && bestMatch != "" {
			return nil, fmt.Errorf("%w: %q (did you mean %q?)", ErrToolNotFound, req.ToolName, bestMatch)
		}

		return nil, ErrToolNotFound
	}
	serviceID := t.Tool().GetServiceId()
	// ⚡ Bolt Optimization: Use direct load to avoid expensive config cloning/stripping in GetServiceInfo
	serviceInfo, ok := tm.serviceInfo.Load(serviceID)

	var preHooks []PreCallHook
	var postHooks []PostCallHook
	if ok {
		if serviceInfo.HealthStatus == HealthStatusUnhealthy {
			log.Warn("Service is unhealthy, denying execution", "serviceID", serviceID)
			return nil, fmt.Errorf("service %s is currently unhealthy", serviceID)
		}
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

	start := time.Now()
	result, err := chain(ctx, req)
	duration := time.Since(start)

	if err != nil {
		log.Error("Tool execution failed", "error", err, "duration", duration.String())
	} else {
		log.Info("Tool execution successful", "duration", duration.String())
	}
	return result, err
}

// AddServiceInfo stores metadata about a service, indexed by its ID.
//
// Summary: stores metadata about a service, indexed by its ID.
//
// Parameters:
//   - serviceID: string. The serviceID.
//   - info: *ServiceInfo. The info.
//
// Returns:
//   None.
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
// Summary: retrieves the metadata for a service by its ID.
//
// Parameters:
//   - serviceID: string. The serviceID.
//
// Returns:
//   - *ServiceInfo: The *ServiceInfo.
//   - bool: The bool.
func (tm *Manager) GetServiceInfo(serviceID string) (*ServiceInfo, bool) {
	info, ok := tm.serviceInfo.Load(serviceID)
	if !ok {
		return nil, false
	}
	clonedInfo := *info
	if info.Config != nil {
		clonedConfig := proto.Clone(info.Config).(*configv1.UpstreamServiceConfig)
		util.StripSecretsFromService(clonedConfig)
		clonedInfo.Config = clonedConfig
	}
	return &clonedInfo, true
}

// ListServices returns a slice containing all the services currently registered with.
//
// Summary: returns a slice containing all the services currently registered with.
//
// Parameters:
//   None.
//
// Returns:
//   - []*ServiceInfo: The []*ServiceInfo.
func (tm *Manager) ListServices() []*ServiceInfo {
	var services []*ServiceInfo
	tm.serviceInfo.Range(func(_ string, value *ServiceInfo) bool {
		clonedInfo := *value
		if value.Config != nil {
			clonedConfig := proto.Clone(value.Config).(*configv1.UpstreamServiceConfig)
			util.StripSecretsFromService(clonedConfig)
			clonedInfo.Config = clonedConfig
		}
		services = append(services, &clonedInfo)
		return true
	})
	return services
}

// AddTool registers a new tool with the manager. It generates a unique tool ID.
//
// Summary: registers a new tool with the manager. It generates a unique tool ID.
//
// Parameters:
//   - tool: Tool. The tool.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (tm *Manager) AddTool(tool Tool) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Filter tool based on profiles
	if !tm.isToolAllowed(tool.Tool()) {
		return nil
	}

	// Feature 1: Tool Poisoning Mitigation
	// Verify tool integrity if hash is provided
	if err := VerifyIntegrity(tool.Tool()); err != nil {
		logging.GetLogger().Error("Tool integrity check failed", "toolName", tool.Tool().GetName(), "error", err)
		return fmt.Errorf("tool integrity check failed: %w", err)
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

	// Update indices
	serviceID := tool.Tool().GetServiceId()
	if tm.serviceToolIDs[serviceID] == nil {
		tm.serviceToolIDs[serviceID] = make(map[string]struct{})
	}
	tm.serviceToolIDs[serviceID][toolID] = struct{}{}

	// Map client-facing name to internal ID
	// Use tool.Tool().GetName() which is the raw name (e.g. "mcp:list_roots")
	// If the tool has a Service ID, we ONLY expose the namespaced version "service.tool"
	// to prevent auth bypasses and collisions.
	// If it doesn't have a Service ID (e.g. internal tools), we expose the short name.
	var nameKey string
	if tool.Tool().GetServiceId() == "" {
		nameKey = tool.Tool().GetName()
		tm.nameMap.Store(nameKey, toolID)
	} else {
		// Enforce namespacing for service tools
		nameKey = tool.Tool().GetServiceId() + "." + tool.Tool().GetName()
		tm.nameMap.Store(nameKey, toolID)
	}

	// Update name index
	if tm.serviceToolNames[serviceID] == nil {
		tm.serviceToolNames[serviceID] = make(map[string]struct{})
	}
	tm.serviceToolNames[serviceID][nameKey] = struct{}{}

	tm.toolsMutex.Lock()
	tm.cachedTools = nil
	tm.cachedMCPTools = nil
	tm.toolsMutex.Unlock()

	if tm.mcpServer != nil {
		mcpTool, err := ConvertProtoToMCPTool(tool.Tool())
		if err != nil {
			return fmt.Errorf("failed to convert proto tool to mcp tool: %w", err)
		}

		// Enforce namespacing for the MCP server registration as well
		if tool.Tool().GetServiceId() != "" {
			mcpTool.Name = tool.Tool().GetServiceId() + "." + tool.Tool().GetName()
		}

		if tool.Tool().GetInputSchema() != nil {
			// ⚡ BOLT: Direct Struct -> Map conversion avoids JSON serialization overhead
			// Randomized Selection from Top 5 High-Impact Targets
			mcpTool.InputSchema = tool.Tool().GetInputSchema().AsMap()
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
							// ⚡ Bolt Optimization: Use Zero-copy conversion for large JSON payloads
							Text: util.BytesToString(jsonResult),
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
// Summary: retrieves a tool from the manager by its fully qualified name.
//
// Parameters:
//   - toolName: string. The toolName.
//
// Returns:
//   - Tool: The Tool.
//   - bool: The bool.
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

// ListTools returns a slice containing all the tools currently registered with.
//
// Summary: returns a slice containing all the tools currently registered with.
//
// Parameters:
//   None.
//
// Returns:
//   - []Tool: The []Tool.
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

	return tm.rebuildCachedTools()
}

// rebuildCachedTools rebuilds the tools cache. Caller MUST hold tm.toolsMutex.Lock().
func (tm *Manager) rebuildCachedTools() []Tool {
	var tools []Tool
	tm.tools.Range(func(_ string, value Tool) bool {
		// Check service health
		serviceID := value.Tool().GetServiceId()
		// ⚡ Bolt Optimization: Use direct load to avoid expensive config cloning/stripping in GetServiceInfo
		if info, ok := tm.serviceInfo.Load(serviceID); ok {
			if info.HealthStatus == HealthStatusUnhealthy {
				return true // Skip unhealthy tools
			}
		}
		tools = append(tools, value)
		return true
	})
	tm.cachedTools = tools
	// When we rebuild the source of truth, the derived cache must be invalidated
	// if it wasn't already.
	tm.cachedMCPTools = nil
	return tools
}

// ListMCPTools returns a slice containing all the tools currently registered with.
//
// Summary: returns a slice containing all the tools currently registered with.
//
// Parameters:
//   None.
//
// Returns:
//   - []*mcp.Tool: The []*mcp.Tool.
func (tm *Manager) ListMCPTools() []*mcp.Tool {
	tm.toolsMutex.RLock()
	if tm.cachedMCPTools != nil {
		defer tm.toolsMutex.RUnlock()
		return tm.cachedMCPTools
	}
	tm.toolsMutex.RUnlock()

	tm.toolsMutex.Lock()
	defer tm.toolsMutex.Unlock()
	// Double check
	if tm.cachedMCPTools != nil {
		return tm.cachedMCPTools
	}

	tools := tm.cachedTools
	if tools == nil {
		tools = tm.rebuildCachedTools()
	}

	// Build MCPTools cache
	// We perform this under the lock to ensure that the cachedMCPTools matches cachedTools.
	mcpTools := make([]*mcp.Tool, 0, len(tools))
	for _, t := range tools {
		if mt := t.MCPTool(); mt != nil {
			// Enforce namespacing for the tool list to match AddTool and mcpServer registration
			if t.Tool().GetServiceId() != "" {
				expectedName := t.Tool().GetServiceId() + "." + t.Tool().GetName()
				if mt.Name != expectedName {
					// Update the name in the cached MCP tool
					mt.Name = expectedName
				}
			}
			mcpTools = append(mcpTools, mt)
		}
	}

	tm.cachedMCPTools = mcpTools
	return mcpTools
}

// ClearToolsForService removes all tools associated with a given service key from.
//
// Summary: removes all tools associated with a given service key from.
//
// Parameters:
//   - serviceID: string. The serviceID.
//
// Returns:
//   None.
func (tm *Manager) ClearToolsForService(serviceID string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	log := logging.GetLogger().With("serviceID", serviceID)
	log.Debug("Clearing existing tools for serviceID before reload/overwrite.")
	deletedCount := 0

	// ⚡ BOLT: Use secondary index for O(1) cleanup instead of O(N) map scan
	// Randomized Selection from Top 5 High-Impact Targets

	// 1. Cleanup Tools
	if toolIDs, ok := tm.serviceToolIDs[serviceID]; ok {
		for toolID := range toolIDs {
			tm.tools.Delete(toolID)
			deletedCount++
		}
		delete(tm.serviceToolIDs, serviceID)
	}

	// 2. Cleanup NameMap
	if names, ok := tm.serviceToolNames[serviceID]; ok {
		for name := range names {
			tm.nameMap.Delete(name)
		}
		delete(tm.serviceToolNames, serviceID)
	}

	if deletedCount > 0 {
		tm.toolsMutex.Lock()
		tm.cachedTools = nil
		tm.cachedMCPTools = nil
		tm.toolsMutex.Unlock()
	}
	log.Debug("Cleared tools for serviceID", "count", deletedCount)
}
