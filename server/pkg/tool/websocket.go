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

// WebsocketTool implements the Tool interface for a tool exposed via a WebSocket.
//
// Summary: implements the Tool interface for a tool exposed via a WebSocket.
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
// Summary: creates a new WebsocketTool.
//
// Parameters:
//   - tool: *v1.Tool. The tool.
//   - poolManager: *pool.Manager. The poolManager.
//   - serviceID: string. The serviceID.
//   - authenticator: auth.UpstreamAuthenticator. The authenticator.
//   - callDefinition: *configv1.WebsocketCallDefinition. The callDefinition.
//
// Returns:
//   - *WebsocketTool: The *WebsocketTool.
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
// Summary: returns the protobuf definition of the WebSocket tool.
//
// Parameters:
//   None.
//
// Returns:
//   - *v1.Tool: The *v1.Tool.
func (t *WebsocketTool) Tool() *v1.Tool {
	return t.tool
}

// MCPTool returns the MCP tool definition.
//
// Summary: returns the MCP tool definition.
//
// Parameters:
//   None.
//
// Returns:
//   - *mcp.Tool: The *mcp.Tool.
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
// Summary: returns the cache configuration for the WebSocket tool.
//
// Parameters:
//   None.
//
// Returns:
//   - *configv1.CacheConfig: The *configv1.CacheConfig.
func (t *WebsocketTool) GetCacheConfig() *configv1.CacheConfig {
	return t.cache
}

// Execute handles the execution of the WebSocket tool. It retrieves a connection.
//
// Summary: handles the execution of the WebSocket tool. It retrieves a connection.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - req: *ExecutionRequest. The req.
//
// Returns:
//   - any: The any.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
