// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http //nolint:revive,nolintlint // Package name 'http' is intentional for this directory structure.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/alexliesenfeld/health"
	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/doctor"
	mcphealth "github.com/mcpany/core/server/pkg/health"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/pkg/util/schemaconv"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// httpMethodToString converts the protobuf enum for an HTTP method into its
// corresponding string representation from the net/http package.
func httpMethodToString(method configv1.HttpCallDefinition_HttpMethod) (string, error) {
	switch method {
	case configv1.HttpCallDefinition_HTTP_METHOD_GET:
		return http.MethodGet, nil
	case configv1.HttpCallDefinition_HTTP_METHOD_POST:
		return http.MethodPost, nil
	case configv1.HttpCallDefinition_HTTP_METHOD_PUT:
		return http.MethodPut, nil
	case configv1.HttpCallDefinition_HTTP_METHOD_DELETE:
		return http.MethodDelete, nil
	case configv1.HttpCallDefinition_HTTP_METHOD_PATCH:
		return http.MethodPatch, nil
	default:
		return "", fmt.Errorf("unsupported HTTP method: %v", method)
	}
}

// Upstream implements the upstream.Upstream interface for services that are
// exposed via standard HTTP endpoints. It handles the registration of tools
// defined in the service configuration.
type Upstream struct {
	poolManager *pool.Manager
	serviceID   string
	address     string
	checker     health.Checker
}

// CheckHealth performs a health check on the upstream service.
func (u *Upstream) CheckHealth(ctx context.Context) error {
	if u.checker != nil {
		res := u.checker.Check(ctx)
		if res.Status != health.StatusUp {
			return fmt.Errorf("health check failed: %v", res)
		}
		return nil
	}
	if u.address == "" {
		return fmt.Errorf("no address configured")
	}
	return util.CheckConnection(ctx, u.address)
}

// Shutdown gracefully terminates the HTTP upstream service by shutting down the
// associated connection pool.
func (u *Upstream) Shutdown(_ context.Context) error {
	u.poolManager.Deregister(u.serviceID)
	return nil
}

// NewUpstream creates a new instance of Upstream.
//
// Parameters:
//   - poolManager: The connection pool manager to be used for managing HTTP connections.
//
// Returns:
//   - An implementation of the upstream.Upstream interface.
func NewUpstream(poolManager *pool.Manager) upstream.Upstream {
	return &Upstream{
		poolManager: poolManager,
	}
}

