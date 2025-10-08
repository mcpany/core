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

package command

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/mcpxy/core/pkg/logging"
	"github.com/mcpxy/core/pkg/prompt"
	"github.com/mcpxy/core/pkg/resource"
	"github.com/mcpxy/core/pkg/tool"
	"github.com/mcpxy/core/pkg/upstream"
	"github.com/mcpxy/core/pkg/util"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	pb "github.com/mcpxy/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
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
) (string, []*configv1.ToolDefinition, error) {
	log := logging.GetLogger()
	serviceKey, err := util.GenerateServiceKey(serviceConfig.GetName())
	if err != nil {
		return "", nil, err
	}

	commandLineService := serviceConfig.GetCommandLineService()
	if commandLineService == nil {
		return "", nil, fmt.Errorf("command line service config is nil")
	}

	info := &tool.ServiceInfo{
		Name:   serviceConfig.GetName(),
		Config: serviceConfig,
	}
	toolManager.AddServiceInfo(serviceKey, info)

	discoveredTools := u.createAndRegisterCommandTools(ctx, serviceKey, commandLineService, toolManager, isReload)
	log.Info("Registered command service", "serviceKey", serviceKey, "toolsAdded", len(discoveredTools))

	return serviceKey, discoveredTools, nil
}

func (u *CommandUpstream) createAndRegisterCommandTools(ctx context.Context, serviceKey string, commandLineService *configv1.CommandLineUpstreamService, toolManager tool.ToolManagerInterface, isReload bool) []*configv1.ToolDefinition {
	log := logging.GetLogger()
	discoveredTools := make([]*configv1.ToolDefinition, 0, len(commandLineService.GetCalls()))

	if len(commandLineService.GetCalls()) == 0 {
		// If no calls are defined, create a single tool from the command itself.
		command := commandLineService.GetCommand()
		if command == "" {
			return discoveredTools
		}
		// Use the base name of the command as the tool name.
		toolName := filepath.Base(command)
		newToolProto := pb.Tool_builder{
			Name:                proto.String(toolName),
			DisplayName:         proto.String(toolName),
			Description:         proto.String(fmt.Sprintf("Executes the command: %s", command)),
			ServiceId:           proto.String(serviceKey),
			UnderlyingMethodFqn: proto.String(command),
		}.Build()

		newTool := tool.NewCommandTool(newToolProto, command)
		if err := toolManager.AddTool(newTool); err != nil {
			log.Error("Failed to add default tool for command service", "error", err)
		} else {
			discoveredTools = append(discoveredTools, configv1.ToolDefinition_builder{
				Name: proto.String(toolName),
			}.Build())
		}
	} else {
		for _, toolDef := range commandLineService.GetCalls() {
			command := toolDef.GetMethod()
			newToolProto := pb.Tool_builder{
				Name:                proto.String(command),
				DisplayName:         proto.String(command),
				Description:         proto.String(command),
				ServiceId:           proto.String(serviceKey),
				UnderlyingMethodFqn: proto.String(command),
			}.Build()

			newTool := tool.NewCommandTool(newToolProto, command)
			if err := toolManager.AddTool(newTool); err != nil {
				log.Error("Failed to add tool", "error", err)
				continue
			}
			discoveredTools = append(discoveredTools, configv1.ToolDefinition_builder{
				Name: proto.String(command),
			}.Build())
		}
	}

	return discoveredTools
}
