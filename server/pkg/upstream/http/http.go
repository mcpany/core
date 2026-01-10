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

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/alexliesenfeld/health"
	"github.com/mcpany/core/server/pkg/auth"
	healthChecker "github.com/mcpany/core/server/pkg/health"
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

	u.serviceID = sanitizedName
	serviceID := u.serviceID

	if isReload {
		toolManager.ClearToolsForService(serviceID)
	}

	httpService := serviceConfig.GetHttpService()
	if httpService == nil {
		return "", nil, nil, fmt.Errorf("http service config is nil")
	}

	address := httpService.GetAddress()
	if address == "" {
		return "", nil, nil, fmt.Errorf("http service address is required")
	}
	if _, err := url.ParseRequestURI(address); err != nil {
		return "", nil, nil, fmt.Errorf("invalid http service address: %w", err)
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

	// Perform an initial health check to ensure the service is reachable.
	// This provides immediate feedback to the user if the configuration is incorrect
	// or the service is down.
	checker := healthChecker.NewChecker(serviceConfig)
	if checker != nil {
		result := checker.Check(ctx)
		if result.Status != health.StatusUp {
			log.Warn("Initial health check failed", "serviceID", serviceID, "status", result.Status, "error", result.Details)

			// Attempt to get a more descriptive error message by running the connection check directly.
			// This ensures the user gets immediate feedback about why the service is unreachable.
			if err := healthChecker.CheckConnection(ctx, address); err != nil {
				return "", nil, nil, fmt.Errorf("connectivity check failed: %w", err)
			}

			// If CheckConnection passed but the general health check failed (e.g., HTTP status code mismatch),
			// we should also fail the registration to prevent silent failures.
			return "", nil, nil, fmt.Errorf("health check failed (status: %s). Check logs for details.", result.Status)
		}
	}

	info := &tool.ServiceInfo{
		Name:   serviceConfig.GetName(),
		Config: serviceConfig,
	}
	log.Debug("Registering HTTP service", "serviceID", serviceID, "info", info)
	toolManager.AddServiceInfo(serviceID, info)

	address = httpService.GetAddress()

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
				newTool := &configv1.ToolDefinition{
					Name:        proto.String(callID),
					CallId:      proto.String(callID),
					Description: proto.String(fmt.Sprintf("Auto-discovered tool for call %s", callID)),
				}
				// Append to tools list so it gets picked up in createAndRegisterHTTPTools
				httpService.Tools = append(httpService.Tools, newTool)
			}
		}
	}

	discoveredTools := u.createAndRegisterHTTPTools(ctx, serviceID, address, serviceConfig, toolManager, resourceManager, isReload)
	u.createAndRegisterPrompts(ctx, serviceID, serviceConfig, promptManager, isReload)
	log.Info("Registered HTTP service", "serviceID", serviceID, "toolsAdded", len(discoveredTools))

	return serviceID, discoveredTools, nil, nil
}

