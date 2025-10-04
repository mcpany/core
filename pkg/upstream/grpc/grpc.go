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

package grpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jellydator/ttlcache/v3"
	"github.com/mcpxy/core/pkg/auth"
	"github.com/mcpxy/core/pkg/logging"
	"github.com/mcpxy/core/pkg/pool"
	"github.com/mcpxy/core/pkg/prompt"
	"github.com/mcpxy/core/pkg/resource"
	"github.com/mcpxy/core/pkg/tool"
	"github.com/mcpxy/core/pkg/upstream"
	"github.com/mcpxy/core/pkg/upstream/grpc/protobufparser"
	"github.com/mcpxy/core/pkg/util"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	pb "github.com/mcpxy/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/encoding/protojson"
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

	grpcService := serviceConfig.GetGrpcService()
	if grpcService == nil {
		return "", nil, fmt.Errorf("grpc service config is nil")
	}

	address := grpcService.GetAddress()

	upstreamAuthenticator, err := auth.NewUpstreamAuthenticator(serviceConfig.GetUpstreamAuthentication())
	if err != nil {
		return "", nil, fmt.Errorf("failed to create upstream authenticator for gRPC service %s: %w", serviceKey, err)
	}
	grpcCreds := auth.NewPerRPCCredentials(upstreamAuthenticator)

	grpcPool, err := NewGrpcPool(0, 10, 300, address, nil, grpcCreds)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create gRPC pool for %s: %w", address, err)
	}
	u.poolManager.Register(serviceKey, grpcPool)

	item := u.reflectionCache.Get(address)
	var fds *descriptorpb.FileDescriptorSet
	if item != nil {
		fds = item.Value()
	} else {
		var err error
		fds, err = protobufparser.ParseProtoByReflection(ctx, address)
		if err != nil {
			return "", nil, fmt.Errorf("failed to discover service by reflection for %s (target: %s): %w", serviceKey, address, err)
		}
		u.reflectionCache.Set(address, fds, ttlcache.DefaultTTL)
	}

	toolManager.AddServiceInfo(serviceKey, &tool.ServiceInfo{
		Name:   serviceConfig.GetName(),
		Config: serviceConfig,
		Fds:    fds,
	})

	parsedMcpData, err := protobufparser.ExtractMcpDefinitions(fds)
	if err != nil {
		return "", nil, fmt.Errorf("failed to extract MCP definitions for %s: %w", serviceKey, err)
	}

	discoveredTools, err := u.createAndRegisterGRPCTools(ctx, serviceKey, parsedMcpData, toolManager, isReload, fds)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create and register gRPC tools for %s: %w", serviceKey, err)
	}
	log.Info("Registered gRPC service", "serviceKey", serviceKey, "toolsAdded", len(discoveredTools))

	return serviceKey, discoveredTools, nil
}

func (u *GRPCUpstream) createAndRegisterGRPCTools(
	ctx context.Context,
	serviceKey string,
	parsedData *protobufparser.ParsedMcpAnnotations,
	tm tool.ToolManagerInterface,
	isReload bool,
	fds *descriptorpb.FileDescriptorSet,
) ([]*configv1.ToolDefinition, error) {
	log := logging.GetLogger()
	if parsedData == nil {
		return nil, nil
	}

	_, ok := tm.GetServiceInfo(serviceKey)
	if !ok {
		return nil, fmt.Errorf("service info not found for service: %s", serviceKey)
	}

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

		requestFieldsPtrs := make([]*protobufparser.McpField, len(toolDef.RequestFields))
		for i := range toolDef.RequestFields {
			requestFieldsPtrs[i] = &toolDef.RequestFields[i]
		}

		inputSchemaProps, err := convertMcpFieldsToInputSchemaProperties(requestFieldsPtrs)
		if err != nil {
			log.Error("Failed to convert McpFields to InputSchema, skipping.", "tool_name", toolDef.Name, "error", err)
			continue
		}
		propertiesStruct := &structpb.Struct{Fields: inputSchemaProps}
		inputSchema := pb.InputSchema_builder{
			Type:       proto.String("object"),
			Properties: propertiesStruct,
		}.Build()

		newToolProto := pb.Tool_builder{
			Name:                proto.String(toolName),
			DisplayName:         proto.String(toolDef.Name),
			Description:         proto.String(toolDef.Description),
			ServiceId:           proto.String(serviceKey),
			UnderlyingMethodFqn: proto.String(string(methodDescriptor.FullName())),
			RequestTypeFqn:      proto.String(toolDef.RequestType),
			ResponseTypeFqn:     proto.String(toolDef.ResponseType),
			Annotations: pb.ToolAnnotations_builder{
				Title:           proto.String(toolDef.Name),
				ReadOnlyHint:    proto.Bool(toolDef.ReadOnlyHint),
				DestructiveHint: proto.Bool(toolDef.DestructiveHint),
				IdempotentHint:  proto.Bool(toolDef.IdempotentHint),
				OpenWorldHint:   proto.Bool(toolDef.OpenWorldHint),
			}.Build(),
			InputSchema: inputSchema,
		}.Build()

		clonedTool := proto.Clone(newToolProto).(*pb.Tool)
		grpcTool := tool.NewGRPCTool(clonedTool, u.poolManager, serviceKey, methodDescriptor)
		if err := tm.AddTool(grpcTool); err != nil {
			log.Error("Failed to add gRPC tool", "tool_name", toolName, "error", err)
			continue
		}
		discoveredTools = append(discoveredTools, configv1.ToolDefinition_builder{
			Name:        proto.String(toolDef.Name),
			Description: proto.String(toolDef.Description),
		}.Build())
		log.Info("Registered gRPC tool", "tool_id", newToolProto.GetName(), "is_reload", isReload)
	}

	return discoveredTools, nil
}

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

func convertMcpFieldsToInputSchemaProperties(fields []*protobufparser.McpField) (map[string]*structpb.Value, error) {
	properties := make(map[string]*structpb.Value)
	for _, field := range fields {
		schema := map[string]interface{}{
			"type":        "string", // Default to string for simplicity
			"description": field.Description,
		}

		scalarType := strings.ToLower(strings.TrimPrefix(field.Type, "TYPE_"))
		switch scalarType {
		case "double", "float":
			schema["type"] = "number"
		case "int32", "int64", "sint32", "sint64", "uint32", "uint64", "fixed32", "fixed64", "sfixed32", "sfixed64":
			schema["type"] = "integer"
		case "bool":
			schema["type"] = "boolean"
		}

		jsonBytes, err := json.Marshal(schema)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal schema to json: %w", err)
		}

		value := &structpb.Value{}
		if err := protojson.Unmarshal(jsonBytes, value); err != nil {
			return nil, fmt.Errorf("failed to unmarshal schema from json: %w", err)
		}
		properties[field.Name] = value
	}
	return properties, nil
}
