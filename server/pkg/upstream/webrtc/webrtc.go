// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package webrtc provides WebRTC upstream integration.
package webrtc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/auth"
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

type sanitizer func(string) (string, error)

// Upstream implements the upstream.Upstream interface for services that
// communicate over WebRTC data channels.
//
// Summary: WebRTC upstream implementation.
type Upstream struct {
	poolManager       *pool.Manager
	toolNameSanitizer sanitizer
}

// Shutdown is a no-op for the WebRTC upstream, as connections are transient
// and not managed by a persistent pool.
//
// Summary: Shuts down the WebRTC upstream (no-op).
//
// Parameters:
//   - _ : context.Context. Unused.
//
// Returns:
//   - error: Always returns nil.
func (u *Upstream) Shutdown(_ context.Context) error {
	return nil
}

// NewUpstream creates a new instance of WebrtcUpstream.
//
// Summary: Creates a new instance of WebrtcUpstream.
//
// poolManager is the connection pool manager, though it is not currently used
// by the WebRTC upstream as connections are transient.
//
// Parameters:
//   - poolManager: *pool.Manager. The pool manager.
//
// Returns:
//   - upstream.Upstream: A new upstream instance.
func NewUpstream(poolManager *pool.Manager) upstream.Upstream {
	return &Upstream{
		poolManager:       poolManager,
		toolNameSanitizer: util.SanitizeToolName,
	}
}

// Register processes the configuration for a WebRTC service, creating and
// registering tools for each call definition specified in the configuration.
//
// Summary: Registers the WebRTC service and its tools.
//
// Parameters:
//   - ctx: context.Context. The context for the registration.
//   - serviceConfig: *configv1.UpstreamServiceConfig. The service configuration.
//   - toolManager: tool.ManagerInterface. The tool manager.
//   - promptManager: prompt.ManagerInterface. The prompt manager.
//   - resourceManager: resource.ManagerInterface. The resource manager.
//   - isReload: bool. Whether this is a reload.
//
// Returns:
//   - string: The service ID.
//   - []*configv1.ToolDefinition: A list of discovered tool definitions.
//   - []*configv1.ResourceDefinition: A list of discovered resource definitions.
//   - error: An error if the operation fails.
func (u *Upstream) Register(
	ctx context.Context,
	serviceConfig *configv1.UpstreamServiceConfig,
	toolManager tool.ManagerInterface,
	promptManager prompt.ManagerInterface,
	resourceManager resource.ManagerInterface,
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
	u.createAndRegisterPrompts(ctx, serviceID, serviceConfig, promptManager, isReload)
	log.Info("Registered WebRTC service", "serviceID", serviceID, "toolsAdded", len(discoveredTools))

	return serviceID, discoveredTools, nil, nil
}

// createAndRegisterWebrtcTools iterates through the WebRTC call definitions in
// the service configuration, creates a new WebrtcTool for each, and registers it
// with the tool manager.
func (u *Upstream) createAndRegisterWebrtcTools(_ context.Context, serviceID, address string, serviceConfig *configv1.UpstreamServiceConfig, toolManager tool.ManagerInterface, resourceManager resource.ManagerInterface, _ bool) []*configv1.ToolDefinition {
	log := logging.GetLogger()
	webrtcService := serviceConfig.GetWebrtcService()
	definitions := webrtcService.GetTools()
	calls := webrtcService.GetCalls()
	discoveredTools := make([]*configv1.ToolDefinition, 0, len(definitions))

	authenticator, err := auth.NewUpstreamAuthenticator(serviceConfig.GetUpstreamAuth())
	if err != nil {
		log.Error("Failed to create authenticator", "error", err)
		return nil
	}

	for i, definition := range definitions {
		if definition.GetDisable() {
			log.Info("Skipping disabled tool", "toolName", definition.GetName())
			continue
		}

		callID := definition.GetCallId()
		wrtcDef, ok := calls[callID]
		if !ok {
			log.Error("Call definition not found for tool", "call_id", callID, "tool_name", definition.GetName())
			continue
		}
		toolNamePart := definition.GetName()
		if toolNamePart == "" {
			sanitizedSummary := util.SanitizeOperationID(definition.GetDescription())
			if sanitizedSummary != "" {
				toolNamePart = sanitizedSummary
			} else {
				toolNamePart = fmt.Sprintf("op%d", i)
			}
		}

		properties, required, err := schemaconv.ConfigSchemaToProtoProperties(wrtcDef.GetParameters())
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

		if len(required) > 0 {
			requiredVals := make([]*structpb.Value, len(required))
			for i, v := range required {
				requiredVals[i] = structpb.NewStringValue(v)
			}
			inputSchema.Fields["required"] = structpb.NewListValue(&structpb.ListValue{Values: requiredVals})
		}

		newToolProto := pb.Tool_builder{
			Name:                proto.String(toolNamePart),
			ServiceId:           proto.String(serviceID),
			UnderlyingMethodFqn: proto.String(fmt.Sprintf("WEBRTC %s", address)),
			Annotations: pb.ToolAnnotations_builder{
				Title:           proto.String(definition.GetTitle()),
				ReadOnlyHint:    proto.Bool(definition.GetReadOnlyHint()),
				DestructiveHint: proto.Bool(definition.GetDestructiveHint()),
				IdempotentHint:  proto.Bool(definition.GetIdempotentHint()),
				OpenWorldHint:   proto.Bool(definition.GetOpenWorldHint()),
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
			Name:        proto.String(definition.GetName()),
			Description: proto.String(definition.GetDescription()),
		}.Build())
	}

	callIDToName := make(map[string]string)
	for _, d := range definitions {
		callIDToName[d.GetCallId()] = d.GetName()
	}
	for _, resourceDef := range webrtcService.GetResources() {
		if resourceDef.GetDisable() {
			log.Info("Skipping disabled resource", "resourceName", resourceDef.GetName())
			continue
		}
		if resourceDef.GetDynamic() != nil {
			call := resourceDef.GetDynamic().GetWebrtcCall()
			if call == nil {
				continue
			}
			toolName, ok := callIDToName[call.GetId()]
			if !ok {
				log.Error("tool not found for dynamic resource", "call_id", call.GetId())
				continue
			}
			sanitizedToolName, err := u.toolNameSanitizer(toolName)
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

func (u *Upstream) createAndRegisterPrompts(_ context.Context, serviceID string, serviceConfig *configv1.UpstreamServiceConfig, promptManager prompt.ManagerInterface, isReload bool) {
	log := logging.GetLogger()
	webrtcService := serviceConfig.GetWebrtcService()
	for _, promptDef := range webrtcService.GetPrompts() {
		if promptDef.GetDisable() {
			log.Info("Skipping disabled prompt", "promptName", promptDef.GetName())
			continue
		}
		newPrompt := prompt.NewTemplatedPrompt(promptDef, serviceID)
		promptManager.AddPrompt(newPrompt)
		log.Info("Registered prompt", "prompt_name", newPrompt.Prompt().Name, "is_reload", isReload)
	}
}