// createAndRegisterHTTPTools iterates through the HTTP call definitions in the
// service configuration, creates a new HTTPTool for each, and registers it
// with the tool manager.
func (u *Upstream) createAndRegisterHTTPTools(ctx context.Context, serviceID, address string, serviceConfig *configv1.UpstreamServiceConfig, toolManager tool.ManagerInterface, resourceManager resource.ManagerInterface, _ bool) []*configv1.ToolDefinition { //nolint:gocyclo // High complexity due to tool discovery logic
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
		if !tool.ShouldExport(toolNamePart, serviceConfig.GetToolExportPolicy()) {
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

		properties, err := schemaconv.ConfigSchemaToProtoProperties(httpDef.GetParameters())
		if err != nil {
			log.Error("Failed to convert schema to properties", "error", err)
			continue
		}

		requiredParams := []string{}
		for _, param := range httpDef.GetParameters() {
			paramSchema := param.GetSchema()
			if paramSchema.GetIsRequired() {
				requiredParams = append(requiredParams, paramSchema.GetName())
			}
		}

		method, err := httpMethodToString(httpDef.GetMethod())
		if err != nil {
			log.Error("Skipping tool creation due to unsupported HTTP method", "toolName", toolNamePart, "error", err)
			continue
		}

		rawEndpointPath := httpDef.GetEndpointPath()
		endpointURL, err := url.Parse(rawEndpointPath)
		if err != nil {
			log.Error("Failed to parse endpoint path", "path", rawEndpointPath, "error", err)
			continue
		}

		// Fix for double slash bug (e.g., //foo).
		// If the endpoint path starts with double slashes but has no scheme (e.g., //foo),
		// url.Parse treats it as a scheme-relative URL (host="foo", path="").
		// We want to treat it as a path relative to the base URL.
		if endpointURL.Scheme == "" && strings.HasPrefix(rawEndpointPath, "//") {
			// Trim all leading slashes to ensure it's treated as a relative path
			trimmed := strings.TrimLeft(rawEndpointPath, "/")
			endpointURL, err = url.Parse(trimmed)
			if err != nil {
				log.Error("Failed to parse corrected endpoint path", "path", trimmed, "error", err)
				continue
			}
		}

		// Ensure baseURL has a trailing slash so ResolveReference appends to it,
		// BUT ONLY IF we are actually appending a path segment.
		baseForJoin := *baseURL

		// Make endpoint path relative
		relPath := strings.TrimPrefix(endpointURL.Path, "/")

		// If the endpoint path has segments (not empty and not just /), force slash on base.
		// If endpoint path is empty, we DON'T want to force a slash on base, because
		// we want http://host/api + "" -> http://host/api
		// But ResolveReference behavior is:
		// Base: http://host/api, Rel: foo -> http://host/foo (Replaces last segment!)
		// Base: http://host/api/, Rel: foo -> http://host/api/foo
		// So if we have a relPath, we MUST ensure base has slash.
		//
		// Special case: If endpointURL.Path is exactly "/", we want to treat it as a slash
		// relative to the base, effectively forcing a trailing slash on the base.
		// But TrimPrefix(..., "/") makes it empty string.
		// So we check if relPath is NOT empty OR if endpointURL.Path ended with a slash.
		// We also check rawEndpointPath because if we trimmed slashes from //, endpointURL.Path
		// might be empty, but we want to preserve the trailing slash implied by the original input.
		if relPath != "" || strings.HasSuffix(endpointURL.Path, "/") || strings.HasSuffix(rawEndpointPath, "/") {
			if !strings.HasSuffix(baseForJoin.Path, "/") {
				baseForJoin.Path += "/"
				if baseForJoin.RawPath != "" {
					baseForJoin.RawPath += "/"
				}
			}
		}
		relRawPath := ""
		if endpointURL.RawPath != "" {
			relRawPath = strings.TrimPrefix(endpointURL.RawPath, "/")
		}

		// Construct a relative URL using RawPath to preserve encoding
		relURL := &url.URL{
			Path:     relPath,
			RawPath:  relRawPath,
			Fragment: endpointURL.Fragment,
		}

		resolvedURL := baseForJoin.ResolveReference(relURL)
		// ResolveReference discards the base query, so we restore it
		resolvedURL.RawQuery = baseURL.RawQuery
		// ResolveReference also discards the base fragment if the reference has none (empty string).
		// If the endpoint URL didn't specify a fragment, we should preserve the base fragment.
		if endpointURL.Fragment == "" && baseURL.Fragment != "" {
			resolvedURL.Fragment = baseURL.Fragment
		}
		// Merge query parameters, allowing endpoint parameters to override base parameters
		endpointQuery := endpointURL.Query()
		if len(endpointQuery) > 0 {
			// Capture which keys are "flags" (no equals sign) in the base URL
			// by manually inspecting RawQuery, because url.Values does not distinguish.
			originalFlags := make(map[string]bool)
			// Helper to extract flags from a raw query string
			extractFlags := func(rawQuery string) {
				if rawQuery != "" {
					for _, param := range strings.Split(rawQuery, "&") {
						if !strings.Contains(param, "=") && len(param) > 0 {
							// Decode the key to store it canonically.
							// We must handle cases where RawQuery uses %20 but Encode uses +.
							if decodedKey, err := url.QueryUnescape(param); err == nil {
								originalFlags[decodedKey] = true
							}
						}
					}
				}
			}

			extractFlags(baseURL.RawQuery)
			extractFlags(endpointURL.RawQuery)

			query := resolvedURL.Query()
			for k, v := range endpointQuery {
				query[k] = v
			}
			encoded := query.Encode()

			// Post-process to restore flag style for keys that were originally flags
			// and still have empty values (not overridden by a non-empty value).
			// Encode() sorts keys, so we can split by "&" and process each part.
			if len(originalFlags) > 0 {
				parts := strings.Split(encoded, "&")
				for i, part := range parts {
					// Check if this part ends with "=" (meaning empty value)
					if strings.HasSuffix(part, "=") {
						encodedKey := strings.TrimSuffix(part, "=")
						decodedKey, err := url.QueryUnescape(encodedKey)
						if err == nil {
							// Check if this key was originally a flag
							if originalFlags[decodedKey] {
								// Check if the current value in query map is actually empty.
								vals := query[decodedKey]
								if len(vals) == 0 || (len(vals) == 1 && vals[0] == "") {
									parts[i] = encodedKey
								}
							}
						}
					}
				}
				encoded = strings.Join(parts, "&")
			}
			resolvedURL.RawQuery = encoded
		}
		fullURL := resolvedURL.String()

		var inputSchema *structpb.Struct
		if httpDef.GetInputSchema() != nil && len(httpDef.GetInputSchema().GetFields()) > 0 {
			// Clone the input schema to avoid modifying the configuration object
			inputSchema = proto.Clone(httpDef.GetInputSchema()).(*structpb.Struct)

			// Ensure properties from 'parameters' are in 'inputSchema.properties'
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

			// Ensure required parameters from 'parameters' are also in 'inputSchema.required'
			if len(requiredParams) > 0 {
				var existingRequired []any
				if reqVal, ok := inputSchema.Fields["required"]; ok {
					if listVal := reqVal.GetListValue(); listVal != nil {
						for _, v := range listVal.Values {
							existingRequired = append(existingRequired, v.GetStringValue())
						}
					}
				}

				// Merge requiredParams into existingRequired
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

				// Write back to inputSchema
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
				if err != nil {
					log.Error("Failed to create required params list", "error", err)
					continue
				}
				inputSchema.Fields["required"] = structpb.NewListValue(requiredValue)
			}
		}

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
			continue
		}
		discoveredTools = append(discoveredTools, configv1.ToolDefinition_builder{
			Name:        proto.String(toolNamePart),
			Description: proto.String(definition.GetDescription()),
		}.Build())
	}

	callIDToName := make(map[string]string)
	definitions := httpService.GetTools()
	for _, d := range definitions {
		callIDToName[d.GetCallId()] = d.GetName()
	}

	for _, resourceDef := range httpService.GetResources() {
		if resourceDef.GetDisable() {
			log.Info("Skipping disabled resource", "resourceName", resourceDef.GetName())
			continue
		}
		// Check Export Policy
		if !tool.ShouldExport(resourceDef.GetName(), serviceConfig.GetResourceExportPolicy()) {
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

	return discoveredTools
}

func (u *Upstream) createAndRegisterPrompts(_ context.Context, serviceID string, serviceConfig *configv1.UpstreamServiceConfig, promptManager prompt.ManagerInterface, isReload bool) {
	log := logging.GetLogger()
	httpService := serviceConfig.GetHttpService()
	for _, promptDef := range httpService.GetPrompts() {
		if promptDef.GetDisable() {
			log.Info("Skipping disabled prompt", "promptName", promptDef.GetName())
			continue
		}
		// Check Export Policy
		if !tool.ShouldExport(promptDef.GetName(), serviceConfig.GetPromptExportPolicy()) {
			log.Info("Skipping non-exported prompt", "promptName", promptDef.GetName())
			continue
		}
		newPrompt := prompt.NewTemplatedPrompt(promptDef, serviceID)
		promptManager.AddPrompt(newPrompt)
		log.Info("Registered prompt", "prompt_name", newPrompt.Prompt().Name, "is_reload", isReload)
	}
}
