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

package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/mcpxy/core/pkg/auth"
	"github.com/mcpxy/core/pkg/client"
	"github.com/mcpxy/core/pkg/pool"
	"github.com/mcpxy/core/pkg/transformer"
	"github.com/mcpxy/core/pkg/util"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	v1 "github.com/mcpxy/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

// ToolManagerInterface defines the contract for a tool manager, which is
// responsible for the lifecycle and execution of tools within the system.
type ToolManagerInterface interface {
	// GetTool retrieves a tool by its fully qualified name.
	GetTool(toolName string) (Tool, bool)
	// ListTools returns a slice of all registered tools.
	ListTools() []Tool
	// ExecuteTool runs a tool with the given request parameters.
	ExecuteTool(ctx context.Context, req *ExecutionRequest) (any, error)
	// SetMCPServer provides the tool manager with a reference to the MCP server,
	// allowing tools to make calls back to the server if needed.
	SetMCPServer(mcpServer MCPServerProvider)
	// AddTool registers a new tool with the manager.
	AddTool(tool Tool) error
	// GetServiceInfo retrieves metadata about a registered service.
	GetServiceInfo(serviceID string) (*ServiceInfo, bool)
	// AddServiceInfo registers metadata for a service.
	AddServiceInfo(serviceID string, info *ServiceInfo)
	// ClearToolsForService removes all tools associated with a specific service.
	ClearToolsForService(serviceKey string)
}

// Tool is the fundamental interface for any executable tool in the system.
// Each implementation represents a different type of underlying service
// (e.g., gRPC, HTTP, command-line).
type Tool interface {
	// Tool returns the protobuf definition of the tool.
	Tool() *v1.Tool
	// Execute runs the tool with the provided context and request, returning
	// the result or an error.
	Execute(ctx context.Context, req *ExecutionRequest) (any, error)
}

// ServiceInfo holds metadata about a registered upstream service, including its
// configuration and any associated protobuf file descriptors.
type ServiceInfo struct {
	Name   string
	Config *configv1.UpstreamServiceConfig
	Fds    *descriptorpb.FileDescriptorSet
}

// ExecutionRequest represents a request to execute a specific tool, including
// its name and input arguments as a raw JSON message.
type ExecutionRequest struct {
	ToolName   string
	ToolInputs json.RawMessage
}

// ServiceRegistry defines an interface for a component that can look up tools
// and service information. It is used for dependency injection to decouple
// components from the main service registry.
type ServiceRegistry interface {
	GetTool(toolName string) (Tool, bool)
	GetServiceInfo(serviceID string) (*ServiceInfo, bool)
}

// GRPCTool implements the Tool interface for a tool that is exposed via a gRPC
// endpoint. It handles the marshalling of JSON inputs to protobuf messages and
// invoking the gRPC method.
type GRPCTool struct {
	tool           *v1.Tool
	poolManager    *pool.Manager
	serviceKey     string
	method         protoreflect.MethodDescriptor
	requestMessage protoreflect.ProtoMessage
}

// NewGRPCTool creates a new GRPCTool.
//
// tool is the protobuf definition of the tool.
// poolManager is used to get a gRPC client from the connection pool.
// serviceKey identifies the specific gRPC service connection pool.
// method is the protobuf descriptor for the gRPC method to be called.
func NewGRPCTool(tool *v1.Tool, poolManager *pool.Manager, serviceKey string, method protoreflect.MethodDescriptor) *GRPCTool {
	return &GRPCTool{
		tool:           tool,
		poolManager:    poolManager,
		serviceKey:     serviceKey,
		method:         method,
		requestMessage: dynamicpb.NewMessage(method.Input()),
	}
}

// Tool returns the protobuf definition of the gRPC tool.
func (t *GRPCTool) Tool() *v1.Tool {
	return t.tool
}

