// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/alexliesenfeld/health"
	"github.com/jellydator/ttlcache/v3"
	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/auth"
	mcphealth "github.com/mcpany/core/server/pkg/health"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream"
	"github.com/mcpany/core/server/pkg/upstream/grpc/protobufparser"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/pkg/util/schemaconv"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/structpb"
)

// Upstream implements the upstream.Upstream interface for gRPC services.
//
// Summary: implements the upstream.Upstream interface for gRPC services.
type Upstream struct {
	poolManager     *pool.Manager
	reflectionCache *ttlcache.Cache[string, *descriptorpb.FileDescriptorSet]
	toolManager     tool.ManagerInterface
	serviceID       string
	checker         health.Checker
}

// CheckHealth performs a health check on the upstream service.
//
// Summary: performs a health check on the upstream service.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (u *Upstream) CheckHealth(ctx context.Context) error {
	if u.checker != nil {
		res := u.checker.Check(ctx)
		if res.Status != health.StatusUp {
			return fmt.Errorf("health check failed: %v", res)
		}
		return nil
	}
	return nil
}

// NewUpstream creates a new instance of Upstream.
//
// Summary: creates a new instance of Upstream.
//
// Parameters:
//   - poolManager: *pool.Manager. The poolManager.
//
// Returns:
//   - upstream.Upstream: The upstream.Upstream.
func NewUpstream(poolManager *pool.Manager) upstream.Upstream {
	cache := ttlcache.New[string, *descriptorpb.FileDescriptorSet](
		ttlcache.WithTTL[string, *descriptorpb.FileDescriptorSet](5 * time.Minute),
	)
	go cache.Start()

	return &Upstream{
		poolManager:     poolManager,
		reflectionCache: cache,
	}
}

// Shutdown gracefully terminates the gRPC upstream service by shutting down the.
//
// Summary: gracefully terminates the gRPC upstream service by shutting down the.
//
// Parameters:
//   - _: context.Context. The _.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (u *Upstream) Shutdown(_ context.Context) error {
	u.reflectionCache.Stop()
	u.poolManager.Deregister(u.serviceID)
	return nil
}

// Register handles the registration of a gRPC upstream service. It establishes a.
//
// Summary: handles the registration of a gRPC upstream service. It establishes a.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - serviceConfig: *configv1.UpstreamServiceConfig. The serviceConfig.
//   - toolManager: tool.ManagerInterface. The toolManager.
//   - promptManager: prompt.ManagerInterface. The promptManager.
//   - resourceManager: resource.ManagerInterface. The resourceManager.
//   - isReload: bool. The isReload.
//
// Returns:
//   - string: The string.
//   - []*configv1.ToolDefinition: The []*configv1.ToolDefinition.
//   - []*configv1.ResourceDefinition: The []*configv1.ResourceDefinition.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (u *Upstream) Register(
	ctx context.Context,
	serviceConfig *configv1.UpstreamServiceConfig,
	toolManager tool.ManagerInterface,
	promptManager prompt.ManagerInterface,
	resourceManager resource.ManagerInterface,
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

	u.serviceID = sanitizedName // for internal use
	serviceID := u.serviceID

	u.checker = mcphealth.NewChecker(serviceConfig)

	grpcService := serviceConfig.GetGrpcService()
	if grpcService == nil {
		return "", nil, nil, fmt.Errorf("grpc service config is nil")
	}

	upstreamAuthenticator, err := auth.NewUpstreamAuthenticator(serviceConfig.GetUpstreamAuth())
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to create upstream authenticator for gRPC service %s: %w", serviceID, err)
	}
	grpcCreds := auth.NewPerRPCCredentials(upstreamAuthenticator)

	grpcPool, err := NewGrpcPool(0, 10, 300*time.Second, nil, grpcCreds, serviceConfig, false)
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

	toolManager.AddServiceInfo(serviceID, &tool.ServiceInfo{
		Name:   serviceConfig.GetName(),
		Config: serviceConfig,
		Fds:    fds,
	})

	parsedMcpData, err := protobufparser.ExtractMcpDefinitions(fds)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to extract MCP definitions for %s: %w", serviceID, err)
	}

	discoveredTools, err := u.createAndRegisterGRPCTools(ctx, serviceID, parsedMcpData, toolManager, resourceManager, isReload, fds)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to create and register gRPC tools for %s: %w", serviceID, err)
	}

	var discoveredToolsFromDescriptors []*configv1.ToolDefinition
	if serviceConfig.GetAutoDiscoverTool() {
		var err error
		discoveredToolsFromDescriptors, err = u.createAndRegisterGRPCToolsFromDescriptors(ctx, serviceID, toolManager, resourceManager, isReload, fds)
		if err != nil {
			return "", nil, nil, fmt.Errorf("failed to create and register gRPC tools from descriptors for %s: %w", serviceID, err)
		}
		discoveredTools = append(discoveredTools, discoveredToolsFromDescriptors...)
	} else if serviceConfig.GetGrpcService().GetUseReflection() {
		// Log that we are skipping auto-discovery despite reflection being on, if that's worth noting?
		// "auto_discover_tool" is new.
		// For backward compatibility, if "auto_discover_tool" is false, we might STOP doing what we were doing?
		// User instructions: "Use auto_discover_tool to UpstreamServiceConfig".
		// I'll stick to the flag.
		logging.GetLogger().Debug("Auto-discovery disabled for gRPC service with reflection enabled", "serviceID", serviceID)
	}

	discoveredToolsFromConfig, err := u.createAndRegisterGRPCToolsFromConfig(ctx, serviceID, toolManager, resourceManager, isReload, fds)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to create and register gRPC tools from config for %s: %w", serviceID, err)
	}
	discoveredTools = append(discoveredTools, discoveredToolsFromConfig...)

	err = u.createAndRegisterPromptsFromConfig(ctx, serviceID, promptManager, isReload)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to create and register prompts from config for %s: %w", serviceID, err)
	}

	log.Info("Registered gRPC service", "serviceID", serviceID, "toolsAdded", len(discoveredTools))

	return serviceID, discoveredTools, nil, nil
}

