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
// WebsocketTool implements the Tool interface for a tool exposed via a WebSocket
// MCPTool returns the MCP tool definition.
//
// Summary: Retrieves the MCP-compatible tool definition.
//
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// Returns:
// Execute handles the execution of the WebSocket tool.
//
// Summary: Executes the tool over WebSocket.
//
// It retrieves a connection from the pool, sends the tool inputs as a message,
// and waits for a single response message, which it then processes and returns.
//
// Parameters:
//   - ctx: context.Context. The execution context.
//   - req: *ExecutionRequest. The request containing input arguments.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Returns:
//   - any: The execution result.
// StreamExecute executes the tool in streaming mode.
//
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Summary: Executes the tool in streaming mode.
//
// Parameters:
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//   - ctx: context.Context. The context for the request.
//   - req: *ExecutionRequest. The request object containing parameters.
//
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Returns:
//   - <-chan any: A channel that emits streaming results.
//   - error: An error if the operation fails or streaming is not supported.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (t *WebsocketTool) StreamExecute(ctx context.Context, req *ExecutionRequest) (<-chan any, error) {
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

// Execute executes the WebSocket tool.
//
// Summary: Executes the WebSocket request and waits for a response.
//
// Parameters:
//   - ctx (context.Context): The context for execution.
//   - req (*ExecutionRequest): The request parameters.
//
// Returns:
//   - any: The response from the WebSocket.
//   - error: An error if the WebSocket communication fails.
//
// Errors:
//   - Returns an error if the websocket pool is not found.
//   - Returns an error if getting a connection from the pool fails.
//   - Returns an error if marshalling the inputs fails.
//   - Returns an error if secret resolution fails.
//   - Returns an error if input transformation fails.
//   - Returns an error if sending or reading a message fails.
//
// Side Effects:
//   - Makes a WebSocket network call.
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
