// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/transformer"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// WebsocketTool implements the Tool interface for a tool exposed via a WebSocket
// connection. It handles sending and receiving messages over a persistent
// WebSocket connection managed by a connection pool.
type WebsocketTool struct {
	tool              *v1.Tool
	mcpTool           *mcp.Tool
	mcpToolOnce       sync.Once
	poolManager       *pool.Manager
	serviceID         string
	authenticator     auth.UpstreamAuthenticator
	parameters        []*configv1.WebsocketParameterMapping
	inputTransformer  *configv1.InputTransformer
	outputTransformer *configv1.OutputTransformer
	cache             *configv1.CacheConfig
}

// NewWebsocketTool creates a new WebsocketTool.
//
// tool is the protobuf definition of the tool.
// poolManager is used to get a WebSocket client from the connection pool.
// serviceID identifies the specific WebSocket service connection pool.
// authenticator handles adding authentication credentials to the connection request.
// callDefinition contains the configuration for the WebSocket call, such as
// parameter mappings and transformers.
func NewWebsocketTool(
	tool *v1.Tool,
	poolManager *pool.Manager,
	serviceID string,
	authenticator auth.UpstreamAuthenticator,
	callDefinition *configv1.WebsocketCallDefinition,
) *WebsocketTool {
	return &WebsocketTool{
		tool:              tool,
		poolManager:       poolManager,
		serviceID:         serviceID,
		authenticator:     authenticator,
		parameters:        callDefinition.GetParameters(),
		inputTransformer:  callDefinition.GetInputTransformer(),
		outputTransformer: callDefinition.GetOutputTransformer(),
		cache:             callDefinition.GetCache(),
	}
}

// Tool returns the protobuf definition of the WebSocket tool.
//
// Summary: Retrieves the protobuf definition.
//
// Returns:
//   - *v1.Tool: The protobuf tool definition.
func (t *WebsocketTool) Tool() *v1.Tool {
	return t.tool
}

// MCPTool returns the MCP tool definition.
//
// Summary: Retrieves the MCP-compliant tool definition.
//
// Returns:
//   - *mcp.Tool: The MCP tool definition.
//
// Side Effects:
//   - Lazily converts the proto definition to MCP format on first call.
func (t *WebsocketTool) MCPTool() *mcp.Tool {
	t.mcpToolOnce.Do(func() {
		var err error
		t.mcpTool, err = ConvertProtoToMCPTool(t.tool)
		if err != nil {
			logging.GetLogger().Error("Failed to convert tool to MCP tool", "toolName", t.tool.GetName(), "error", err)
		}
	})
	return t.mcpTool
}

// GetCacheConfig returns the cache configuration for the WebSocket tool.
//
// Summary: Retrieves the cache configuration.
//
// Returns:
//   - *configv1.CacheConfig: The cache configuration.
func (t *WebsocketTool) GetCacheConfig() *configv1.CacheConfig {
	return t.cache
}

// Execute handles the execution of the WebSocket tool.
//
// Summary: Executes the tool via a WebSocket connection.
//
// Description:
// It retrieves a connection from the pool, sends the tool inputs as a message,
// and waits for a single response message, which it then processes and returns.
//
// Parameters:
//   - ctx: context.Context. The context for the execution.
//   - req: *ExecutionRequest. The request containing inputs.
//
// Returns:
//   - any: The result of the execution.
//   - error: An error if connection retrieval, sending, or receiving fails.
//
// Side Effects:
//   - Sends a message over a WebSocket connection.
//   - Reads a message from a WebSocket connection.
func (t *WebsocketTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	wsPool, ok := pool.Get[*client.WebsocketClientWrapper](t.poolManager, t.serviceID)
	if !ok {
		return nil, fmt.Errorf("no websocket pool found for service: %s", t.serviceID)
	}

	wrapper, err := wsPool.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get websocket connection from pool: %w", err)
	}
	defer wsPool.Put(wrapper)

	_ = t.authenticator

	var inputs map[string]any
	if err := json.Unmarshal(req.ToolInputs, &inputs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool inputs: %w", err)
	}

	for _, param := range t.parameters {
		if secret := param.GetSecret(); secret != nil {
			secretValue, err := util.ResolveSecret(ctx, secret)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve secret for parameter %q: %w", param.GetSchema().GetName(), err)
			}
			inputs[param.GetSchema().GetName()] = secretValue
		}
	}

	var message []byte
	if t.inputTransformer != nil && t.inputTransformer.GetTemplate() != "" { //nolint:staticcheck
		tpl, err := transformer.NewTemplate(t.inputTransformer.GetTemplate(), "{{", "}}") //nolint:staticcheck
		if err != nil {
			return nil, fmt.Errorf("failed to create input template: %w", err)
		}
		rendered, err := tpl.Render(inputs)
		if err != nil {
			return nil, fmt.Errorf("failed to render input template: %w", err)
		}
		message = []byte(rendered)
	} else {
		message, err = json.Marshal(inputs)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal inputs to json: %w", err)
		}
	}

	if err := wrapper.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
		return nil, fmt.Errorf("failed to send message over websocket: %w", err)
	}

	_, response, err := wrapper.Conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("failed to read message from websocket: %w", err)
	}

	if t.outputTransformer != nil {
		parser := transformer.NewTextParser()
		outputFormat := configv1.OutputTransformer_OutputFormat_name[int32(t.outputTransformer.GetFormat())]
		return parser.Parse(outputFormat, response, t.outputTransformer.GetExtractionRules(), t.outputTransformer.GetJqQuery())
	}

	var result map[string]any
	if err := json.Unmarshal(response, &result); err != nil {
		return string(response), nil
	}

	return result, nil
}