// Execute handles the execution of the gRPC tool. It retrieves a client from the
// pool, unmarshals the JSON input into a protobuf request message, invokes the
// gRPC method, and marshals the protobuf response back to JSON.
func (t *GRPCTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	grpcPool, ok := pool.Get[*client.GrpcClientWrapper](t.poolManager, t.serviceKey)
	if !ok {
		return nil, fmt.Errorf("no grpc pool found for service: %s", t.serviceKey)
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
	if err := grpcClient.Invoke(ctx, grpcMethodName, t.requestMessage, responseMessage); err != nil {
		return nil, fmt.Errorf("failed to invoke grpc method: %w", err)
	}

	responseJSON, err := protojson.Marshal(responseMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal grpc response to json: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(responseJSON, &result); err != nil {
		return string(responseJSON), nil
	}

	return result, nil
}

// HTTPTool implements the Tool interface for a tool exposed via an HTTP endpoint.
// It constructs and sends an HTTP request based on the tool definition and
// input, handling parameter mapping, authentication, and transformations.
type HTTPTool struct {
	tool              *v1.Tool
	poolManager       *pool.Manager
	serviceKey        string
	authenticator     auth.UpstreamAuthenticator
	parameterMappings []*configv1.HttpParameterMapping
	inputTransformer  *configv1.InputTransformer
	outputTransformer *configv1.OutputTransformer
}

// NewHTTPTool creates a new HTTPTool.
//
// tool is the protobuf definition of the tool.
// poolManager is used to get an HTTP client from the connection pool.
// serviceKey identifies the specific HTTP service connection pool.
// authenticator handles adding authentication credentials to the request.
// callDefinition contains the configuration for the HTTP call, such as
// parameter mappings and transformers.
func NewHTTPTool(tool *v1.Tool, poolManager *pool.Manager, serviceKey string, authenticator auth.UpstreamAuthenticator, callDefinition *configv1.HttpCallDefinition) *HTTPTool {
	return &HTTPTool{
		tool:              tool,
		poolManager:       poolManager,
		serviceKey:        serviceKey,
		authenticator:     authenticator,
		parameterMappings: callDefinition.GetParameterMappings(),
		inputTransformer:  callDefinition.GetInputTransformer(),
		outputTransformer: callDefinition.GetOutputTransformer(),
	}
}

// Tool returns the protobuf definition of the HTTP tool.
func (t *HTTPTool) Tool() *v1.Tool {
	return t.tool
}

// Execute handles the execution of the HTTP tool. It builds an HTTP request by
// mapping input parameters to the path, query, and body, applies any
// configured transformations, sends the request, and processes the response.
func (t *HTTPTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	httpPool, ok := pool.Get[*client.HttpClientWrapper](t.poolManager, t.serviceKey)
	if !ok {
		return nil, fmt.Errorf("no http pool found for service: %s", t.serviceKey)
	}

	httpClient, err := httpPool.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get client from pool: %w", err)
	}
	defer httpPool.Put(httpClient)

	methodAndURL := strings.Fields(t.tool.GetUnderlyingMethodFqn())
	if len(methodAndURL) != 2 {
		return "", fmt.Errorf("invalid http tool definition: expected method and URL, got %q", t.tool.GetUnderlyingMethodFqn())
	}
	method, url := methodAndURL[0], methodAndURL[1]

	var inputs map[string]any
	if len(req.ToolInputs) > 0 {
		if err := json.Unmarshal(req.ToolInputs, &inputs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tool inputs: %w", err)
		}
	}

	for _, param := range t.parameterMappings {
		if param.GetLocation() == configv1.HttpParameterMapping_PATH {
			if val, ok := inputs[param.GetInputParameterName()]; ok {
				url = strings.ReplaceAll(url, "{"+param.GetTargetParameterName()+"}", fmt.Sprintf("%v", val))
				delete(inputs, param.GetInputParameterName())
			}
		}
	}

	var body io.Reader
	var contentType string
	if inputs != nil {
		if method == http.MethodPost || method == http.MethodPut {
			if t.inputTransformer != nil && t.inputTransformer.GetTemplate() != "" {
				tpl, err := transformer.NewTextTemplate(t.inputTransformer.GetTemplate())
				if err != nil {
					return nil, fmt.Errorf("failed to create input template: %w", err)
				}
				renderedBody, err := tpl.Render(inputs)
				if err != nil {
					return nil, fmt.Errorf("failed to render input template: %w", err)
				}
				body = strings.NewReader(renderedBody)
			} else {
				jsonBytes, err := json.Marshal(inputs)
				if err != nil {
					return "", fmt.Errorf("failed to marshal tool inputs to json: %w", err)
				}
				body = bytes.NewReader(jsonBytes)
				contentType = "application/json"
			}
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return "", fmt.Errorf("failed to create http request: %w", err)
	}

	if contentType != "" {
		httpReq.Header.Set("Content-Type", contentType)
	}

	if t.authenticator != nil {
		if err := t.authenticator.Authenticate(httpReq); err != nil {
			return nil, fmt.Errorf("failed to authenticate request: %w", err)
		}
	}

	if method == http.MethodGet || method == http.MethodDelete {
		q := httpReq.URL.Query()
		for key, value := range inputs {
			q.Add(key, fmt.Sprintf("%v", value))
		}
		httpReq.URL.RawQuery = q.Encode()
	}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read http response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("upstream HTTP request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	if t.outputTransformer != nil {
		parser := transformer.NewTextParser()
		outputFormat := configv1.OutputTransformer_OutputFormat_name[int32(t.outputTransformer.GetFormat())]
		parsedResult, err := parser.Parse(outputFormat, respBody, t.outputTransformer.GetExtractionRules())
		if err != nil {
			return nil, fmt.Errorf("failed to parse output: %w", err)
		}

		if t.outputTransformer.GetTemplate() != "" {
			tpl, err := transformer.NewTextTemplate(t.outputTransformer.GetTemplate())
			if err != nil {
				return nil, fmt.Errorf("failed to create output template: %w", err)
			}
			renderedOutput, err := tpl.Render(parsedResult)
			if err != nil {
				return nil, fmt.Errorf("failed to render output template: %w", err)
			}
			return map[string]any{"result": renderedOutput}, nil
		}
		return parsedResult, nil
	}

	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return string(respBody), nil
	}

	return result, nil
}

// MCPTool implements the Tool interface for a tool that is exposed via another
// MCP-compliant service. It acts as a proxy, forwarding the tool call to the
// downstream MCP service.
type MCPTool struct {
	tool              *v1.Tool
	client            client.MCPClient
	inputTransformer  *configv1.InputTransformer
	outputTransformer *configv1.OutputTransformer
}

// NewMCPTool creates a new MCPTool.
//
// tool is the protobuf definition of the tool.
// client is the MCP client used to communicate with the downstream service.
// callDefinition contains configuration for input/output transformations.
func NewMCPTool(tool *v1.Tool, client client.MCPClient, callDefinition *configv1.MCPCallDefinition) *MCPTool {
	return &MCPTool{
		tool:              tool,
		client:            client,
		inputTransformer:  callDefinition.GetInputTransformer(),
		outputTransformer: callDefinition.GetOutputTransformer(),
	}
}

// Tool returns the protobuf definition of the MCP tool.
func (t *MCPTool) Tool() *v1.Tool {
	return t.tool
}

// Execute handles the execution of the MCP tool. It forwards the tool call,
// including its name and arguments, to the downstream MCP service using the
// configured client and applies any necessary transformations to the request
// and response.
func (t *MCPTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	_, bareToolName, err := util.ParseToolName(req.ToolName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tool name: %w", err)
	}

	var inputs map[string]any
	if err := json.Unmarshal(req.ToolInputs, &inputs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool inputs: %w", err)
	}

	var arguments json.RawMessage
	if t.inputTransformer != nil && t.inputTransformer.GetTemplate() != "" {
		tpl, err := transformer.NewTextTemplate(t.inputTransformer.GetTemplate())
		if err != nil {
			return nil, fmt.Errorf("failed to create input template: %w", err)
		}
		rendered, err := tpl.Render(inputs)
		if err != nil {
			return nil, fmt.Errorf("failed to render input template: %w", err)
		}
		arguments = []byte(rendered)
	} else {
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
		return nil, nil // No content to parse
	}

	var responseBytes []byte
	if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
		responseBytes = []byte(textContent.Text)
	} else {
		// Fallback for other content types - marshal the whole content part
		responseBytes, err = json.Marshal(result.Content)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tool output: %w", err)
		}
	}

	if t.outputTransformer != nil {
		parser := transformer.NewTextParser()
		outputFormat := configv1.OutputTransformer_OutputFormat_name[int32(t.outputTransformer.GetFormat())]
		parsedResult, err := parser.Parse(outputFormat, responseBytes, t.outputTransformer.GetExtractionRules())
		if err != nil {
			return nil, fmt.Errorf("failed to parse output: %w", err)
		}

		if t.outputTransformer.GetTemplate() != "" {
			tpl, err := transformer.NewTextTemplate(t.outputTransformer.GetTemplate())
			if err != nil {
				return nil, fmt.Errorf("failed to create output template: %w", err)
			}
			renderedOutput, err := tpl.Render(parsedResult)
			if err != nil {
				return nil, fmt.Errorf("failed to render output template: %w", err)
			}
			return map[string]any{"result": renderedOutput}, nil
		}
		return parsedResult, nil
	}

	var resultMap map[string]any
	if err := json.Unmarshal(responseBytes, &resultMap); err != nil {
		// If unmarshalling to a map fails, return the raw string content
		return string(responseBytes), nil
	}

	return resultMap, nil
}

