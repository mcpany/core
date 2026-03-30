// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/transformer"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// OpenAPITool implements the Tool interface for a tool defined in an OpenAPI
// specification.
//
// Summary: Tool implementation for OpenAPI services.
//
// It constructs and sends an HTTP request based on the OpenAPI
// operation definition.
type OpenAPITool struct {
	tool                 *v1.Tool
	mcpTool              *mcp.Tool
	mcpToolOnce          sync.Once
	client               client.HTTPClient
	parameterDefs        map[string]string
	method               string
	url                  string
	authenticator        auth.UpstreamAuthenticator
	inputTransformer     *configv1.InputTransformer
	outputTransformer    *configv1.OutputTransformer
	webhookClient        *WebhookClient
	cache                *configv1.CacheConfig
	cachedInputTemplate  *transformer.TextTemplate
	cachedOutputTemplate *transformer.TextTemplate
	initError            error
}

// NewOpenAPITool creates a new OpenAPITool instance.
//
// Summary: Initializes a new OpenAPITool.
//
// Parameters:
//   - tool: *v1.Tool. The protobuf definition of the tool.
//   - client: client.HTTPClient. The HTTP client for requests.
//   - parameterDefs: map[string]string. Mapping of parameter names to their locations.
//   - method: string. The HTTP method.
//   - url: string. The URL template.
//   - authenticator: auth.UpstreamAuthenticator. The authenticator for upstream requests.
//   - callDefinition: *configv1.OpenAPICallDefinition. The configuration for the OpenAPI call.
//
// Returns:
//   - *OpenAPITool: The initialized OpenAPITool.
func NewOpenAPITool(tool *v1.Tool, client client.HTTPClient, parameterDefs map[string]string, method, url string, authenticator auth.UpstreamAuthenticator, callDefinition *configv1.OpenAPICallDefinition) *OpenAPITool {
	var webhookClient *WebhookClient
	if it := callDefinition.GetInputTransformer(); it != nil && it.GetWebhook() != nil {
		webhookClient = NewWebhookClient(it.GetWebhook())
	}
	t := &OpenAPITool{
		tool:              tool,
		client:            client,
		parameterDefs:     parameterDefs,
		method:            method,
		url:               url,
		authenticator:     authenticator,
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

// Tool returns the protobuf definition of the OpenAPI tool.
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
func (t *OpenAPITool) Tool() *v1.Tool {
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
func (t *OpenAPITool) MCPTool() *mcp.Tool {
	t.mcpToolOnce.Do(func() {
		var err error
		t.mcpTool, err = ConvertProtoToMCPTool(t.tool)
		if err != nil {
			logging.GetLogger().Error("Failed to convert tool to MCP tool", "toolName", t.tool.GetName(), "error", err)
		}
	})
	return t.mcpTool
}

// GetCacheConfig returns the cache configuration for the OpenAPI tool.
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
func (t *OpenAPITool) GetCacheConfig() *configv1.CacheConfig {
	return t.cache
}

// Execute handles the execution of the OpenAPI tool.
//
// Summary: Executes the OpenAPI tool call.
//
// It constructs an HTTP request based on the operation's method, URL, and parameter definitions,
// sends the request, and processes the response, applying transformations as needed.
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
//   - Makes an HTTP request to the upstream service.
//   - Logs execution details.
//
// IsStreaming returns true if the tool supports streaming.
//
// Summary: Checks if the tool supports streaming execution.
//
// Returns:
//   - bool: True if streaming is supported.
func (t *OpenAPITool) IsStreaming() bool {
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
func (t *OpenAPITool) StreamExecute(ctx context.Context, req *ExecutionRequest) (<-chan any, error) {
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

// Execute handles the execution of the OpenAPI tool.
//
// Summary: Executes the OpenAPI tool call.
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
//   - Returns an error if URL validation fails.
//   - Returns an error if HTTP request creation or execution fails.
//   - Returns an error if the upstream HTTP response status is >= 400.
//   - Returns an error if reading the response body fails.
//   - Returns an error if output parsing or transformation fails.
//
// Side Effects:
//   - Makes an HTTP request to the upstream service.
func (t *OpenAPITool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	if t.initError != nil {
		return nil, t.initError
	}
	if logging.GetLogger().Enabled(ctx, slog.LevelDebug) {
		logging.GetLogger().Debug("executing tool", "tool", req.ToolName, "inputs", prettyPrint(req.ToolInputs, contentTypeJSON))
	}
	var inputs map[string]any
	if len(bytes.TrimSpace(req.ToolInputs)) == 0 {
		req.ToolInputs = []byte("{}")
	}

	if err := fastJSONNumber.Unmarshal(req.ToolInputs, &inputs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool inputs: %w", err)
	}

	url := t.url
	for paramName, paramValue := range inputs {
		if t.parameterDefs[paramName] == "path" {
			url = strings.ReplaceAll(url, "{{"+paramName+"}}", util.ToString(paramValue))
			delete(inputs, paramName)
		}
	}

	var body io.Reader
	var contentType string
	if t.method == http.MethodPost || t.method == http.MethodPut {
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
				body = bytes.NewReader(respData)
				if fastJSON.Valid(respData) {
					contentType = contentTypeJSON
				}
			}
		case t.cachedInputTemplate != nil:
			renderedBody, err := t.cachedInputTemplate.Render(inputs)
			if err != nil {
				return nil, fmt.Errorf("failed to render input template: %w", err)
			}
			body = strings.NewReader(renderedBody)
		case t.inputTransformer != nil && t.inputTransformer.GetTemplate() != "":

			return nil, fmt.Errorf("input template configured but not cached (initialization error?)")
		default:
			jsonBytes, err := fastJSON.Marshal(inputs)
			if err != nil {
				return "", fmt.Errorf("failed to marshal tool inputs to json: %w", err)
			}
			body = bytes.NewReader(jsonBytes)
			contentType = contentTypeJSON
		}
	}

	if err := validation.IsSafeURL(url); err != nil {
		return nil, fmt.Errorf("unsafe url: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, t.method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	if contentType != "" {
		httpReq.Header.Set("Content-Type", contentType)
	}

	if t.authenticator != nil {
		if err := t.authenticator.Authenticate(httpReq); err != nil {
			return nil, fmt.Errorf("failed to authenticate OpenAPI request: %w", err)
		}
	}

	if t.method == http.MethodGet {
		q := httpReq.URL.Query()
		for paramName, paramValue := range inputs {
			if t.parameterDefs[paramName] == "query" {
				if slice, ok := paramValue.([]interface{}); ok {
					for _, v := range slice {
						q.Add(paramName, util.ToString(v))
					}
				} else {
					q.Add(paramName, util.ToString(paramValue))
				}
			}
		}
		httpReq.URL.RawQuery = q.Encode()
	}

	resp, err := t.client.Do(httpReq)
	if err != nil {

		logging.GetLogger().ErrorContext(ctx, "Failed to execute OpenAPI HTTP request", "tool", t.tool.GetName(), "error", err)
		return nil, fmt.Errorf("failed to execute http request")
	}
	defer func() { _ = resp.Body.Close() }()

	maxSize := getMaxHTTPResponseSize()

	reader := io.LimitReader(resp.Body, maxSize+1)
	respBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read http response body: %w", err)
	}
	if int64(len(respBody)) > maxSize {
		return nil, fmt.Errorf("response body exceeds maximum size of %d bytes", maxSize)
	}

	if resp.StatusCode >= 400 {

		logging.GetLogger().ErrorContext(ctx, "Upstream OpenAPI request failed", "tool", t.tool.GetName(), "status", resp.StatusCode, "response", string(respBody))
		return nil, fmt.Errorf("upstream OpenAPI request failed with status %d", resp.StatusCode)
	}

	if t.outputTransformer != nil {
		if t.outputTransformer.GetFormat() == configv1.OutputTransformer_RAW_BYTES {
			return map[string]any{"raw": respBody}, nil
		}
		parser := transformer.NewTextParser()
		outputFormat := configv1.OutputTransformer_OutputFormat_name[int32(t.outputTransformer.GetFormat())]
		parsedResult, err := parser.Parse(outputFormat, respBody, t.outputTransformer.GetExtractionRules(), t.outputTransformer.GetJqQuery())
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

	var result map[string]any
	if err := fastJSON.Unmarshal(respBody, &result); err != nil {
		return string(respBody), nil
	}

	return result, nil
}
