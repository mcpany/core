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

package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/client"
	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/transformer"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
)

// WebsocketTool implements the Tool interface for a tool exposed via a WebSocket
// connection. It handles sending and receiving messages over a persistent
// WebSocket connection managed by a connection pool.
type WebsocketTool struct {
	tool              *v1.Tool
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
func (t *WebsocketTool) Tool() *v1.Tool {
	return t.tool
}

// GetCacheConfig returns the cache configuration for the WebSocket tool.
func (t *WebsocketTool) GetCacheConfig() *configv1.CacheConfig {
	return t.cache
}

// ServiceID returns the ID of the service that the tool belongs to.
func (t *WebsocketTool) ServiceID() string {
	return t.serviceID
}

// Execute handles the execution of the WebSocket tool. It retrieves a connection
// from the pool, sends the tool inputs as a message, and waits for a single
// response message, which it then processes and returns.
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
			secretValue, err := util.ResolveSecret(secret)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve secret for parameter %q: %w", param.GetSchema().GetName(), err)
			}
			inputs[param.GetSchema().GetName()] = secretValue
		}
	}

	var message []byte
	if t.inputTransformer != nil && t.inputTransformer.GetTemplate() != "" {
		tpl, err := transformer.NewTemplate(t.inputTransformer.GetTemplate(), "{{", "}}")
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
		return parser.Parse(outputFormat, response, t.outputTransformer.GetExtractionRules())
	}

	var result map[string]any
	if err := json.Unmarshal(response, &result); err != nil {
		return string(response), nil
	}

	return result, nil
}