// OpenAPITool implements the Tool interface for a tool defined in an OpenAPI
// specification. It constructs and sends an HTTP request based on the OpenAPI
// operation definition.
type OpenAPITool struct {
	tool              *v1.Tool
	client            client.HttpClient
	parameterDefs     map[string]string
	method            string
	url               string
	authenticator     auth.UpstreamAuthenticator
	inputTransformer  *configv1.InputTransformer
	outputTransformer *configv1.OutputTransformer
}

// NewOpenAPITool creates a new OpenAPITool.
//
// tool is the protobuf definition of the tool.
// client is the HTTP client used to make the request.
// parameterDefs maps parameter names to their location (e.g., "path", "query").
// method is the HTTP method for the operation.
// url is the URL template for the endpoint.
// authenticator handles adding authentication credentials to the request.
// callDefinition contains configuration for input/output transformations.
func NewOpenAPITool(tool *v1.Tool, client client.HttpClient, parameterDefs map[string]string, method, url string, authenticator auth.UpstreamAuthenticator, callDefinition *configv1.OpenAPICallDefinition) *OpenAPITool {
	return &OpenAPITool{
		tool:              tool,
		client:            client,
		parameterDefs:     parameterDefs,
		method:            method,
		url:               url,
		authenticator:     authenticator,
		inputTransformer:  callDefinition.GetInputTransformer(),
		outputTransformer: callDefinition.GetOutputTransformer(),
	}
}

