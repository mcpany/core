// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/metrics"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/resilience"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

// GRPCTool implements the Tool interface for a tool that is exposed via a gRPC
// endpoint.
//
// Summary: Tool implementation for gRPC services.
//
// It handles the marshalling of JSON inputs to protobuf messages and
// invoking the gRPC method.
type GRPCTool struct {
	tool              *v1.Tool
	mcpTool           *mcp.Tool
	mcpToolOnce       sync.Once
	poolManager       *pool.Manager
	serviceID         string
	method            protoreflect.MethodDescriptor
	requestMessage    protoreflect.ProtoMessage
	cache             *configv1.CacheConfig
	resilienceManager *resilience.Manager
}

// NewGRPCTool creates a new GRPCTool instance.
//
// Summary: Initializes a new GRPCTool.
//
// Parameters:
//   - tool: *v1.Tool. The protobuf definition of the tool.
//   - poolManager: *pool.Manager. The connection pool manager for gRPC connections.
//   - serviceID: string. The identifier for the service.
//   - method: protoreflect.MethodDescriptor. The gRPC method descriptor.
//   - callDefinition: *configv1.GrpcCallDefinition. The configuration for the gRPC call.
//   - resilienceConfig: *configv1.ResilienceConfig. The resilience configuration.
//
// Returns:
//   - *GRPCTool: The initialized GRPCTool.
func NewGRPCTool(tool *v1.Tool, poolManager *pool.Manager, serviceID string, method protoreflect.MethodDescriptor, callDefinition *configv1.GrpcCallDefinition, resilienceConfig *configv1.ResilienceConfig) *GRPCTool {
	return &GRPCTool{
		tool:              tool,
		poolManager:       poolManager,
		serviceID:         serviceID,
		method:            method,
		requestMessage:    dynamicpb.NewMessage(method.Input()),
		cache:             callDefinition.GetCache(),
		resilienceManager: resilience.NewManager(resilienceConfig),
	}
}

// Tool returns the protobuf definition of the gRPC tool.
//
// Returns:
//   - *v1.Tool: The underlying protobuf definition.
//
// Summary: Executes Tool operation.
//
// Parameters:
//   - TODO: Document parameters.
//
// Returns:
//   - TODO: Document returns.
//
// Errors:
//   - TODO: Document errors.
//
// Side Effects:
//   - None.
func (t *GRPCTool) Tool() *v1.Tool {
	return t.tool
}

// MCPTool returns the MCP-compliant tool definition.
//
// It lazily converts the internal protobuf definition to the MCP format on first access.
//
// Returns:
//   - *mcp.Tool: The MCP tool definition.
//
// Summary: Executes MCPTool operation.
//
// Parameters:
//   - TODO: Document parameters.
//
// Returns:
//   - TODO: Document returns.
//
// Errors:
//   - TODO: Document errors.
//
// Side Effects:
//   - None.
func (t *GRPCTool) MCPTool() *mcp.Tool {
	t.mcpToolOnce.Do(func() {
		var err error
		t.mcpTool, err = ConvertProtoToMCPTool(t.tool)
		if err != nil {
			logging.GetLogger().Error("Failed to convert tool to MCP tool", "toolName", t.tool.GetName(), "error", err)
		}
	})
	return t.mcpTool
}

// GetCacheConfig returns the cache configuration for the gRPC tool.
//
// Returns:
//   - *configv1.CacheConfig: The cache configuration, if any.
//
// Summary: Retrieves GetCacheConfig operation.
//
// Parameters:
//   - TODO: Document parameters.
//
// Returns:
//   - TODO: Document returns.
//
// Errors:
//   - TODO: Document errors.
//
// Side Effects:
//   - None.
func (t *GRPCTool) GetCacheConfig() *configv1.CacheConfig {
	return t.cache
}

// Execute handles the execution of the gRPC tool.
//
// Summary: Executes the gRPC tool call.
//
// It retrieves a client from the pool, unmarshals the JSON input into a protobuf request message,
// invokes the gRPC method, and marshals the protobuf response back to JSON.
//
// Parameters:
//   - ctx: context.Context. The execution context.
//   - req: *ExecutionRequest. The execution request.
//
// Returns:
//   - any: The execution result (usually a map or JSON string).
//   - error: An error if execution fails.
//
// Side Effects:
//   - Makes a gRPC call to the upstream service.
//   - Updates metrics (latency, success/error counts).
//   - Logs execution details.
//
// IsStreaming returns true if the tool supports streaming.
//
// Summary: Checks if the tool supports streaming execution.
//
// Returns:
//   - bool: True if streaming is supported.
func (t *GRPCTool) IsStreaming() bool {
	return false
}