// Register processes the configuration for an HTTP service, creates a connection
// pool for it, and then creates and registers tools for each call definition
// specified in the configuration.
func (u *Upstream) Register(
	ctx context.Context,
	serviceConfig *configv1.UpstreamServiceConfig,
	toolManager tool.ManagerInterface,
	promptManager prompt.ManagerInterface,
	resourceManager resource.ManagerInterface,
	isReload bool,
) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	log := logging.GetLogger()

	// Calculate SHA256 for the ID
	h := sha256.New()
	h.Write([]byte(serviceConfig.GetName()))
	serviceConfig.SetId(hex.EncodeToString(h.Sum(nil)))

	// Sanitize the service name
	sanitizedName, err := util.SanitizeServiceName(serviceConfig.GetName())
	if err != nil {
		return "", nil, nil, err
	}
	serviceConfig.SetSanitizedName(sanitizedName)

	// Store new ID in local variable
	serviceID := sanitizedName

	u.checker = mcphealth.NewChecker(serviceConfig)

	if isReload {
		// Clear tools using the OLD service ID if available
		idToClear := u.serviceID
		if idToClear == "" {
			idToClear = serviceID
		}
		toolManager.ClearToolsForService(idToClear)
	}

	u.serviceID = serviceID

	httpService := serviceConfig.GetHttpService()
	if httpService == nil {
		return "", nil, nil, fmt.Errorf("http service config is nil")
	}

	address := httpService.GetAddress()
	if address == "" {
		return "", nil, nil, fmt.Errorf("http service address is required")
	}
	u.address = address

	uURL, err := url.ParseRequestURI(address)
	if err != nil {
		return "", nil, nil, fmt.Errorf("invalid http service address: %w", err)
	}
	if uURL.Scheme != "http" && uURL.Scheme != "https" {
		return "", nil, nil, fmt.Errorf("invalid http service address scheme: %s (must be http or https)", uURL.Scheme)
	}

	poolConfig := serviceConfig.GetConnectionPool()
	maxConnections := 100
	maxIdleConnections := 10
	idleTimeout := 90

	if poolConfig != nil {
		if poolConfig.GetMaxConnections() > 0 {
			maxConnections = int(poolConfig.GetMaxConnections())
		}
		if poolConfig.GetMaxIdleConnections() > 0 {
			maxIdleConnections = int(poolConfig.GetMaxIdleConnections())
		}
		if poolConfig.GetIdleTimeout() != nil && poolConfig.GetIdleTimeout().GetSeconds() > 0 {
			idleTimeout = int(poolConfig.GetIdleTimeout().GetSeconds())
		}
	}

	httpPool, err := NewHTTPPool(maxIdleConnections, maxConnections, time.Duration(idleTimeout)*time.Second, serviceConfig)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to create HTTP pool for %s: %w", serviceID, err)
	}
	u.poolManager.Register(serviceID, httpPool)

	info := &tool.ServiceInfo{
		Name:   serviceConfig.GetName(),
		Config: serviceConfig,
	}
	log.Debug("Registering HTTP service", "serviceID", serviceID, "info", info)
	toolManager.AddServiceInfo(serviceID, info)

	address = httpService.GetAddress()

	// Verify that the upstream is reachable.
	// This is a startup check to warn the user if the service configuration is incorrect or the service is down.
	if err := util.CheckConnection(ctx, address); err != nil {
		// Track 1: Friction Fighter - Automated Diagnosis
		// If basic connectivity fails, we run a full Doctor check to give actionable advice.
		log.Warn("⚠️  Upstream service appears unreachable. Running diagnostic...", "service", serviceConfig.GetName())
		diagnosis := doctor.CheckService(ctx, serviceConfig)

		// Format the diagnosis box
		border := strings.Repeat("-", 60)
		var msg string
		if diagnosis.Status == doctor.StatusOk {
			// Transient failure that recovered
			msg = fmt.Sprintf("\n%s\nCONNECTION FLAPPING: %s\n%s\nStatus:  %s\nDetails: %s\n%s\n",
				border,
				serviceConfig.GetName(),
				border,
				"RECOVERED",
				"Connection succeeded during diagnostic check (transient failure).",
				border,
			)
			log.Warn(msg)
		} else {
			// Confirmed failure
			msg = fmt.Sprintf("\n%s\nFAILED TO CONNECT: %s\n%s\nStatus:  %s\nDetails: %s\nError:   %v\n%s\n",
				border,
				serviceConfig.GetName(),
				border,
				diagnosis.Status,
				diagnosis.Message,
				diagnosis.Error,
				border,
			)
			// We log as Error to make it stand out, even though we continue startup
			log.Error(msg)
		}

		log.Warn("Tools will still be registered but will likely fail at runtime.",
			"service", serviceConfig.GetName(),
			"address", address,
		)
	}

	// Auto-discovery of tools from calls
	if serviceConfig.GetAutoDiscoverTool() {
		for callID := range httpService.GetCalls() {
			// Check if a tool already exists for this call
			exists := false
			for _, t := range httpService.GetTools() {
				if t.GetCallId() == callID {
					exists = true
					break
				}
			}
			if !exists {
				// Create a default tool definition
				newTool := configv1.ToolDefinition_builder{
					Name:        proto.String(callID),
					CallId:      proto.String(callID),
					Description: proto.String(fmt.Sprintf("Auto-discovered tool for call %s", callID)),
				}.Build()
				// Append to tools list so it gets picked up in createAndRegisterHTTPTools
				httpService.SetTools(append(httpService.GetTools(), newTool))
			}
		}
	}

	// Compile Export Policies once
	toolExportPolicy, err := tool.NewCompiledExportPolicy(serviceConfig.GetToolExportPolicy())
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to compile tool export policy: %w", err)
	}
	resourceExportPolicy, err := tool.NewCompiledExportPolicy(serviceConfig.GetResourceExportPolicy())
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to compile resource export policy: %w", err)
	}
	promptExportPolicy, err := tool.NewCompiledExportPolicy(serviceConfig.GetPromptExportPolicy())
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to compile prompt export policy: %w", err)
	}

	discoveredTools := u.createAndRegisterHTTPTools(ctx, serviceID, address, serviceConfig, toolManager, resourceManager, isReload, toolExportPolicy, resourceExportPolicy)
	u.createAndRegisterPrompts(ctx, serviceID, serviceConfig, promptManager, isReload, promptExportPolicy)
	log.Info("Registered HTTP service", "serviceID", serviceID, "toolsAdded", len(discoveredTools))

	return serviceID, discoveredTools, nil, nil
}