// Tool returns the protobuf definition of the OpenAPI tool.
func (t *OpenAPITool) Tool() *v1.Tool {
	return t.tool
}

// Execute handles the execution of the OpenAPI tool. It constructs an HTTP
// request based on the operation's method, URL, and parameter definitions,
// sends the request, and processes the response, applying transformations as
// needed.
func (t *OpenAPITool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	var inputs map[string]any
	if err := json.Unmarshal(req.ToolInputs, &inputs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool inputs: %w", err)
	}

	url := t.url
	for paramName, paramValue := range inputs {
		if t.parameterDefs[paramName] == "path" {
			url = strings.ReplaceAll(url, "{"+paramName+"}", fmt.Sprintf("%v", paramValue))
			delete(inputs, paramName)
		}
	}

	var body io.Reader
	var contentType string
	if t.method == http.MethodPost || t.method == http.MethodPut {
		if t.inputTransformer != nil && t.inputTransformer.GetTemplate() != "" {
			tpl, err := transformer.NewTextTemplate(t.inputTransformer.GetTemplate())
			if err != nil {
				return nil, fmt.Errorf("failed to create input template: %w", err)
			}
			renderedBody, err := tpl.Render(inputs)
			if err != nil {
				return nil, fmt.Errorf("failed to render input template: %w", err)
			}
			body = strings.NewReader(renderedBody)
		} else {
			jsonBytes, err := json.Marshal(inputs)
			if err != nil {
				return "", fmt.Errorf("failed to marshal tool inputs to json: %w", err)
			}
			body = bytes.NewReader(jsonBytes)
			contentType = "application/json"
		}
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
				q.Add(paramName, fmt.Sprintf("%v", paramValue))
			}
		}
		httpReq.URL.RawQuery = q.Encode()
	}

	resp, err := t.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read http response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("upstream OpenAPI request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	if t.outputTransformer != nil {
		parser := transformer.NewTextParser()
		outputFormat := configv1.OutputTransformer_OutputFormat_name[int32(t.outputTransformer.GetFormat())]
		parsedResult, err := parser.Parse(outputFormat, respBody, t.outputTransformer.GetExtractionRules())
		if err != nil {
			return nil, fmt.Errorf("failed to parse output: %w", err)
		}

		if t.outputTransformer.GetTemplate() != "" {
			tpl, err := transformer.NewTextTemplate(t.outputTransformer.GetTemplate())
			if err != nil {
				return nil, fmt.Errorf("failed to create output template: %w", err)
			}
			renderedOutput, err := tpl.Render(parsedResult)
			if err != nil {
				return nil, fmt.Errorf("failed to render output template: %w", err)
			}
			return map[string]any{"result": renderedOutput}, nil
		}
		return parsedResult, nil
	}

	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return string(respBody), nil
	}

	return result, nil
}