// StreamExecute executes the tool in streaming mode.
//
// Summary: Executes the tool in streaming mode.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - req: *ExecutionRequest. The request object containing parameters.
//
// Returns:
//   - <-chan any: A channel that emits streaming results.
//   - error: An error if the operation fails or streaming is not supported.
func (t *GRPCTool) StreamExecute(ctx context.Context, req *ExecutionRequest) (<-chan any, error) {
	ch := make(chan any, 1)
	go func() {
		defer close(ch)
		res, err := t.Execute(ctx, req)
		if err != nil {
			ch <- err
		} else {
			ch <- res
		}
	}()
	return ch, nil
}

// Execute handles the execution of the gRPC tool.
//
// Summary: Executes the gRPC tool call.
//
// Parameters:
//   - ctx: context.Context. The execution context.
//   - req: *ExecutionRequest. The execution request.
//
// Returns:
//   - any: The execution result.
//   - error: An error if execution fails.
//
// Errors:
//   - Returns an error if the grpc pool is not found.
//   - Returns an error if getting a client from the pool fails.
//   - Returns an error if unmarshalling the tool inputs fails.
//   - Returns an error if the grpc method invocation fails.
//
// Side Effects:
//   - Makes a gRPC call to the upstream service.
func (t *GRPCTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	if logging.GetLogger().Enabled(ctx, slog.LevelDebug) {
		logging.GetLogger().Debug("executing tool", "tool", req.ToolName, "inputs", prettyPrint(req.ToolInputs, contentTypeJSON))
	}
	defer metrics.MeasureSince(metricGrpcRequestLatency, time.Now())
	grpcPool, ok := pool.Get[*client.GrpcClientWrapper](t.poolManager, t.serviceID)
	if !ok {
		metrics.IncrCounter(metricGrpcRequestError, 1)
		return nil, fmt.Errorf("no grpc pool found for service: %s", t.serviceID)
	}

	grpcClient, err := grpcPool.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get client from pool: %w", err)
	}
	defer grpcPool.Put(grpcClient)

	if err := protojson.Unmarshal(req.ToolInputs, t.requestMessage); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool inputs to protobuf: %w", err)
	}

	responseMessage := dynamicpb.NewMessage(t.method.Output())
	fqn := t.tool.GetUnderlyingMethodFqn()
	lastDot := strings.LastIndex(fqn, ".")
	if lastDot == -1 {
		return nil, fmt.Errorf("invalid method FQN: %s", fqn)
	}
	serviceName := fqn[:lastDot]
	methodName := fqn[lastDot+1:]
	grpcMethodName := fmt.Sprintf("/%s/%s", serviceName, methodName)

	if req.DryRun {
		logging.GetLogger().Info("Dry run execution", "tool", req.ToolName)
		jsonBytes, _ := protojson.Marshal(t.requestMessage)
		var payloadMap map[string]any
		_ = fastJSON.Unmarshal(jsonBytes, &payloadMap)
		return map[string]any{
			"dry_run": true,
			"request": map[string]any{
				"method":  grpcMethodName,
				"payload": payloadMap,
			},
		}, nil
	}

	work := func(ctx context.Context) error {
		return grpcClient.Invoke(ctx, grpcMethodName, t.requestMessage, responseMessage)
	}

	if err := t.resilienceManager.Execute(ctx, work); err != nil {
		metrics.IncrCounter(metricGrpcRequestError, 1)
		return nil, fmt.Errorf("failed to invoke grpc method: %w", err)
	}
	metrics.IncrCounter(metricGrpcRequestSuccess, 1)

	responseJSON, err := protojson.Marshal(responseMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal grpc response to json: %w", err)
	}

	// ⚡ Bolt: Use json-iterator
	var result map[string]any
	if err := fastJSON.Unmarshal(responseJSON, &result); err != nil {
		return string(responseJSON), nil
	}

	return result, nil
}
