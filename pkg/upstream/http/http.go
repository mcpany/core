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

package http

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream"
	"github.com/mcpany/core/pkg/util"
	"github.com/mcpany/core/pkg/util/schemaconv"
	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"sort"

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

// HTTPUpstream implements the upstream.Upstream interface for services that are
// exposed via standard HTTP endpoints. It handles the registration of tools
// defined in the service configuration.
type HTTPUpstream struct {
	poolManager *pool.Manager
}

// NewHTTPUpstream creates a new instance of HTTPUpstream.
//
// poolManager is the connection pool manager to be used for managing HTTP
// connections.
func NewHTTPUpstream(poolManager *pool.Manager) upstream.Upstream {
	return &HTTPUpstream{
		poolManager: poolManager,
	}
}

// Register processes the configuration for an HTTP service, creates a connection
// pool for it, and then creates and registers tools for each call definition
// specified in the configuration.
func (u *HTTPUpstream) Register(
	ctx context.Context,
	serviceConfig *configv1.UpstreamServiceConfig,
	toolManager tool.ToolManagerInterface,
	promptManager prompt.PromptManagerInterface,
	resourceManager resource.ResourceManagerInterface,
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

	serviceID := sanitizedName // for internal use

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

	httpPool, err := NewHttpPool(maxIdleConnections, maxConnections, idleTimeout, serviceConfig)
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
	discoveredTools := u.createAndRegisterHTTPTools(ctx, serviceID, address, serviceConfig, toolManager, resourceManager, isReload)
	u.createAndRegisterPrompts(ctx, serviceID, serviceConfig, promptManager, isReload)
	log.Info("Registered HTTP service", "serviceID", serviceID, "toolsAdded", len(discoveredTools))

	return serviceID, discoveredTools, nil, nil
}

// createAndRegisterHTTPTools iterates through the HTTP call definitions in the
// service configuration, creates a new HTTPTool for each, and registers it
// with the tool manager.
func (u *HTTPUpstream) createAndRegisterHTTPTools(ctx context.Context, serviceID, address string, serviceConfig *configv1.UpstreamServiceConfig, toolManager tool.ToolManagerInterface, resourceManager resource.ResourceManagerInterface, isReload bool) []*configv1.ToolDefinition {
	log := logging.GetLogger()
	httpService := serviceConfig.GetHttpService()
	discoveredTools := make([]*configv1.ToolDefinition, 0, len(httpService.GetTools()))
	calls := httpService.GetCalls()
	callIDToDefinition := make(map[string]*configv1.ToolDefinition)
	for _, d := range httpService.GetTools() {
		callIDToDefinition[d.GetCallId()] = d
	}

	authenticator, err := auth.NewUpstreamAuthenticator(serviceConfig.GetUpstreamAuthentication())
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

		baseURL, err := url.Parse(address)
		if err != nil {
			log.Error("Failed to parse base URL", "address", address, "error", err)
			continue
		}

		endpointURL, err := url.Parse(httpDef.GetEndpointPath())
		if err != nil {
			log.Error("Failed to parse endpoint path", "path", httpDef.GetEndpointPath(), "error", err)
			continue
		}

		fullURL := baseURL.ResolveReference(endpointURL).String()

		if properties == nil {
			properties = &structpb.Struct{Fields: make(map[string]*structpb.Value)}
		}
		inputSchema := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"type":       structpb.NewStringValue("object"),
				"properties": structpb.NewStructValue(properties),
			},
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
			}.Build(),
		}.Build()

		log.DebugContext(ctx, "Tool protobuf is generated", "toolProto", newToolProto)

		httpTool := tool.NewHTTPTool(newToolProto, u.poolManager, serviceID, authenticator, httpDef, serviceConfig.GetResilience())
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
		}
	}

	return discoveredTools
}

func (u *HTTPUpstream) createAndRegisterPrompts(ctx context.Context, serviceID string, serviceConfig *configv1.UpstreamServiceConfig, promptManager prompt.PromptManagerInterface, isReload bool) {
	log := logging.GetLogger()
	httpService := serviceConfig.GetHttpService()
	for _, promptDef := range httpService.GetPrompts() {
		if promptDef.GetDisable() {
			log.Info("Skipping disabled prompt", "promptName", promptDef.GetName())
			continue
		}
		newPrompt := prompt.NewTemplatedPrompt(promptDef, serviceID)
		promptManager.AddPrompt(newPrompt)
		log.Info("Registered prompt", "prompt_name", newPrompt.Prompt().Name, "is_reload", isReload)
	}
}