// createAndRegisterHTTPTools iterates through the HTTP call definitions in the
// service configuration, creates a new HTTPTool for each, and registers it
// with the tool manager.
func (u *Upstream) createAndRegisterHTTPTools(
	ctx context.Context,
	serviceID, address string,
	serviceConfig *configv1.UpstreamServiceConfig,
	toolManager tool.ManagerInterface,
	resourceManager resource.ManagerInterface,
	_ bool,
	toolExportPolicy *tool.CompiledExportPolicy,
	resourceExportPolicy *tool.CompiledExportPolicy,
) []*configv1.ToolDefinition {
	log := logging.GetLogger()
	httpService := serviceConfig.GetHttpService()
	discoveredTools := make([]*configv1.ToolDefinition, 0, len(httpService.GetTools()))
	calls := httpService.GetCalls()
	callIDToDefinition := make(map[string]*configv1.ToolDefinition)
	for _, d := range httpService.GetTools() {
		callIDToDefinition[d.GetCallId()] = d
	}

	authenticator, err := auth.NewUpstreamAuthenticator(serviceConfig.GetUpstreamAuth())
	if err != nil {
		log.Error("Failed to create authenticator, proceeding without authentication", "serviceID", serviceID, "error", err)
		authenticator = nil
	}

	// Sort call IDs for deterministic ordering
	sortedCallIDs := make([]string, 0, len(calls))
	for callID := range calls {
		sortedCallIDs = append(sortedCallIDs, callID)
	}
	sort.Strings(sortedCallIDs)

	callPolicies := serviceConfig.GetCallPolicies()
	compiledCallPolicies, err := tool.CompileCallPolicies(callPolicies)
	if err != nil {
		log.Error("Failed to compile call policies", "error", err)
		return nil
	}

	// Optimization: Parse baseURL once outside the loop to avoid redundant parsing for each call.
	baseURL, err := url.Parse(address)
	if err != nil {
		log.Error("Failed to parse base URL", "address", address, "error", err)
		return nil
	}

	for _, callID := range sortedCallIDs {
		httpDef := calls[callID]

		definition, ok := callIDToDefinition[callID]
		if !ok {
			log.Error("Tool definition not found for call", "call_id", callID)
			continue
		}

		if definition.GetDisable() {
			log.Info("Skipping disabled tool", "toolName", definition.GetName())
			continue
		}

		toolNamePart := definition.GetName()
		if toolNamePart == "" {
			sanitizedSummary := util.SanitizeOperationID(definition.GetDescription())
			if sanitizedSummary != "" {
				toolNamePart = sanitizedSummary
			} else {
				toolNamePart = fmt.Sprintf("op_%s", callID)
			}
		}

		// Check Export Policy
		if !tool.ShouldExportCompiled(toolNamePart, toolExportPolicy) {
			log.Info("Skipping non-exported tool", "toolName", toolNamePart)
			continue
		}

		allowed, err := tool.EvaluateCompiledCallPolicy(compiledCallPolicies, toolNamePart, callID, nil)
		if err != nil {
			log.Error("Failed to evaluate call policy", "error", err)
			continue
		}
		if !allowed {
			log.Info("Skipping blocked tool/call", "toolName", toolNamePart, "callID", callID)
			continue
		}

		// Process individual tool registration
		discoveredTool := u.processToolRegistration(ctx, serviceID, toolNamePart, definition, httpDef, baseURL, authenticator, serviceConfig, callPolicies, callID, toolManager)
		if discoveredTool != nil {
			discoveredTools = append(discoveredTools, discoveredTool)
		}
	}

	callIDToName := make(map[string]string)
	definitions := httpService.GetTools()
	for _, d := range definitions {
		callIDToName[d.GetCallId()] = d.GetName()
	}

	u.processResources(serviceID, serviceConfig, httpService, callIDToName, toolManager, resourceManager, resourceExportPolicy)

	return discoveredTools
}

