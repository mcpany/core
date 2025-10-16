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

package http

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mcpxy/core/pkg/auth"
	"github.com/mcpxy/core/pkg/logging"
	"github.com/mcpxy/core/pkg/pool"
	"github.com/mcpxy/core/pkg/prompt"
	"github.com/mcpxy/core/pkg/resource"
	"github.com/mcpxy/core/pkg/tool"
	"github.com/mcpxy/core/pkg/upstream"
	"github.com/mcpxy/core/pkg/util"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	pb "github.com/mcpxy/core/proto/mcp_router/v1"
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
) (string, []*configv1.ToolDefinition, error) {
	log := logging.GetLogger()
	serviceKey, err := util.GenerateServiceKey(serviceConfig.GetName())
	if err != nil {
		return "", nil, err
	}

	if isReload {
		toolManager.ClearToolsForService(serviceKey)
	}

	httpService := serviceConfig.GetHttpService()
	if httpService == nil {
		return "", nil, fmt.Errorf("http service config is nil")
	}

	poolConfig := serviceConfig.GetConnectionPool()
	maxConnections := 10
	maxIdleConnections := 0
	idleTimeout := 300

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

	httpPool, err := NewHttpPool(maxIdleConnections, maxConnections, idleTimeout)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create HTTP pool for %s: %w", serviceKey, err)
	}
	u.poolManager.Register(serviceKey, httpPool)

	info := &tool.ServiceInfo{
		Name:   serviceConfig.GetName(),
		Config: serviceConfig,
	}
	toolManager.AddServiceInfo(serviceKey, info)

	address := httpService.GetAddress()
	discoveredTools := u.createAndRegisterHTTPTools(ctx, serviceKey, address, serviceConfig, toolManager, isReload)
	log.Info("Registered HTTP service", "serviceKey", serviceKey, "toolsAdded", len(discoveredTools))

	return serviceKey, discoveredTools, nil
}

// createAndRegisterHTTPTools iterates through the HTTP call definitions in the
// service configuration, creates a new HTTPTool for each, and registers it
// with the tool manager.
func (u *HTTPUpstream) createAndRegisterHTTPTools(ctx context.Context, serviceKey, address string, serviceConfig *configv1.UpstreamServiceConfig, toolManager tool.ToolManagerInterface, isReload bool) []*configv1.ToolDefinition {
	log := logging.GetLogger()
	discoveredTools := make([]*configv1.ToolDefinition, 0, len(serviceConfig.GetHttpService().GetCalls()))
	httpService := serviceConfig.GetHttpService()
	definitions := httpService.GetCalls()

	authenticator, err := auth.NewUpstreamAuthenticator(serviceConfig.GetUpstreamAuthentication())
	if err != nil {
		log.Error("Failed to create authenticator, proceeding without authentication", "serviceKey", serviceKey, "error", err)
		authenticator = nil
	}

	for i, httpDef := range definitions {
		toolNamePart := httpDef.GetOperationId()
		if toolNamePart == "" {
			sanitizedSummary := util.SanitizeOperationID(httpDef.GetDescription())
			if sanitizedSummary != "" {
				toolNamePart = sanitizedSummary
			} else {
				toolNamePart = fmt.Sprintf("op%d", i)
			}
		}

		properties, err := structpb.NewStruct(make(map[string]interface{}))
		if err != nil {
			log.Error("Failed to create properties struct", "error", err)
			continue
		}

		for _, param := range httpDef.GetParameterMappings() {
			properties.Fields[param.GetInputParameterName()] = structpb.NewStringValue("")
		}

		method, err := httpMethodToString(httpDef.GetMethod())
		if err != nil {
			log.Error("Skipping tool creation due to unsupported HTTP method", "toolName", toolNamePart, "error", err)
			continue
		}

		fullURL, err := url.JoinPath(address, httpDef.GetEndpointPath())
		if err != nil {
			log.Error("Failed to join URL parts", "address", address, "endpoint", httpDef.GetEndpointPath(), "error", err)
			continue
		}
		newToolProto := pb.Tool_builder{
			Name:                proto.String(toolNamePart),
			ServiceId:           proto.String(serviceKey),
			UnderlyingMethodFqn: proto.String(fmt.Sprintf("%s %s", method, fullURL)),
			InputSchema: pb.InputSchema_builder{
				Type:       proto.String("object"),
				Properties: properties,
			}.Build(),
		}.Build()

		httpTool := tool.NewHTTPTool(newToolProto, u.poolManager, serviceKey, authenticator, httpDef)
		if err := toolManager.AddTool(httpTool); err != nil {
			log.Error("Failed to add tool", "error", err)
			continue
		}
		discoveredTools = append(discoveredTools, configv1.ToolDefinition_builder{
			Name:        proto.String(httpDef.GetOperationId()),
			Description: proto.String(httpDef.GetDescription()),
		}.Build())
	}
	return discoveredTools
}
