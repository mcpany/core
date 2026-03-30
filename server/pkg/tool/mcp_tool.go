// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"bytes"
	"context"

	stdjson "encoding/json"
	"fmt"
	"log/slog"
	"sync"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/transformer"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPTool implements the Tool interface for a tool that is exposed via another
// MCP-compliant service.
//
// Summary: Tool implementation for proxying to MCP services.
//
// It acts as a proxy, forwarding the tool call to the
// downstream MCP service.
type MCPTool struct {
	tool                 *v1.Tool
	mcpTool              *mcp.Tool
	mcpToolOnce          sync.Once
	client               client.MCPClient
	inputTransformer     *configv1.InputTransformer
	outputTransformer    *configv1.OutputTransformer
	webhookClient        *WebhookClient
	cache                *configv1.CacheConfig
	cachedInputTemplate  *transformer.TextTemplate
	cachedOutputTemplate *transformer.TextTemplate
	initError            error
}

// NewMCPTool creates a new MCPTool instance.
//
// Summary: Initializes a new MCPTool.
//
// Parameters:
//   - tool: *v1.Tool. The protobuf definition of the tool.
//   - client: client.MCPClient. The MCP client for downstream communication.
//   - callDefinition: *configv1.MCPCallDefinition. The configuration for the MCP call.
//
// Returns:
//   - *MCPTool: The initialized MCPTool.
func NewMCPTool(tool *v1.Tool, client client.MCPClient, callDefinition *configv1.MCPCallDefinition) *MCPTool {
	var webhookClient *WebhookClient
	if it := callDefinition.GetInputTransformer(); it != nil && it.GetWebhook() != nil {
		webhookClient = NewWebhookClient(it.GetWebhook())
	}
	t := &MCPTool{
		tool:              tool,
		client:            client,
		inputTransformer:  callDefinition.GetInputTransformer(),
		outputTransformer: callDefinition.GetOutputTransformer(),
		webhookClient:     webhookClient,
		cache:             callDefinition.GetCache(),
	}

	if it := t.inputTransformer; it != nil && it.GetTemplate() != "" {
		tpl, err := transformer.NewTemplate(it.GetTemplate(), "{{", "}}")
		if err != nil {
			t.initError = fmt.Errorf("failed to parse input template: %w", err)
		} else {
			t.cachedInputTemplate = tpl
		}
	}
	if ot := t.outputTransformer; ot != nil && ot.GetTemplate() != "" {
		tpl, err := transformer.NewTemplate(ot.GetTemplate(), "{{", "}}")
		if err != nil {
			t.initError = fmt.Errorf("failed to parse output template: %w", err)
		} else {
			t.cachedOutputTemplate = tpl
		}
	}
	return t
}