func (u *Upstream) processToolRegistration(
	ctx context.Context,
	serviceID, toolNamePart string,
	definition *configv1.ToolDefinition,
	httpDef *configv1.HttpCallDefinition,
	baseURL *url.URL,
	authenticator auth.UpstreamAuthenticator,
	serviceConfig *configv1.UpstreamServiceConfig,
	callPolicies []*configv1.CallPolicy,
	callID string,
	toolManager tool.ManagerInterface,
) *configv1.ToolDefinition {
	log := logging.GetLogger()

	properties, requiredParams, err := schemaconv.ConfigSchemaToProtoProperties(httpDef.GetParameters())
	if err != nil {
		log.Error("Failed to convert schema to properties", "error", err)
		return nil
	}

	method, err := httpMethodToString(httpDef.GetMethod())
	if err != nil {
		log.Error("Skipping tool creation due to unsupported HTTP method", "toolName", toolNamePart, "error", err)
		return nil
	}

	fullURL := resolveFullURL(baseURL, httpDef.GetEndpointPath())
	if fullURL == "" {
		return nil
	}

	inputSchema := constructInputSchema(httpDef, properties, requiredParams)

	newToolProto := pb.Tool_builder{
		Name:                proto.String(toolNamePart),
		Description:         proto.String(definition.GetDescription()),
		ServiceId:           proto.String(serviceID),
		UnderlyingMethodFqn: proto.String(fmt.Sprintf("%s %s", method, fullURL)),
		Annotations: pb.ToolAnnotations_builder{
			Title:           proto.String(definition.GetTitle()),
			ReadOnlyHint:    proto.Bool(definition.GetReadOnlyHint()),
			DestructiveHint: proto.Bool(definition.GetDestructiveHint()),
			IdempotentHint:  proto.Bool(definition.GetIdempotentHint()),
			OpenWorldHint:   proto.Bool(definition.GetOpenWorldHint()),
			InputSchema:     inputSchema,
			OutputSchema:    httpDef.GetOutputSchema(),
		}.Build(),
	}.Build()

	log.DebugContext(ctx, "Tool protobuf is generated", "toolProto", newToolProto)

	httpTool := tool.NewHTTPTool(newToolProto, u.poolManager, serviceID, authenticator, httpDef, serviceConfig.GetResilience(), callPolicies, callID)
	if err := toolManager.AddTool(httpTool); err != nil {
		log.Error("Failed to add tool", "error", err)
		return nil
	}
	return configv1.ToolDefinition_builder{
		Name:        proto.String(toolNamePart),
		Description: proto.String(definition.GetDescription()),
	}.Build()
}

func resolveFullURL(baseURL *url.URL, rawEndpointPath string) string {
	log := logging.GetLogger()
	endpointURL, err := url.Parse(rawEndpointPath)
	if err != nil {
		// If parsing failed and the path starts with //, it might be because url.Parse
		// treated it as a scheme-relative URL with an invalid host (e.g., //foo%2Fbar).
		// Try parsing it as a path by prepending a slash.
		if strings.HasPrefix(rawEndpointPath, "//") {
			if parsedURL, err2 := url.Parse("/" + rawEndpointPath); err2 == nil {
				// Remove the leading slash we added
				parsedURL.Path = strings.TrimPrefix(parsedURL.Path, "/")
				if parsedURL.RawPath != "" {
					parsedURL.RawPath = strings.TrimPrefix(parsedURL.RawPath, "/")
				}
				endpointURL = parsedURL
				err = nil
			}
		}

		if err != nil {
			log.Error("Failed to parse endpoint path", "path", rawEndpointPath, "error", err)
			return ""
		}
	}

	// Fix for double slash bug (e.g., //foo).
	if endpointURL.Scheme == "" && strings.HasPrefix(rawEndpointPath, "//") {
		if endpointURL.Host != "" {
			prefix := "//"
			if endpointURL.User != nil {
				prefix += endpointURL.User.String() + "@"
			}
			prefix += endpointURL.Host

			endpointURL.Path = prefix + endpointURL.Path
			if endpointURL.RawPath != "" {
				endpointURL.RawPath = prefix + endpointURL.RawPath
			}
			endpointURL.Host = ""
			endpointURL.User = nil
		} else if endpointURL.Path == "" {
			endpointURL.Path = "//"
		}
	}

	baseForJoin := *baseURL
	relPath := endpointURL.Path
	relRawPath := endpointURL.RawPath

	if strings.HasPrefix(relPath, "/") {
		relPath = "." + relPath
		if relRawPath != "" {
			relRawPath = "." + relRawPath
		}
	}

	if relPath != "" || strings.HasSuffix(endpointURL.Path, "/") || strings.HasSuffix(rawEndpointPath, "/") {
		if !strings.HasSuffix(baseForJoin.Path, "/") {
			baseForJoin.Path += "/"
			if baseForJoin.RawPath != "" {
				baseForJoin.RawPath += "/"
			}
		}
	}

	relURL := &url.URL{
		Path:     relPath,
		RawPath:  relRawPath,
		Fragment: endpointURL.Fragment,
	}

	resolvedURL := baseForJoin.ResolveReference(relURL)
	resolvedURL.RawQuery = baseURL.RawQuery
	if endpointURL.Fragment == "" && baseURL.Fragment != "" {
		resolvedURL.Fragment = baseURL.Fragment
	}

	// Merge query parameters logic
	if strings.Trim(endpointURL.RawQuery, "&") != "" {
		// Simplified query merge logic for brevity in refactor
		// In a full refactor, this could be its own helper function
		// For now, we inline a simpler version or assume the logic is complex enough to stay here?
		// Actually, let's keep the manual parsing logic as it was critical for edge cases
		// But to reduce complexity, we extract it.
		resolvedURL.RawQuery = mergeQuery(resolvedURL.RawQuery, endpointURL.RawQuery)
	}
	return resolvedURL.String()
}

