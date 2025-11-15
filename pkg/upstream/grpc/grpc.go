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

package grpc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jellydator/ttlcache/v3"
	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream"
	"github.com/mcpany/core/pkg/upstream/grpc/protobufparser"
	"github.com/mcpany/core/pkg/util"
	"github.com/mcpany/core/pkg/util/schemaconv"
	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/structpb"
)

// GRPCUpstream implements the upstream.Upstream interface for gRPC services.
// It uses gRPC reflection to discover services and methods, and creates tools
// for them. It also manages a connection pool and a cache for reflection data.
type GRPCUpstream struct {
	poolManager     *pool.Manager
	reflectionCache *ttlcache.Cache[string, *descriptorpb.FileDescriptorSet]
	toolManager     tool.ToolManagerInterface
}

// NewGRPCUpstream creates a new instance of GRPCUpstream.
//
// poolManager is the connection pool manager to be used for managing gRPC
// connections.
func NewGRPCUpstream(poolManager *pool.Manager) upstream.Upstream {
	cache := ttlcache.New[string, *descriptorpb.FileDescriptorSet](
		ttlcache.WithTTL[string, *descriptorpb.FileDescriptorSet](5 * time.Minute),
	)
	go cache.Start()

	return &GRPCUpstream{
		poolManager:     poolManager,
		reflectionCache: cache,
	}
}

// Register handles the registration of a gRPC upstream service. It establishes a
// connection pool, uses gRPC reflection to discover the service's protobuf
// definitions, and then creates and registers tools based on the discovered
// methods and any MCP annotations.
func (u *GRPCUpstream) Register(
	ctx context.Context,
	serviceConfig *configv1.UpstreamServiceConfig,
	serviceRegistry tool.ServiceRegistry,
	toolManager tool.ToolManagerInterface,
	promptManager prompt.PromptManagerInterface,
	resourceManager resource.ResourceManagerInterface,
	isReload bool,
) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	u.toolManager = toolManager
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

	grpcService := serviceConfig.GetGrpcService()
	if grpcService == nil {
		return "", nil, nil, fmt.Errorf("grpc service config is nil")
	}

	upstreamAuthenticator, err := auth.NewUpstreamAuthenticator(serviceConfig.GetUpstreamAuthentication())
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to create upstream authenticator for gRPC service %s: %w", serviceID, err)
	}
	grpcCreds := auth.NewPerRPCCredentials(upstreamAuthenticator)

	grpcPool, err := NewGrpcPool(0, 10, 300, nil, grpcCreds, serviceConfig, false)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to create gRPC pool for %s: %w", serviceConfig.GetName(), err)
	}
	u.poolManager.Register(serviceID, grpcPool)

	var fds *descriptorpb.FileDescriptorSet
	if grpcService.GetUseReflection() {
		item := u.reflectionCache.Get(grpcService.GetAddress())
		if item != nil {
			fds = item.Value()
		} else {
			var err error
			fds, err = protobufparser.ParseProtoByReflection(ctx, grpcService.GetAddress())
			if err != nil {
				return "", nil, nil, fmt.Errorf("failed to discover service by reflection for %s (target: %s): %w", serviceID, grpcService.GetAddress(), err)
			}
			u.reflectionCache.Set(grpcService.GetAddress(), fds, ttlcache.DefaultTTL)
		}
	} else {
		var err error
		fds, err = protobufparser.ParseProtoFromDefs(ctx, grpcService.GetProtoDefinitions(), grpcService.GetProtoCollection())
		if err != nil {
			return "", nil, nil, fmt.Errorf("failed to parse proto definitions for %s: %w", serviceID, err)
		}
	}

	serviceRegistry.AddServiceInfo(serviceID, &tool.ServiceInfo{
		Name:   serviceConfig.GetName(),
		Config: serviceConfig,
		Fds:    fds,
	})

	parsedMcpData, err := protobufparser.ExtractMcpDefinitions(fds)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to extract MCP definitions for %s: %w", serviceID, err)
	}

	discoveredTools, err := u.createAndRegisterGRPCTools(ctx, serviceID, parsedMcpData, serviceRegistry, toolManager, resourceManager, isReload, fds)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to create and register gRPC tools for %s: %w", serviceID, err)
	}

	discoveredToolsFromDescriptors, err := u.createAndRegisterGRPCToolsFromDescriptors(ctx, serviceID, serviceRegistry, toolManager, resourceManager, isReload, fds)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to create and register gRPC tools from descriptors for %s: %w", serviceID, err)
	}
	discoveredTools = append(discoveredTools, discoveredToolsFromDescriptors...)

	discoveredToolsFromConfig, err := u.createAndRegisterGRPCToolsFromConfig(ctx, serviceID, serviceRegistry, toolManager, resourceManager, isReload, fds)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to create and register gRPC tools from config for %s: %w", serviceID, err)
	}
	discoveredTools = append(discoveredTools, discoveredToolsFromConfig...)

	err = u.createAndRegisterPromptsFromConfig(ctx, serviceID, serviceRegistry, promptManager, isReload)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to create and register prompts from config for %s: %w", serviceID, err)
	}

	log.Info("Registered gRPC service", "serviceID", serviceID, "toolsAdded", len(discoveredTools))

	return serviceID, discoveredTools, nil, nil
}

