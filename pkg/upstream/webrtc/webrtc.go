/*
 * Copyright 2025 Author(s) of MCPX
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

package webrtc

import (
	"context"
	"errors"
	"fmt"

	"github.com/mcpxy/mcpx/pkg/auth"
	"github.com/mcpxy/mcpx/pkg/logging"
	"github.com/mcpxy/mcpx/pkg/pool"
	"github.com/mcpxy/mcpx/pkg/prompt"
	"github.com/mcpxy/mcpx/pkg/resource"
	"github.com/mcpxy/mcpx/pkg/tool"
	"github.com/mcpxy/mcpx/pkg/upstream"
	"github.com/mcpxy/mcpx/pkg/util"
	configv1 "github.com/mcpxy/mcpx/proto/config/v1"
	pb "github.com/mcpxy/mcpx/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// WebrtcUpstream implements the upstream.Upstream interface for services that
// communicate over WebRTC data channels.
type WebrtcUpstream struct {
	poolManager *pool.Manager
}

// NewWebrtcUpstream creates a new instance of WebrtcUpstream.
//
// poolManager is the connection pool manager, though it is not currently used
// by the WebRTC upstream as connections are transient.
func NewWebrtcUpstream(poolManager *pool.Manager) upstream.Upstream {
	return &WebrtcUpstream{
		poolManager: poolManager,
	}
}

// Register processes the configuration for a WebRTC service, creating and
// registering tools for each call definition specified in the configuration.
func (u *WebrtcUpstream) Register(
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

	webrtcService := serviceConfig.GetWebrtcService()
	if webrtcService == nil {
		return "", nil, fmt.Errorf("webrtc service config is nil")
	}

	info := &tool.ServiceInfo{
		Name:   serviceConfig.GetName(),
		Config: serviceConfig,
	}
	toolManager.AddServiceInfo(serviceKey, info)

	address := webrtcService.GetAddress()
	discoveredTools := u.createAndRegisterWebrtcTools(ctx, serviceKey, address, serviceConfig, toolManager, isReload)
	log.Info("Registered WebRTC service", "serviceKey", serviceKey, "toolsAdded", len(discoveredTools))

	return serviceKey, discoveredTools, nil
}

func (u *WebrtcUpstream) createAndRegisterWebrtcTools(ctx context.Context, serviceKey, address string, serviceConfig *configv1.UpstreamServiceConfig, toolManager tool.ToolManagerInterface, isReload bool) []*configv1.ToolDefinition {
	log := logging.GetLogger()
	webrtcService := serviceConfig.GetWebrtcService()
	definitions := webrtcService.GetCalls()
	discoveredTools := make([]*configv1.ToolDefinition, 0, len(definitions))

	authenticator, err := auth.NewUpstreamAuthenticator(serviceConfig.GetUpstreamAuthentication())
	if err != nil {
		log.Error("Failed to create authenticator", "error", err)
		return nil
	}

	for i, wrtcDef := range definitions {
		toolNamePart := wrtcDef.GetOperationId()
		if toolNamePart == "" {
			sanitizedSummary := util.SanitizeOperationID(wrtcDef.GetDescription())
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

		for _, param := range wrtcDef.GetParameterMappings() {
			properties.Fields[param.GetInputParameterName()] = structpb.NewStringValue("")
		}

		newToolProto := pb.Tool_builder{
			Name:                proto.String(toolNamePart),
			ServiceId:           proto.String(serviceKey),
			UnderlyingMethodFqn: proto.String(fmt.Sprintf("WEBRTC %s", address)),
			InputSchema: pb.InputSchema_builder{
				Type:       proto.String("object"),
				Properties: properties,
			}.Build(),
		}.Build()

		wrtcTool, err := tool.NewWebrtcTool(newToolProto, u.poolManager, serviceKey, authenticator, wrtcDef)
		if err != nil {
			log.Error("Failed to create webrtc tool", "error", err)
			continue
		}

		if err := toolManager.AddTool(wrtcTool); err != nil {
			log.Error("Failed to add webrtc tool", "error", err)
			continue
		}

		discoveredTools = append(discoveredTools, configv1.ToolDefinition_builder{
			Name:        proto.String(wrtcDef.GetOperationId()),
			Description: proto.String(wrtcDef.GetDescription()),
		}.Build())
	}
	return discoveredTools
}