func mergeQuery(baseQuery, endpointQuery string) string {
	type queryPart struct {
		raw        string
		key        string
		isInvalid  bool
		keyDecoded bool
	}

	parseQueryManual := func(rawQuery string) []queryPart {
		var parts []queryPart
		if rawQuery == "" {
			return parts
		}
		for _, p := range strings.Split(rawQuery, "&") {
			if p == "" {
				continue
			}
			qp := queryPart{raw: p}
			var key, value string
			if idx := strings.Index(p, "="); idx >= 0 {
				key = p[:idx]
				value = p[idx+1:]
			} else {
				key = p
			}
			decodedKey, errKey := url.QueryUnescape(key)
			if errKey == nil {
				qp.key = decodedKey
				qp.keyDecoded = true
			}
			_, errVal := url.QueryUnescape(value)
			if errKey != nil || errVal != nil {
				qp.isInvalid = true
			}
			parts = append(parts, qp)
		}
		return parts
	}

	baseParts := parseQueryManual(baseQuery)
	endParts := parseQueryManual(endpointQuery)
	endPartsByKey := make(map[string][]string)
	for _, p := range endParts {
		if p.keyDecoded {
			endPartsByKey[p.key] = append(endPartsByKey[p.key], p.raw)
		}
	}

	finalParts := make([]string, 0, len(baseParts)+len(endParts))
	keysOverridden := make(map[string]bool)

	for _, bp := range baseParts {
		if bp.keyDecoded {
			if parts, ok := endPartsByKey[bp.key]; ok {
				if !keysOverridden[bp.key] {
					finalParts = append(finalParts, parts...)
					keysOverridden[bp.key] = true
				}
				continue
			}
		}
		finalParts = append(finalParts, bp.raw)
	}

	for _, ep := range endParts {
		if !ep.keyDecoded {
			finalParts = append(finalParts, ep.raw)
			continue
		}
		if !keysOverridden[ep.key] {
			finalParts = append(finalParts, ep.raw)
		}
	}
	return strings.Join(finalParts, "&")
}