// createAndRegisterGRPCTools iterates through the parsed MCP annotations, which
// contain tool definitions extracted from protobuf options. For each tool, it
// constructs a GRPCTool and registers it with the tool manager.
//
func (u *Upstream) createAndRegisterGRPCTools(
	_ context.Context,
	serviceID string,
	parsedData *protobufparser.ParsedMcpAnnotations,
	tm tool.ManagerInterface,
	resourceManager resource.ManagerInterface,
	isReload bool,
	fds *descriptorpb.FileDescriptorSet,
) ([]*configv1.ToolDefinition, error) {
	log := logging.GetLogger()
	if parsedData == nil {
		return nil, nil
	}

	serviceInfo, ok := tm.GetServiceInfo(serviceID)
	if !ok {
		return nil, fmt.Errorf("service info not found for service: %s", serviceID)
	}
	grpcService := serviceInfo.Config.GetGrpcService()

	disabledTools := make(map[string]bool)
	for _, t := range grpcService.GetTools() {
		if t.GetDisable() {
			disabledTools[t.GetName()] = true
		}
	}

	files, err := protodesc.NewFiles(fds)
	if err != nil {
		return nil, fmt.Errorf("failed to create protodesc files: %w", err)
	}

	discoveredTools := make([]*configv1.ToolDefinition, 0, len(parsedData.Tools))
	for _, toolDef := range parsedData.Tools {
		if disabledTools[toolDef.Name] {
			log.Info("Skipping disabled tool (annotation)", "toolName", toolDef.Name)
			continue
		}
		// Check Export Policy
		serviceInfo, _ := tm.GetServiceInfo(serviceID)
		if serviceInfo != nil && !tool.ShouldExport(toolDef.Name, serviceInfo.Config.GetToolExportPolicy()) {
			log.Info("Skipping non-exported tool (annotation)", "toolName", toolDef.Name)
			continue
		}

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

		// Check if the tool is already registered
		sanitizedName, err := util.SanitizeToolName(toolName)
		if err != nil {
			log.Error("Failed to sanitize tool name for duplicate check", "name", toolName, "error", err)
			continue
		}
		toolID := fmt.Sprintf("%s.%s", serviceID, sanitizedName)
		if _, ok := tm.GetTool(toolID); ok {
			continue
		}

		clonedTool := proto.Clone(newToolProto).(*pb.Tool)
		grpcTool := tool.NewGRPCTool(clonedTool, u.poolManager, serviceID, methodDescriptor, nil, serviceInfo.Config.GetResilience())
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

	u.registerDynamicResources(serviceID, grpcService, resourceManager, tm)

	return discoveredTools, nil
}

// Deprecated: Logic is redundant with createAndRegisterGRPCTools (which uses annotations) or createAndRegisterGRPCToolsFromConfig (which uses config).
// Keeping this stub to satisfy existing callers if any, but it returns empty.
// If you need reflection-based discovery, use createAndRegisterGRPCTools.
func (u *Upstream) createAndRegisterGRPCToolsFromDescriptors(
	_ context.Context,
	_ string,
	_ tool.ManagerInterface,
	_ resource.ManagerInterface,
	_ bool,
	_ *descriptorpb.FileDescriptorSet,
) ([]*configv1.ToolDefinition, error) {
	return nil, nil
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

func (u *Upstream) createAndRegisterGRPCToolsFromConfig(
	_ context.Context,
	serviceID string,
	tm tool.ManagerInterface,
	_ resource.ManagerInterface,
	_ bool,
	fds *descriptorpb.FileDescriptorSet,
) ([]*configv1.ToolDefinition, error) {
	log := logging.GetLogger()
	if fds == nil {
		return nil, nil
	}

	serviceInfo, ok := tm.GetServiceInfo(serviceID)
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
		if definition.GetDisable() {
			log.Info("Skipping disabled tool (config)", "toolName", definition.GetName())
			continue
		}
		// Check Export Policy
		serviceInfo, _ := tm.GetServiceInfo(serviceID)
		if serviceInfo != nil && !tool.ShouldExport(definition.GetName(), serviceInfo.Config.GetToolExportPolicy()) {
			log.Info("Skipping non-exported tool (config)", "toolName", definition.GetName())
			continue
		}
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

		propertiesStruct, err := schemaconv.MethodDescriptorToProtoProperties(methodDescriptor)
		if err != nil {
			log.Error("Failed to convert MethodDescriptor to InputSchema, skipping.", "method", methodDescriptor.FullName(), "error", err)
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

		outputPropertiesStruct, err := schemaconv.MethodOutputDescriptorToProtoProperties(methodDescriptor)
		if err != nil {
			log.Error("Failed to convert MethodDescriptor to OutputSchema, skipping.", "method", methodDescriptor.FullName(), "error", err)
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

		grpcTool := tool.NewGRPCTool(newToolProto, u.poolManager, serviceID, methodDescriptor, grpcDef, serviceInfo.Config.GetResilience())
		if err := tm.AddTool(grpcTool); err != nil {
			log.Error("Failed to add tool", "error", err)
			continue
		}
		discoveredTools = append(discoveredTools, configv1.ToolDefinition_builder{
			Name:        proto.String(definition.GetName()),
			Description: proto.String(definition.GetDescription()),
		}.Build())
	}
	return discoveredTools, nil
}

func (u *Upstream) createAndRegisterPromptsFromConfig(
	_ context.Context,
	serviceID string,
	promptManager prompt.ManagerInterface,
	isReload bool,
) error {
	log := logging.GetLogger()
	serviceInfo, ok := u.toolManager.GetServiceInfo(serviceID)
	if !ok {
		return fmt.Errorf("service info not found for service: %s", serviceID)
	}
	grpcService := serviceInfo.Config.GetGrpcService()

	for _, promptDef := range grpcService.GetPrompts() {
		if promptDef.GetName() == "" {
			log.Error("Skipping prompt with missing name")
			continue
		}
		if promptDef.GetDisable() {
			log.Info("Skipping disabled prompt", "promptName", promptDef.GetName())
			continue
		}
		// Check Export Policy
		if !tool.ShouldExport(promptDef.GetName(), serviceInfo.Config.GetPromptExportPolicy()) {
			continue
		}
		newPrompt := prompt.NewTemplatedPrompt(promptDef, serviceID)
		promptManager.AddPrompt(newPrompt)
		log.Info("Registered prompt", "prompt_name", newPrompt.Prompt().Name, "is_reload", isReload)
	}

	return nil
}
func (u *Upstream) registerDynamicResources(
	serviceID string,
	grpcService *configv1.GrpcUpstreamService,
	rm resource.ManagerInterface,
	tm tool.ManagerInterface,
) {
	log := logging.GetLogger()
	definitions := grpcService.GetTools()
	callIDToName := make(map[string]string)
	for _, d := range definitions {
		if d != nil {
			callIDToName[d.GetCallId()] = d.GetName()
		}
	}
	for _, resourceDef := range grpcService.GetResources() {
		if resourceDef.GetDisable() {
			log.Info("Skipping disabled resource", "resourceName", resourceDef.GetName())
			continue
		}
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
			rm.AddResource(dynamicResource)
		}
	}
}
