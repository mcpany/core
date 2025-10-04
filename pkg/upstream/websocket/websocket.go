/*
 * Copyright 2025 Author(s) of MCPXY
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

package websocket

import (
	"context"
	"errors"
	"fmt"
	"time"

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

// WebsocketUpstream implements the upstream.Upstream interface for services that
// are exposed via a WebSocket connection. It manages a connection pool and
// registers tools based on the service configuration.
type WebsocketUpstream struct {
	poolManager *pool.Manager
}

// NewWebsocketUpstream creates a new instance of WebsocketUpstream.
//
// poolManager is the connection pool manager to be used for managing WebSocket
// connections.
func NewWebsocketUpstream(poolManager *pool.Manager) upstream.Upstream {
	return &WebsocketUpstream{
		poolManager: poolManager,
	}
}

// Register processes the configuration for a WebSocket service. It creates a
// connection pool and registers tools for each call definition specified in the
// configuration.
func (u *WebsocketUpstream) Register(
	ctx context.Context,
	serviceConfig *configv1.UpstreamServiceConfig,
	toolManager tool.ToolManagerInterface,
	promptManager prompt.PromptManagerInterface,
	resourceManager resource.ResourceManagerInterface,
	isReload bool,
) (string, []*configv1.ToolDefinition, error) {
	if serviceConfig == nil {
		return "", nil, errors.New("service config is nil")
	}
	log := logging.GetLogger()
	serviceKey, err := util.GenerateServiceKey(serviceConfig.GetName())
	if err != nil {
		return "", nil, err
	}

	websocketService := serviceConfig.GetWebsocketService()
	if websocketService == nil {
		return "", nil, fmt.Errorf("websocket service config is nil")
	}

	address := websocketService.GetAddress()
	wsPool, err := NewWebsocketPool(10, 300*time.Second, address)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create websocket pool for %s: %w", serviceKey, err)
	}
	u.poolManager.Register(serviceKey, wsPool)

	info := &tool.ServiceInfo{
		Name:   serviceConfig.GetName(),
		Config: serviceConfig,
	}
	toolManager.AddServiceInfo(serviceKey, info)

	discoveredTools := u.createAndRegisterWebsocketTools(ctx, serviceKey, address, serviceConfig, toolManager, isReload)
	log.Info("Registered Websocket service", "serviceKey", serviceKey, "toolsAdded", len(discoveredTools))

	return serviceKey, discoveredTools, nil
}

func (u *WebsocketUpstream) createAndRegisterWebsocketTools(ctx context.Context, serviceKey, address string, serviceConfig *configv1.UpstreamServiceConfig, toolManager tool.ToolManagerInterface, isReload bool) []*configv1.ToolDefinition {
	log := logging.GetLogger()
	websocketService := serviceConfig.GetWebsocketService()
	definitions := websocketService.GetCalls()
	discoveredTools := make([]*configv1.ToolDefinition, 0, len(definitions))

	authenticator, err := auth.NewUpstreamAuthenticator(serviceConfig.GetUpstreamAuthentication())
	if err != nil {
		log.Error("Failed to create authenticator, proceeding without authentication", "serviceKey", serviceKey, "error", err)
		authenticator = nil
	}

	for i, wsDef := range definitions {
		toolNamePart := wsDef.GetOperationId()
		if toolNamePart == "" {
			sanitizedSummary := util.SanitizeOperationID(wsDef.GetDescription())
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

		for _, param := range wsDef.GetParameterMappings() {
			properties.Fields[param.GetInputParameterName()] = structpb.NewStringValue("")
		}

		newToolProto := pb.Tool_builder{
			Name:                proto.String(toolNamePart),
			ServiceId:           proto.String(serviceKey),
			UnderlyingMethodFqn: proto.String(fmt.Sprintf("WS %s", address)),
			InputSchema: pb.InputSchema_builder{
				Type:       proto.String("object"),
				Properties: properties,
			}.Build(),
		}.Build()

		wsTool := tool.NewWebsocketTool(newToolProto, u.poolManager, serviceKey, authenticator, wsDef)
		if err := toolManager.AddTool(wsTool); err != nil {
			log.Error("Failed to add websocket tool", "error", err)
			continue
		}

		discoveredTools = append(discoveredTools, configv1.ToolDefinition_builder{
			Name:        proto.String(wsDef.GetOperationId()),
			Description: proto.String(wsDef.GetDescription()),
		}.Build())
	}
	return discoveredTools
}