func constructInputSchema(httpDef *configv1.HttpCallDefinition, properties *structpb.Struct, requiredParams []string) *structpb.Struct {
	log := logging.GetLogger()
	var inputSchema *structpb.Struct
	if httpDef.GetInputSchema() != nil && len(httpDef.GetInputSchema().GetFields()) > 0 {
		inputSchema = proto.Clone(httpDef.GetInputSchema()).(*structpb.Struct)
		if properties != nil {
			var existingProps map[string]*structpb.Value
			if propsVal, ok := inputSchema.Fields["properties"]; ok {
				if structVal := propsVal.GetStructValue(); structVal != nil {
					existingProps = structVal.Fields
				}
			}
			if existingProps == nil {
				existingProps = make(map[string]*structpb.Value)
				inputSchema.Fields["properties"] = structpb.NewStructValue(&structpb.Struct{Fields: existingProps})
			}
			for k, v := range properties.Fields {
				if _, ok := existingProps[k]; !ok {
					existingProps[k] = v
				}
			}
		}
		if len(requiredParams) > 0 {
			var existingRequired []any
			if reqVal, ok := inputSchema.Fields["required"]; ok {
				if listVal := reqVal.GetListValue(); listVal != nil {
					for _, v := range listVal.Values {
						if _, ok := v.GetKind().(*structpb.Value_StringValue); ok {
							existingRequired = append(existingRequired, v.GetStringValue())
						}
					}
				}
			}
			for _, reqParam := range requiredParams {
				exists := false
				for _, existing := range existingRequired {
					if existingStr, ok := existing.(string); ok && existingStr == reqParam {
						exists = true
						break
					}
				}
				if !exists {
					existingRequired = append(existingRequired, reqParam)
				}
			}
			if len(existingRequired) > 0 {
				requiredValue, err := structpb.NewList(existingRequired)
				if err != nil {
					log.Error("Failed to create required params list", "error", err)
				} else {
					inputSchema.Fields["required"] = structpb.NewListValue(requiredValue)
				}
			}
		}
	} else {
		if properties == nil {
			properties = &structpb.Struct{Fields: make(map[string]*structpb.Value)}
		}
		inputSchema = &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"type":       structpb.NewStringValue("object"),
				"properties": structpb.NewStructValue(properties),
			},
		}
		if len(requiredParams) > 0 {
			requiredList := make([]any, len(requiredParams))
			for i, v := range requiredParams {
				requiredList[i] = v
			}
			requiredValue, err := structpb.NewList(requiredList)
			if err == nil {
				inputSchema.Fields["required"] = structpb.NewListValue(requiredValue)
			} else {
				log.Error("Failed to create required params list", "error", err)
			}
		}
	}
	return inputSchema
}

func (u *Upstream) processResources(
	serviceID string,
	_ *configv1.UpstreamServiceConfig,
	httpService *configv1.HttpUpstreamService,
	callIDToName map[string]string,
	toolManager tool.ManagerInterface,
	resourceManager resource.ManagerInterface,
	resourceExportPolicy *tool.CompiledExportPolicy,
) {
	log := logging.GetLogger()
	for _, resourceDef := range httpService.GetResources() {
		if resourceDef.GetDisable() {
			log.Info("Skipping disabled resource", "resourceName", resourceDef.GetName())
			continue
		}
		// Check Export Policy
		if !tool.ShouldExportCompiled(resourceDef.GetName(), resourceExportPolicy) {
			log.Info("Skipping non-exported resource", "resourceName", resourceDef.GetName())
			continue
		}
		if resourceDef.GetDynamic() != nil {
			call := resourceDef.GetDynamic().GetHttpCall()
			if call == nil {
				continue
			}
			toolName, ok := callIDToName[call.GetId()]
			if !ok {
				log.Error("tool not found for dynamic resource", "call_id", call.GetId())
				continue
			}
			sanitizedToolName, err := util.SanitizeToolName(toolName)
			if err != nil {
				log.Error("Failed to sanitize tool name", "error", err)
				continue
			}
			tool, ok := toolManager.GetTool(serviceID + "." + sanitizedToolName)
			if !ok {
				log.Error("Tool not found for dynamic resource", "toolName", toolName)
				continue
			}
			dynamicResource, err := resource.NewDynamicResource(resourceDef, tool)
			if err != nil {
				log.Error("Failed to create dynamic resource", "error", err)
				continue
			}
			resourceManager.AddResource(dynamicResource)
		} else {
			// Static resource
			staticRes := resource.NewStaticResource(resourceDef, serviceID)
			resourceManager.AddResource(staticRes)
		}
	}
}

func (u *Upstream) createAndRegisterPrompts(
	_ context.Context,
	serviceID string,
	serviceConfig *configv1.UpstreamServiceConfig,
	promptManager prompt.ManagerInterface,
	isReload bool,
	promptExportPolicy *tool.CompiledExportPolicy,
) {
	log := logging.GetLogger()
	httpService := serviceConfig.GetHttpService()
	for _, promptDef := range httpService.GetPrompts() {
		if promptDef.GetDisable() {
			log.Info("Skipping disabled prompt", "promptName", promptDef.GetName())
			continue
		}
		// Check Export Policy
		if !tool.ShouldExportCompiled(promptDef.GetName(), promptExportPolicy) {
			log.Info("Skipping non-exported prompt", "promptName", promptDef.GetName())
			continue
		}
		newPrompt := prompt.NewTemplatedPrompt(promptDef, serviceID)
		promptManager.AddPrompt(newPrompt)
		log.Info("Registered prompt", "prompt_name", newPrompt.Prompt().Name, "is_reload", isReload)
	}
}
