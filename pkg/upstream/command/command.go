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

// CommandUpstream implements the upstream.Upstream interface for services that
// are exposed as command-line tools. It discovers and registers tools based on
// a list of commands defined in the service configuration.
type CommandUpstream struct{}

// NewCommandUpstream creates a new instance of CommandUpstream.
func NewCommandUpstream() upstream.Upstream {
	return &CommandUpstream{}
}

// Register processes the configuration for a command-line service, creates a
// new tool for each defined command, and registers them with the tool manager.
func (u *CommandUpstream) Register(
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
		toolManager,
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

	return serviceID, discoveredTools, nil, nil
}

// createAndRegisterCommandTools iterates through the command definitions in the
// service configuration, creates a new CommandTool for each, and registers it
// with the tool manager.
func (u *CommandUpstream) createAndRegisterCommandTools(
	_ context.Context,
	serviceID string,
	commandLineService *configv1.CommandLineUpstreamService,
	toolManager tool.ToolManagerInterface,
	_ bool,
) ([]*configv1.ToolDefinition, error) {
	log := logging.GetLogger()
	discoveredTools := make([]*configv1.ToolDefinition, 0, len(commandLineService.GetCalls()))

	for _, toolDef := range commandLineService.GetCalls() {
		schema := toolDef.GetSchema()
		command := schema.GetName()

		inputProperties, err := schemaconv.ConfigSchemaToProtoProperties(toolDef.GetParameters())
		if err != nil {
			log.Error("Failed to convert config schema to proto properties", "error", err)
			continue
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
			Description:         proto.String(schema.GetDescription()),
			ServiceId:           proto.String(serviceID),
			UnderlyingMethodFqn: proto.String(command),
			InputSchema:         inputSchema,
			OutputSchema:        outputSchema,
		}.Build()

		newTool := tool.NewCommandTool(newToolProto, commandLineService.GetCommand(), toolDef)
		if err := toolManager.AddTool(newTool); err != nil {
			log.Error("Failed to add tool", "error", err)
			return nil, err
		}
		discoveredTools = append(discoveredTools, configv1.ToolDefinition_builder{
			Name: proto.String(command),
		}.Build())
	}

	return discoveredTools, nil
}
