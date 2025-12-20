// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package command provides command execution functionality.
package command

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/mcpany/core/pkg/logging"
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

// Upstream implements the upstream.Upstream interface for services that
// are exposed as command-line tools. It discovers and registers tools based on
// a list of commands defined in the service configuration.
type Upstream struct{}

// Shutdown implements the upstream.Upstream interface.
func (u *Upstream) Shutdown(_ context.Context) error {
	// Noop for command upstream
	return nil
}

// NewUpstream creates a new instance of CommandUpstream.
func NewUpstream() upstream.Upstream {
	return &Upstream{}
}

// Register processes the configuration for a command-line service, creates a
// new tool for each defined command, and registers them with the tool manager.
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

	serviceID := sanitizedName // for internal use

	commandLineService := serviceConfig.GetCommandLineService()
	if commandLineService == nil {
		return "", nil, nil, fmt.Errorf("command line service config is nil")
	}

	info := &tool.ServiceInfo{
		Name:   serviceConfig.GetName(),
		Config: serviceConfig,
	}
	toolManager.AddServiceInfo(serviceID, info)

	discoveredTools, err := u.createAndRegisterCommandTools(
		ctx,
		serviceID,
		commandLineService,
		serviceConfig.GetCallPolicies(),
		toolManager,
		resourceManager,
		isReload,
	)
	if err != nil {
		return "", nil, nil, err
	}
	log.Info(
		"Registered command service",
		"serviceID",
		serviceID,
		"toolsAdded",
		len(discoveredTools),
	)

	u.createAndRegisterPrompts(ctx, serviceID, commandLineService, promptManager, isReload)

	return serviceID, discoveredTools, nil, nil
}

// createAndRegisterCommandTools iterates through the command definitions in the
// service configuration, creates a new CommandTool for each, and registers it
// with the tool manager.
func (u *Upstream) createAndRegisterCommandTools(
	_ context.Context,
	serviceID string,
	commandLineService *configv1.CommandLineUpstreamService,
	callPolicies []*configv1.CallPolicy,
	toolManager tool.ManagerInterface,
	resourceManager resource.ManagerInterface,
	_ bool,
) ([]*configv1.ToolDefinition, error) {
	log := logging.GetLogger()
	discoveredTools := make([]*configv1.ToolDefinition, 0, len(commandLineService.GetTools()))
	definitions := commandLineService.GetTools()
	calls := commandLineService.GetCalls()

	for _, definition := range definitions {
		if definition.GetDisable() {
			log.Info("Skipping disabled tool", "toolName", definition.GetName())
			continue
		}

		callID := definition.GetCallId()
		callDef, ok := calls[callID]
		if !ok {
			log.Error("Call definition not found for tool", "call_id", callID, "tool_name", definition.GetName())
			continue
		}

		command := definition.GetName()

		inputProperties, err := schemaconv.ConfigSchemaToProtoProperties(callDef.GetParameters())
		if err != nil {
			log.Error("Failed to convert config schema to proto properties", "error", err)
			continue
		}

		if inputProperties.Fields == nil {
			inputProperties.Fields = make(map[string]*structpb.Value)
		}

		inputSchema := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"type":       structpb.NewStringValue("object"),
				"properties": structpb.NewStructValue(inputProperties),
			},
		}

		outputProperties, err := structpb.NewStruct(map[string]interface{}{
			"command":         map[string]interface{}{"type": "string", "description": "The command that was executed."},
			"args":            map[string]interface{}{"type": "array", "description": "The arguments passed to the command."},
			"stdout":          map[string]interface{}{"type": "string", "description": "The standard output of the command."},
			"stderr":          map[string]interface{}{"type": "string", "description": "The standard error of the command."},
			"combined_output": map[string]interface{}{"type": "string", "description": "The combined standard output and standard error."},
			"start_time":      map[string]interface{}{"type": "string", "description": "The time the command started executing."},
			"end_time":        map[string]interface{}{"type": "string", "description": "The time the command finished executing."},
			"return_code":     map[string]interface{}{"type": "integer", "description": "The exit code of the command."},
			"status":          map[string]interface{}{"type": "string", "description": "The execution status of the command (e.g., success, error, timeout)."},
		})
		if err != nil {
			log.Error("Failed to create output properties", "error", err)
			continue
		}
		outputSchema := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"type":       structpb.NewStringValue("object"),
				"properties": structpb.NewStructValue(outputProperties),
			},
		}

		newToolProto := pb.Tool_builder{
			Name:                proto.String(command),
			DisplayName:         proto.String(command),
			Description:         proto.String(definition.GetDescription()),
			ServiceId:           proto.String(serviceID),
			UnderlyingMethodFqn: proto.String(command),
			InputSchema:         inputSchema,
			OutputSchema:        outputSchema,
		}.Build()

		var newTool tool.Tool
		if commandLineService.GetLocal() {
			newTool = tool.NewLocalCommandTool(newToolProto, commandLineService, callDef, callPolicies, callID)
		} else {
			newTool = tool.NewCommandTool(newToolProto, commandLineService, callDef, callPolicies, callID)
		}
		if err := toolManager.AddTool(newTool); err != nil {
			log.Error("Failed to add tool", "error", err)
			return nil, err
		}
		discoveredTools = append(discoveredTools, configv1.ToolDefinition_builder{
			Name: proto.String(command),
		}.Build())
	}

	callIDToName := make(map[string]string)
	for _, d := range definitions {
		callIDToName[d.GetCallId()] = d.GetName()
	}
	for _, resourceDef := range commandLineService.GetResources() {
		if resourceDef.GetDisable() {
			log.Info("Skipping disabled resource", "resourceName", resourceDef.GetName())
			continue
		}

		if resourceDef.GetDynamic() != nil {
			call := resourceDef.GetDynamic().GetCommandLineCall()
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
		} else if resourceDef.GetStatic() != nil {
			staticResource := resource.NewStaticResource(resourceDef, serviceID)
			resourceManager.AddResource(staticResource)
		}
	}

	return discoveredTools, nil
}

func (u *Upstream) createAndRegisterPrompts(
	_ context.Context,
	serviceID string,
	commandLineService *configv1.CommandLineUpstreamService,
	promptManager prompt.ManagerInterface,
	_ bool,
) {
	log := logging.GetLogger()
	for _, promptDef := range commandLineService.GetPrompts() {
		if promptDef.GetDisable() {
			log.Info("Skipping disabled prompt", "promptName", promptDef.GetName())
			continue
		}
		newPrompt := prompt.NewTemplatedPrompt(promptDef, serviceID)
		promptManager.AddPrompt(newPrompt)
		log.Info("Registered prompt", "prompt_name", newPrompt.Prompt().Name)
	}
}