// CommandTool implements the Tool interface for a tool that is executed as a
// local command-line process. It maps tool inputs to command-line arguments and
// environment variables.
type CommandTool struct {
	tool    *v1.Tool
	command string
}

// NewCommandTool creates a new CommandTool.
//
// tool is the protobuf definition of the tool.
// command is the command to be executed.
func NewCommandTool(tool *v1.Tool, command string) *CommandTool {
	return &CommandTool{
		tool:    tool,
		command: command,
	}
}

// Tool returns the protobuf definition of the command-line tool.
func (t *CommandTool) Tool() *v1.Tool {
	return t.tool
}

// Execute handles the execution of the command-line tool. It constructs a command
// with arguments and environment variables derived from the tool inputs, runs
// the command, and returns its output.
func (t *CommandTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	var inputs map[string]any
	if err := json.Unmarshal(req.ToolInputs, &inputs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool inputs: %w", err)
	}

	var args []string
	if inputs != nil {
		if argsVal, ok := inputs["args"]; ok {
			if argsList, ok := argsVal.([]interface{}); ok {
				for _, arg := range argsList {
					if argStr, ok := arg.(string); ok {
						args = append(args, argStr)
					} else {
						return nil, fmt.Errorf("non-string value in 'args' array")
					}
				}
			} else {
				return nil, fmt.Errorf("'args' parameter must be an array of strings")
			}
		}
	}

	cmd := exec.CommandContext(ctx, t.command, args...)
	env := os.Environ()
	for key, value := range inputs {
		if key != "args" {
			env = append(env, fmt.Sprintf("%s=%v", key, value))
		}
	}
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w, output: %s", err, string(output))
	}

	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		return string(output), nil
	}

	return result, nil
}