// Tool returns the protobuf definition of the MCP tool.
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
func (t *MCPTool) Tool() *v1.Tool {
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
func (t *MCPTool) MCPTool() *mcp.Tool {
	t.mcpToolOnce.Do(func() {
		var err error
		t.mcpTool, err = ConvertProtoToMCPTool(t.tool)
		if err != nil {
			logging.GetLogger().Error("Failed to convert tool to MCP tool", "toolName", t.tool.GetName(), "error", err)
		}
	})
	return t.mcpTool
}

// GetCacheConfig returns the cache configuration for the MCP tool.
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
func (t *MCPTool) GetCacheConfig() *configv1.CacheConfig {
	return t.cache
}

// Execute handles the execution of the MCP tool.
//
// Summary: Executes the MCP tool call.
//
// It forwards the tool call, including its name and arguments, to the downstream MCP service using the
// configured client and applies any necessary transformations to the request and response.
//
// Parameters:
//   - ctx: context.Context. The execution context.
//   - req: *ExecutionRequest. The execution request.
//
// Returns:
//   - any: The execution result.
//   - error: An error if execution fails.
//
// Side Effects:
//   - Makes an MCP call to the upstream service.
//   - Logs execution details.
//
// IsStreaming returns true if the tool supports streaming.
//
// Summary: Checks if the tool supports streaming execution.
//
// Returns:
//   - bool: True if streaming is supported.
func (t *MCPTool) IsStreaming() bool {
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
func (t *MCPTool) StreamExecute(ctx context.Context, req *ExecutionRequest) (<-chan any, error) {
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

// Execute handles the execution of the MCP tool.
//
// Summary: Executes the MCP tool call.
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
//   - Returns an error if initialization failed.
//   - Returns an error if unmarshalling tool inputs fails.
//   - Returns an error if the transformation webhook fails.
//   - Returns an error if calling the tool on the downstream MCP service fails.
//   - Returns an error if output parsing or transformation fails.
//
// Side Effects:
//   - Makes a call to a downstream MCP service.
func (t *MCPTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	if t.initError != nil {
		return nil, t.initError
	}
	if logging.GetLogger().Enabled(ctx, slog.LevelDebug) {
		logging.GetLogger().Debug("executing tool", "tool", req.ToolName, "inputs", prettyPrint(req.ToolInputs, contentTypeJSON))
	}

	bareToolName := t.tool.GetName()

	var inputs map[string]any
	if len(bytes.TrimSpace(req.ToolInputs)) == 0 {
		req.ToolInputs = []byte("{}")
	}

	if err := fastJSONNumber.Unmarshal(req.ToolInputs, &inputs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool inputs: %w", err)
	}

	var arguments stdjson.RawMessage // Use stdjson for compatibility with SDK or struct? mcp.CallToolParams expects json.RawMessage (from encoding/json)

	switch {
	case t.webhookClient != nil:
		data := map[string]any{
			"kind":      configv1.WebhookKind_WEBHOOK_KIND_TRANSFORM_INPUT,
			"tool_name": req.ToolName,
			"inputs":    inputs,
		}
		respEvent, err := t.webhookClient.Call(ctx, "com.mcpany.tool.transform_input", data)
		if err != nil {
			return nil, fmt.Errorf("transformation webhook failed: %w", err)
		}
		respData := respEvent.Data()
		if len(respData) > 0 {
			arguments = stdjson.RawMessage(respData)
		}
	case t.cachedInputTemplate != nil:
		rendered, err := t.cachedInputTemplate.Render(inputs)
		if err != nil {
			return nil, fmt.Errorf("failed to render input template: %w", err)
		}
		arguments = []byte(rendered)
	case t.inputTransformer != nil && t.inputTransformer.GetTemplate() != "":

		return nil, fmt.Errorf("input template configured but not cached (initialization error?)")
	default:
		arguments = req.ToolInputs
	}

	callToolParams := &mcp.CallToolParams{
		Name:      bareToolName,
		Arguments: arguments,
	}

	result, err := t.client.CallTool(ctx, callToolParams)
	if err != nil {
		return nil, fmt.Errorf("failed to execute tool %q: %w", req.ToolName, err)
	}

	if len(result.Content) == 0 {
		return nil, nil
	}

	var responseBytes []byte
	if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
		responseBytes = []byte(textContent.Text)
	} else {

		responseBytes, err = fastJSON.Marshal(result.Content)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tool output: %w", err)
		}
	}

	if t.outputTransformer != nil {
		if t.outputTransformer.GetFormat() == configv1.OutputTransformer_RAW_BYTES {
			return map[string]any{"raw": responseBytes}, nil
		}
		parser := transformer.NewTextParser()
		outputFormat := configv1.OutputTransformer_OutputFormat_name[int32(t.outputTransformer.GetFormat())]
		parsedResult, err := parser.Parse(outputFormat, responseBytes, t.outputTransformer.GetExtractionRules(), t.outputTransformer.GetJqQuery())
		if err != nil {
			return nil, fmt.Errorf("failed to parse output: %w", err)
		}

		if t.outputTransformer.GetTemplate() != "" {
			if t.cachedOutputTemplate == nil {
				return nil, fmt.Errorf("output template configured but not cached (initialization error?)")
			}
			resultMap, ok := parsedResult.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("output must be a map to be used with a template, got %T", parsedResult)
			}
			renderedOutput, err := t.cachedOutputTemplate.Render(resultMap)
			if err != nil {
				return nil, fmt.Errorf("failed to render output template: %w", err)
			}
			return map[string]any{"result": renderedOutput}, nil
		}
		return parsedResult, nil
	}

	var resultMap map[string]any
	if err := fastJSON.Unmarshal(responseBytes, &resultMap); err != nil {

		return string(responseBytes), nil
	}

	return resultMap, nil
}