// createAndRegisterGRPCTools iterates through the parsed MCP annotations, which
// contain tool definitions extracted from protobuf options. For each tool, it
// constructs a GRPCTool and registers it with the tool manager.
func (u *GRPCUpstream) createAndRegisterGRPCTools(
	ctx context.Context,
	serviceID string,
	parsedData *protobufparser.ParsedMcpAnnotations,
	serviceRegistry tool.ServiceRegistry,
	tm tool.ToolManagerInterface,
	resourceManager resource.ResourceManagerInterface,
	isReload bool,
	fds *descriptorpb.FileDescriptorSet,
) ([]*configv1.ToolDefinition, error) {
	log := logging.GetLogger()
	if parsedData == nil {
		return nil, nil
	}

	serviceInfo, ok := serviceRegistry.GetServiceInfo(serviceID)
	if !ok {
		return nil, fmt.Errorf("service info not found for service: %s", serviceID)
	}
	grpcService := serviceInfo.Config.GetGrpcService()

	files, err := protodesc.NewFiles(fds)
	if err != nil {
		return nil, fmt.Errorf("failed to create protodesc files: %w", err)
	}

	discoveredTools := make([]*configv1.ToolDefinition, 0, len(parsedData.Tools))
	for _, toolDef := range parsedData.Tools {
		toolName := toolDef.Name
		if toolName == "" {
			toolName = toolDef.MethodName
		}

		methodDescriptor, err := findMethodDescriptor(files, toolDef.FullMethodName)
		if err != nil {
			log.Error("Failed to find method descriptor, skipping tool.", "tool_name", toolDef.Name, "method_fqn", toolDef.FullMethodName, "error", err)
			continue
		}

		requestFields := make([]schemaconv.McpFieldParameter, len(toolDef.RequestFields))
		for i := range toolDef.RequestFields {
			requestFields[i] = &toolDef.RequestFields[i]
		}

		propertiesStruct, err := schemaconv.McpFieldsToProtoProperties(requestFields)
		if err != nil {
			log.Error("Failed to convert McpFields to InputSchema, skipping.", "tool_name", toolDef.Name, "error", err)
			continue
		}
		if propertiesStruct == nil {
			propertiesStruct = &structpb.Struct{Fields: make(map[string]*structpb.Value)}
		}
		inputSchema := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"type":       structpb.NewStringValue("object"),
				"properties": structpb.NewStructValue(propertiesStruct),
			},
		}

		responseFields := make([]schemaconv.McpFieldParameter, len(toolDef.ResponseFields))
		for i := range toolDef.ResponseFields {
			responseFields[i] = &toolDef.ResponseFields[i]
		}
		outputPropertiesStruct, err := schemaconv.McpFieldsToProtoProperties(responseFields)
		if err != nil {
			log.Error("Failed to convert McpFields to OutputSchema, skipping.", "tool_name", toolDef.Name, "error", err)
			continue
		}
		if outputPropertiesStruct == nil {
			outputPropertiesStruct = &structpb.Struct{Fields: make(map[string]*structpb.Value)}
		}
		outputSchema := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"type":       structpb.NewStringValue("object"),
				"properties": structpb.NewStructValue(outputPropertiesStruct),
			},
		}

		newToolProto := pb.Tool_builder{
			Name:                proto.String(toolName),
			DisplayName:         proto.String(toolDef.Name),
			Description:         proto.String(toolDef.Description),
			ServiceId:           proto.String(serviceID),
			UnderlyingMethodFqn: proto.String(string(methodDescriptor.FullName())),
			RequestTypeFqn:      proto.String(toolDef.RequestType),
			ResponseTypeFqn:     proto.String(toolDef.ResponseType),
			InputSchema:         inputSchema,
			OutputSchema:        outputSchema,
			Annotations: pb.ToolAnnotations_builder{
				Title:           proto.String(toolDef.Name),
				ReadOnlyHint:    proto.Bool(toolDef.ReadOnlyHint),
				DestructiveHint: proto.Bool(toolDef.DestructiveHint),
				IdempotentHint:  proto.Bool(toolDef.IdempotentHint),
				OpenWorldHint:   proto.Bool(toolDef.OpenWorldHint),
				InputSchema:     inputSchema,
				OutputSchema:    outputSchema,
			}.Build(),
		}.Build()

		clonedTool := proto.Clone(newToolProto).(*pb.Tool)
		grpcTool := tool.NewGRPCTool(clonedTool, u.poolManager, serviceID, methodDescriptor, nil)
		if err := tm.AddTool(grpcTool); err != nil {
			log.Error("Failed to add gRPC tool", "tool_name", toolName, "error", err)
			continue
		}
		discoveredTools = append(discoveredTools, configv1.ToolDefinition_builder{
			Name:         proto.String(toolDef.Name),
			Description:  proto.String(toolDef.Description),
			InputSchema:  inputSchema,
			OutputSchema: outputSchema,
		}.Build())
		log.Info("Registered gRPC tool", "tool_id", newToolProto.GetName(), "is_reload", isReload)
	}

	definitions := grpcService.GetTools()
	callIDToName := make(map[string]string)
	for _, d := range definitions {
		if d != nil {
			callIDToName[d.GetCallId()] = d.GetName()
		}
	}
	for _, resourceDef := range grpcService.GetResources() {
		if resourceDef.GetDynamic() != nil {
			call := resourceDef.GetDynamic().GetGrpcCall()
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
			tool, ok := tm.GetTool(serviceID + "." + sanitizedToolName)
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

	return discoveredTools, nil
}

func (u *GRPCUpstream) createAndRegisterGRPCToolsFromDescriptors(
	ctx context.Context,
	serviceID string,
	serviceRegistry tool.ServiceRegistry,
	tm tool.ToolManagerInterface,
	resourceManager resource.ResourceManagerInterface,
	isReload bool,
	fds *descriptorpb.FileDescriptorSet,
) ([]*configv1.ToolDefinition, error) {
	log := logging.GetLogger()
	if fds == nil {
		return nil, nil
	}

	serviceInfo, ok := serviceRegistry.GetServiceInfo(serviceID)
	if !ok {
		return nil, fmt.Errorf("service info not found for service: %s", serviceID)
	}
	grpcService := serviceInfo.Config.GetGrpcService()

	files, err := protodesc.NewFiles(fds)
	if err != nil {
		return nil, fmt.Errorf("failed to create protodesc files: %w", err)
	}

	discoveredTools := make([]*configv1.ToolDefinition, 0)
	files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		services := fd.Services()
		for i := 0; i < services.Len(); i++ {
			serviceDesc := services.Get(i)
			methods := serviceDesc.Methods()
			for j := 0; j < methods.Len(); j++ {
				methodDesc := methods.Get(j)

				// Check if the tool is already registered
				toolID := fmt.Sprintf("%s.%s", serviceID, methodDesc.Name())
				if _, ok := tm.GetTool(toolID); ok {
					continue
				}

				propertiesStruct, err := schemaconv.MethodDescriptorToProtoProperties(methodDesc)
				if err != nil {
					log.Error("Failed to convert MethodDescriptor to InputSchema, skipping.", "method", methodDesc.FullName(), "error", err)
					continue
				}

				if propertiesStruct == nil {
					propertiesStruct = &structpb.Struct{Fields: make(map[string]*structpb.Value)}
				}
				inputSchema := &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"type":       structpb.NewStringValue("object"),
						"properties": structpb.NewStructValue(propertiesStruct),
					},
				}

				outputPropertiesStruct, err := schemaconv.MethodOutputDescriptorToProtoProperties(methodDesc)
				if err != nil {
					log.Error("Failed to convert MethodDescriptor to OutputSchema, skipping.", "method", methodDesc.FullName(), "error", err)
					continue
				}

				if outputPropertiesStruct == nil {
					outputPropertiesStruct = &structpb.Struct{Fields: make(map[string]*structpb.Value)}
				}
				outputSchema := &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"type":       structpb.NewStringValue("object"),
						"properties": structpb.NewStructValue(outputPropertiesStruct),
					},
				}

				newToolProto := pb.Tool_builder{
					Name:                proto.String(string(methodDesc.Name())),
					DisplayName:         proto.String(string(methodDesc.Name())),
					Description:         proto.String(string(methodDesc.FullName())),
					ServiceId:           proto.String(serviceID),
					UnderlyingMethodFqn: proto.String(string(methodDesc.FullName())),
					RequestTypeFqn:      proto.String(string(methodDesc.Input().FullName())),
					ResponseTypeFqn:     proto.String(string(methodDesc.Output().FullName())),
					InputSchema:         inputSchema,
					OutputSchema:        outputSchema,
					Annotations: pb.ToolAnnotations_builder{
						Title:        proto.String(string(methodDesc.Name())),
						InputSchema:  inputSchema,
						OutputSchema: outputSchema,
					}.Build(),
				}.Build()

				clonedTool := proto.Clone(newToolProto).(*pb.Tool)
				grpcTool := tool.NewGRPCTool(clonedTool, u.poolManager, serviceID, methodDesc, nil)
				if err := tm.AddTool(grpcTool); err != nil {
					log.Error("Failed to add gRPC tool", "tool_name", methodDesc.Name(), "error", err)
					continue
				}
				discoveredTools = append(discoveredTools, configv1.ToolDefinition_builder{
					Name:         proto.String(string(methodDesc.Name())),
					Description:  proto.String(string(methodDesc.FullName())),
					InputSchema:  inputSchema,
					OutputSchema: outputSchema,
				}.Build())
				log.Info("Registered gRPC tool from descriptor", "tool_id", newToolProto.GetName(), "is_reload", isReload)
			}
		}
		return true
	})

	definitions := grpcService.GetTools()
	callIDToName := make(map[string]string)
	for _, d := range definitions {
		if d != nil {
			callIDToName[d.GetCallId()] = d.GetName()
		}
	}
	for _, resourceDef := range grpcService.GetResources() {
		if resourceDef.GetDynamic() != nil {
			call := resourceDef.GetDynamic().GetGrpcCall()
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
			tool, ok := tm.GetTool(serviceID + "." + sanitizedToolName)
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

	return discoveredTools, nil
}

// findMethodDescriptor locates a MethodDescriptor within a set of protobuf file
// descriptors, given a fully qualified method name (e.g.,
// "my.package.MyService/MyMethod").
func findMethodDescriptor(files *protoregistry.Files, fullMethodName string) (protoreflect.MethodDescriptor, error) {
	// Normalize the method name by removing any leading slash.
	normalizedMethodName := strings.TrimPrefix(fullMethodName, "/")
	lastSeparator := strings.LastIndex(normalizedMethodName, "/")
	if lastSeparator == -1 {
		lastSeparator = strings.LastIndex(normalizedMethodName, ".")
	}

	if lastSeparator == -1 {
		return nil, fmt.Errorf("invalid full method name: %s", fullMethodName)
	}

	serviceName := protoreflect.FullName(normalizedMethodName[:lastSeparator])
	methodName := protoreflect.Name(normalizedMethodName[lastSeparator+1:])

	desc, err := files.FindDescriptorByName(serviceName)
	if err != nil {
		return nil, fmt.Errorf("could not find descriptor for service '%s': %w", serviceName, err)
	}

	serviceDesc, ok := desc.(protoreflect.ServiceDescriptor)
	if !ok {
		return nil, fmt.Errorf("descriptor for '%s' is not a service descriptor", serviceName)
	}

	methodDesc := serviceDesc.Methods().ByName(methodName)
	if methodDesc == nil {
		return nil, fmt.Errorf("method '%s' not found in service '%s'", methodName, serviceName)
	}
	return methodDesc, nil
}

func (u *GRPCUpstream) createAndRegisterGRPCToolsFromConfig(
	ctx context.Context,
	serviceID string,
	serviceRegistry tool.ServiceRegistry,
	tm tool.ToolManagerInterface,
	resourceManager resource.ResourceManagerInterface,
	isReload bool,
	fds *descriptorpb.FileDescriptorSet,
) ([]*configv1.ToolDefinition, error) {
	log := logging.GetLogger()
	if fds == nil {
		return nil, nil
	}

	serviceInfo, ok := serviceRegistry.GetServiceInfo(serviceID)
	if !ok {
		return nil, fmt.Errorf("service info not found for service: %s", serviceID)
	}
	grpcService := serviceInfo.Config.GetGrpcService()

	files, err := protodesc.NewFiles(fds)
	if err != nil {
		return nil, fmt.Errorf("failed to create protodesc files: %w", err)
	}

	discoveredTools := make([]*configv1.ToolDefinition, 0, len(grpcService.GetTools()))
	definitions := grpcService.GetTools()
	calls := grpcService.GetCalls()

	for _, definition := range definitions {
		callID := definition.GetCallId()
		grpcDef, ok := calls[callID]
		if !ok {
			log.Error("Call definition not found for tool", "call_id", callID, "tool_name", definition.GetName())
			continue
		}

		fullMethodName := fmt.Sprintf("%s.%s", grpcDef.GetService(), grpcDef.GetMethod())
		methodDescriptor, err := findMethodDescriptor(files, fullMethodName)
		if err != nil {
			log.Error("Failed to find method descriptor, skipping tool.", "tool_name", definition.GetName(), "method_fqn", fullMethodName, "error", err)
			continue
		}

		inputSchema, err := schemaconv.MethodDescriptorToProtoProperties(methodDescriptor)
		if err != nil {
			log.Error("Failed to convert MethodDescriptor to InputSchema, skipping.", "method", methodDescriptor.FullName(), "error", err)
			continue
		}

		outputSchema, err := schemaconv.MethodOutputDescriptorToProtoProperties(methodDescriptor)
		if err != nil {
			log.Error("Failed to convert MethodDescriptor to OutputSchema, skipping.", "method", methodDescriptor.FullName(), "error", err)
			continue
		}

		newToolProto := pb.Tool_builder{
			Name:                proto.String(definition.GetName()),
			Description:         proto.String(definition.GetDescription()),
			ServiceId:           proto.String(serviceID),
			UnderlyingMethodFqn: proto.String(fullMethodName),
			Annotations: pb.ToolAnnotations_builder{
				Title:           proto.String(definition.GetTitle()),
				ReadOnlyHint:    proto.Bool(definition.GetReadOnlyHint()),
				DestructiveHint: proto.Bool(definition.GetDestructiveHint()),
				IdempotentHint:  proto.Bool(definition.GetIdempotentHint()),
				OpenWorldHint:   proto.Bool(definition.GetOpenWorldHint()),
				InputSchema:     inputSchema,
				OutputSchema:    outputSchema,
			}.Build(),
		}.Build()

		grpcTool := tool.NewGRPCTool(newToolProto, u.poolManager, serviceID, methodDescriptor, grpcDef)
		if err := tm.AddTool(grpcTool); err != nil {
			log.Error("Failed to add tool", "error", err)
			continue
		}
		discoveredTools = append(discoveredTools, configv1.ToolDefinition_builder{
			Name:         proto.String(definition.GetName()),
			Description:  proto.String(definition.GetDescription()),
			InputSchema:  inputSchema,
			OutputSchema: outputSchema,
		}.Build())
	}
	return discoveredTools, nil
}

func (u *GRPCUpstream) createAndRegisterPromptsFromConfig(
	ctx context.Context,
	serviceID string,
	serviceRegistry tool.ServiceRegistry,
	promptManager prompt.PromptManagerInterface,
	isReload bool,
) error {
	log := logging.GetLogger()
	serviceInfo, ok := serviceRegistry.GetServiceInfo(serviceID)
	if !ok {
		return fmt.Errorf("service info not found for service: %s", serviceID)
	}
	grpcService := serviceInfo.Config.GetGrpcService()

	for _, promptDef := range grpcService.GetPrompts() {
		newPrompt := prompt.NewTemplatedPrompt(promptDef, serviceID)
		promptManager.AddPrompt(newPrompt)
		log.Info("Registered prompt", "prompt_name", newPrompt.Prompt().Name, "is_reload", isReload)
	}

	return nil
}
