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

package webrtc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

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
) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	if serviceConfig == nil {
		return "", nil, nil, errors.New("service config is nil")
	}
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

	webrtcService := serviceConfig.GetWebrtcService()
	if webrtcService == nil {
		return "", nil, nil, fmt.Errorf("webrtc service config is nil")
	}

	info := &tool.ServiceInfo{
		Name:   serviceConfig.GetName(),
		Config: serviceConfig,
	}
	toolManager.AddServiceInfo(serviceID, info)

	address := webrtcService.GetAddress()
	discoveredTools := u.createAndRegisterWebrtcTools(ctx, serviceID, address, serviceConfig, toolManager, resourceManager, isReload)
	log.Info("Registered WebRTC service", "serviceID", serviceID, "toolsAdded", len(discoveredTools))

	return serviceID, discoveredTools, nil, nil
}

// createAndRegisterWebrtcTools iterates through the WebRTC call definitions in
// the service configuration, creates a new WebrtcTool for each, and registers it
// with the tool manager.
func (u *WebrtcUpstream) createAndRegisterWebrtcTools(ctx context.Context, serviceID, address string, serviceConfig *configv1.UpstreamServiceConfig, toolManager tool.ToolManagerInterface, resourceManager resource.ResourceManagerInterface, isReload bool) []*configv1.ToolDefinition {
	log := logging.GetLogger()
	webrtcService := serviceConfig.GetWebrtcService()
	definitions := webrtcService.GetTools()
	discoveredTools := make([]*configv1.ToolDefinition, 0, len(definitions))

	authenticator, err := auth.NewUpstreamAuthenticator(serviceConfig.GetUpstreamAuthentication())
	if err != nil {
		log.Error("Failed to create authenticator", "error", err)
		return nil
	}

	for i, toolDefinition := range definitions {
		wrtcDef := toolDefinition.GetCall()
		schema := wrtcDef.GetSchema()
		toolNamePart := schema.GetName()
		if toolNamePart == "" {
			sanitizedSummary := util.SanitizeOperationID(schema.GetDescription())
			if sanitizedSummary != "" {
				toolNamePart = sanitizedSummary
			} else {
				toolNamePart = fmt.Sprintf("op%d", i)
			}
		}

		properties, err := schemaconv.ConfigSchemaToProtoProperties(wrtcDef.GetParameters())
		if err != nil {
			log.Error("Failed to convert schema to properties", "error", err)
			continue
		}

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
			ServiceId:           proto.String(serviceID),
			UnderlyingMethodFqn: proto.String(fmt.Sprintf("WEBRTC %s", address)),
			Annotations: pb.ToolAnnotations_builder{
				Title:           proto.String(schema.GetTitle()),
				ReadOnlyHint:    proto.Bool(schema.GetReadOnlyHint()),
				DestructiveHint: proto.Bool(schema.GetDestructiveHint()),
				IdempotentHint:  proto.Bool(schema.GetIdempotentHint()),
				OpenWorldHint:   proto.Bool(schema.GetOpenWorldHint()),
				InputSchema:     inputSchema,
			}.Build(),
		}.Build()

		wrtcTool, err := tool.NewWebrtcTool(newToolProto, u.poolManager, serviceID, authenticator, wrtcDef)
		if err != nil {
			log.Error("Failed to create webrtc tool", "error", err)
			continue
		}

		if err := toolManager.AddTool(wrtcTool); err != nil {
			log.Error("Failed to add webrtc tool", "error", err)
			continue
		}

		discoveredTools = append(discoveredTools, configv1.ToolDefinition_builder{
			Name:        proto.String(schema.GetName()),
			Description: proto.String(schema.GetDescription()),
		}.Build())
	}

	for _, resourceDef := range webrtcService.GetResources() {
		if resourceDef.GetDynamic() != nil {
			toolName := resourceDef.GetDynamic().GetWebrtcCall().GetSchema().GetName()
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
