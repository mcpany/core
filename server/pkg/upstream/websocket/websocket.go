// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package websocket provides WebSocket upstream integration.
package websocket

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

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

// Upstream implements the upstream.Upstream interface for services that
// are exposed via a WebSocket connection. It manages a connection pool and
// registers tools based on the service configuration.
type Upstream struct {
	poolManager *pool.Manager
	serviceID   string
}

// Shutdown gracefully terminates the WebSocket upstream service by shutting down
// the associated connection pool.
//
// Parameters:
//   - ctx: The context for the shutdown operation.
//
// Returns:
//   - error: An error if the shutdown operation fails, or nil on success.
func (u *Upstream) Shutdown(_ context.Context) error {
	u.poolManager.Deregister(u.serviceID)
	return nil
}

// NewUpstream creates a new instance of WebsocketUpstream.
//
// Parameters:
//   - poolManager: The connection pool manager to be used for managing WebSocket connections.
//
// Returns:
//   - upstream.Upstream: A new Upstream instance for WebSocket services.
func NewUpstream(poolManager *pool.Manager) upstream.Upstream {
	return &Upstream{
		poolManager: poolManager,
	}
}

// Register processes the configuration for a WebSocket service. It creates a
// connection pool and registers tools for each call definition specified in the
// configuration.
//
// Parameters:
//   - ctx: The context for the registration process.
//   - serviceConfig: The configuration for the upstream service.
//   - toolManager: The manager where discovered tools will be registered.
//   - promptManager: The manager where discovered prompts will be registered.
//   - resourceManager: The manager where discovered resources will be registered.
//   - isReload: Indicates whether this is an initial registration or a reload.
//
// Returns:
//   - string: A unique service key.
//   - []*configv1.ToolDefinition: A list of discovered tool definitions.
//   - []*configv1.ResourceDefinition: A list of discovered resource definitions.
//   - error: An error if registration fails.
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

	u.serviceID = sanitizedName
	serviceID := u.serviceID

	websocketService := serviceConfig.GetWebsocketService()
	if websocketService == nil {
		return "", nil, nil, fmt.Errorf("websocket service config is nil")
	}

	address := websocketService.GetAddress()
	wsPool, err := NewPool(10, 300*time.Second, address)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to create websocket pool for %s: %w", serviceID, err)
	}
	u.poolManager.Register(serviceID, wsPool)

	info := &tool.ServiceInfo{
		Name:   serviceConfig.GetName(),
		Config: serviceConfig,
	}
	toolManager.AddServiceInfo(serviceID, info)

	discoveredTools := u.createAndRegisterWebsocketTools(ctx, serviceID, address, serviceConfig, toolManager, resourceManager, isReload)
	u.createAndRegisterPrompts(ctx, serviceID, serviceConfig, promptManager, isReload)
	log.Info("Registered Websocket service", "serviceID", serviceID, "toolsAdded", len(discoveredTools))

	return serviceID, discoveredTools, nil, nil
}

// createAndRegisterWebsocketTools iterates through the WebSocket call
// definitions in the service configuration, creates a new WebsocketTool for each,
// and registers it with the tool manager.
func (u *Upstream) createAndRegisterWebsocketTools(_ context.Context, serviceID, address string, serviceConfig *configv1.UpstreamServiceConfig, toolManager tool.ManagerInterface, resourceManager resource.ManagerInterface, _ bool) []*configv1.ToolDefinition {
	log := logging.GetLogger()
	websocketService := serviceConfig.GetWebsocketService()
	definitions := websocketService.GetTools()
	calls := websocketService.GetCalls()
	discoveredTools := make([]*configv1.ToolDefinition, 0, len(definitions))

	authenticator, err := auth.NewUpstreamAuthenticator(serviceConfig.GetUpstreamAuth())
	if err != nil {
		log.Error("Failed to create authenticator, proceeding without authentication", "serviceID", serviceID, "error", err)
		authenticator = nil
	}

	for i, definition := range definitions {
		if definition.GetDisable() {
			log.Info("Skipping disabled tool", "toolName", definition.GetName())
			continue
		}

		callID := definition.GetCallId()
		wsDef, ok := calls[callID]
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

		var inputSchema *structpb.Struct
		if wsDef.GetInputSchema() != nil && len(wsDef.GetInputSchema().GetFields()) > 0 {
			inputSchema = wsDef.GetInputSchema()
		} else {
			properties, required, err := schemaconv.ConfigSchemaToProtoProperties(wsDef.GetParameters())
			if err != nil {
				log.Error("Failed to convert schema to properties", "error", err)
				continue
			}

			if properties == nil {
				properties = &structpb.Struct{Fields: make(map[string]*structpb.Value)}
			}
			inputSchema = &structpb.Struct{
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
		}

		newToolProto := pb.Tool_builder{
			Name:                proto.String(toolNamePart),
			ServiceId:           proto.String(serviceID),
			UnderlyingMethodFqn: proto.String(fmt.Sprintf("WS %s", address)),
			Annotations: pb.ToolAnnotations_builder{
				Title:           proto.String(definition.GetTitle()),
				ReadOnlyHint:    proto.Bool(definition.GetReadOnlyHint()),
				DestructiveHint: proto.Bool(definition.GetDestructiveHint()),
				IdempotentHint:  proto.Bool(definition.GetIdempotentHint()),
				OpenWorldHint:   proto.Bool(definition.GetOpenWorldHint()),
				InputSchema:     inputSchema,
				OutputSchema:    wsDef.GetOutputSchema(),
			}.Build(),
		}.Build()

		wsTool := tool.NewWebsocketTool(newToolProto, u.poolManager, serviceID, authenticator, wsDef)
		if err := toolManager.AddTool(wsTool); err != nil {
			log.Error("Failed to add websocket tool", "error", err)
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
	for _, resourceDef := range websocketService.GetResources() {
		if resourceDef.GetDisable() {
			log.Info("Skipping disabled resource", "resourceName", resourceDef.GetName())
			continue
		}
		if resourceDef.GetDynamic() != nil {
			call := resourceDef.GetDynamic().GetWebsocketCall()
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

func (u *Upstream) createAndRegisterPrompts(_ context.Context, serviceID string, serviceConfig *configv1.UpstreamServiceConfig, promptManager prompt.ManagerInterface, isReload bool) {
	log := logging.GetLogger()
	websocketService := serviceConfig.GetWebsocketService()
	for _, promptDef := range websocketService.GetPrompts() {
		if promptDef.GetDisable() {
			log.Info("Skipping disabled prompt", "promptName", promptDef.GetName())
			continue
		}
		newPrompt := prompt.NewTemplatedPrompt(promptDef, serviceID)
		promptManager.AddPrompt(newPrompt)
		log.Info("Registered prompt", "prompt_name", newPrompt.Prompt().Name, "is_reload", isReload)
	}
}
