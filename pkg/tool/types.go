
// Copyright 2024 Author(s) of MCP any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/client"
	"github.com/mcpany/core/pkg/command"
	"github.com/mcpany/core/pkg/consts"
	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/metrics"
	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/resilience"
	"github.com/mcpany/core/pkg/transformer"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

// Tool is the fundamental interface for any executable tool in the system.
// Each implementation represents a different type of underlying service
// (e.g., gRPC, HTTP, command-line).
type Tool interface {
	// Tool returns the protobuf definition of the tool.
	Tool() *v1.Tool
	// Execute runs the tool with the provided context and request, returning
	// the result or an error.
	Execute(ctx context.Context, req *ExecutionRequest) (any, error)
	// GetCacheConfig returns the cache configuration for the tool.
	GetCacheConfig() *configv1.CacheConfig
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
	// ToolInputs is the raw JSON message of the tool inputs. It is used by
	// tools that need to unmarshal the inputs into a specific struct.
	ToolInputs json.RawMessage
	// Arguments is a map of the tool inputs. It is used by tools that need to
	// access the inputs as a map.
	Arguments map[string]interface{}
}

// ServiceRegistry defines an interface for a component that can look up tools
// and service information. It is used for dependency injection to decouple
// components from the main service registry.
type ServiceRegistry interface {
	GetTool(toolName string) (Tool, bool)
	GetServiceInfo(serviceID string) (*ServiceInfo, bool)
}

// ToolExecutionFunc represents the next middleware in the chain.
type ToolExecutionFunc func(ctx context.Context, req *ExecutionRequest) (any, error)

type contextKey string

const toolContextKey = contextKey("tool")

// NewContextWithTool creates a new context with the given tool.
func NewContextWithTool(ctx context.Context, t Tool) context.Context {
	return context.WithValue(ctx, toolContextKey, t)
}

// GetFromContext retrieves a tool from the context.
func GetFromContext(ctx context.Context) (Tool, bool) {
	t, ok := ctx.Value(toolContextKey).(Tool)
	return t, ok
}

// Callable is an interface that represents a callable tool.
type Callable interface {
	Call(ctx context.Context, req *ExecutionRequest) (any, error)
}

// GRPCTool implements the Tool interface for a tool that is exposed via a gRPC
// endpoint. It handles the marshalling of JSON inputs to protobuf messages and
// invoking the gRPC method.
type GRPCTool struct {
	tool           *v1.Tool
	poolManager    *pool.Manager
	serviceID      string
	method         protoreflect.MethodDescriptor
	requestMessage protoreflect.ProtoMessage
	cache          *configv1.CacheConfig
}

// NewGRPCTool creates a new GRPCTool.
//
// tool is the protobuf definition of the tool.
// poolManager is used to get a gRPC client from the connection pool.
// serviceID identifies the specific gRPC service connection pool.
// method is the protobuf descriptor for the gRPC method to be called.
func NewGRPCTool(tool *v1.Tool, poolManager *pool.Manager, serviceID string, method protoreflect.MethodDescriptor, callDefinition *configv1.GrpcCallDefinition) *GRPCTool {
	return &GRPCTool{
		tool:           tool,
		poolManager:    poolManager,
		serviceID:      serviceID,
		method:         method,
		requestMessage: dynamicpb.NewMessage(method.Input()),
		cache:          callDefinition.GetCache(),
	}
}

// Tool returns the protobuf definition of the gRPC tool.
func (t *GRPCTool) Tool() *v1.Tool {
	return t.tool
}

// GetCacheConfig returns the cache configuration for the gRPC tool.
func (t *GRPCTool) GetCacheConfig() *configv1.CacheConfig {
	return t.cache
}

// Execute handles the execution of the gRPC tool. It retrieves a client from the
// pool, unmarshals the JSON input into a protobuf request message, invokes the
// gRPC method, and marshals the protobuf response back to JSON.
func (t *GRPCTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	defer metrics.MeasureSince([]string{"grpc", "request", "latency"}, time.Now())
	grpcPool, ok := pool.Get[*client.GrpcClientWrapper](t.poolManager, t.serviceID)
	if !ok {
		metrics.IncrCounter([]string{"grpc", "request", "error"}, 1)
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
	if err := grpcClient.Invoke(ctx, grpcMethodName, t.requestMessage, responseMessage); err != nil {
		metrics.IncrCounter([]string{"grpc", "request", "error"}, 1)
		return nil, fmt.Errorf("failed to invoke grpc method: %w", err)
	}
	metrics.IncrCounter([]string{"grpc", "request", "success"}, 1)

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
	serviceID         string
	authenticator     auth.UpstreamAuthenticator
	parameters        []*configv1.HttpParameterMapping
	inputTransformer  *configv1.InputTransformer
	outputTransformer *configv1.OutputTransformer
	cache             *configv1.CacheConfig
	resilienceManager *resilience.Manager
}

// NewHTTPTool creates a new HTTPTool.
//
// tool is the protobuf definition of the tool.
// poolManager is used to get an HTTP client from the connection pool.
// serviceID identifies the specific HTTP service connection pool.
// authenticator handles adding authentication credentials to the request.
// callDefinition contains the configuration for the HTTP call, such as
// parameter mappings and transformers.
func NewHTTPTool(tool *v1.Tool, poolManager *pool.Manager, serviceID string, authenticator auth.UpstreamAuthenticator, callDefinition *configv1.HttpCallDefinition, cfg *configv1.ResilienceConfig) *HTTPTool {
	return &HTTPTool{
		tool:              tool,
		poolManager:       poolManager,
		serviceID:         serviceID,
		authenticator:     authenticator,
		parameters:        callDefinition.GetParameters(),
		inputTransformer:  callDefinition.GetInputTransformer(),
		outputTransformer: callDefinition.GetOutputTransformer(),
		cache:             callDefinition.GetCache(),
		resilienceManager: resilience.NewManager(cfg),
	}
}

// Tool returns the protobuf definition of the HTTP tool.
func (t *HTTPTool) Tool() *v1.Tool {
	return t.tool
}

// GetCacheConfig returns the cache configuration for the HTTP tool.
func (t *HTTPTool) GetCacheConfig() *configv1.CacheConfig {
	return t.cache
}

// Execute handles the execution of the HTTP tool. It builds an HTTP request by
// mapping input parameters to the path, query, and body, applies any
// configured transformations, sends the request, and processes the response.
func (t *HTTPTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	defer metrics.MeasureSince([]string{"http", "request", "latency"}, time.Now())

	httpPool, ok := pool.Get[*client.HttpClientWrapper](t.poolManager, t.serviceID)
	if !ok {
		metrics.IncrCounter([]string{"http", "request", "error"}, 1)
		return nil, fmt.Errorf("no http pool found for service: %s", t.serviceID)
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
	method, rawURL := methodAndURL[0], methodAndURL[1]

	urlString, err := url.PathUnescape(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to unescape url: %w", err)
	}

	// Replace placeholders in the URL path
	urlString = util.ReplaceURLPath(urlString, req.Arguments)

	var inputs map[string]any
	if len(req.ToolInputs) > 0 {
		if err := json.Unmarshal(req.ToolInputs, &inputs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tool inputs: %w", err)
		}
	}

	for _, param := range t.parameters {
		if secret := param.GetSecret(); secret != nil {
			secretValue, err := util.ResolveSecret(secret)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve secret for parameter %q: %w", param.GetSchema().GetName(), err)
			}
			urlString = strings.ReplaceAll(urlString, "{{"+param.GetSchema().GetName()+"}}", secretValue)
		} else if schema := param.GetSchema(); schema != nil {
			placeholder := "{{" + schema.GetName() + "}}"
			if val, ok := inputs[schema.GetName()]; ok {
				urlString = strings.ReplaceAll(urlString, placeholder, url.PathEscape(fmt.Sprintf("%v", val)))
				delete(inputs, schema.GetName())
			} else {
				urlString = strings.ReplaceAll(urlString, "/"+placeholder, "")
			}
		}
	}

	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}
	parsedURL.Path = path.Clean(parsedURL.Path)
	urlString = parsedURL.String()

	var body io.Reader
	var contentType string
	if inputs != nil {
		if method == http.MethodPost || method == http.MethodPut {
			if t.inputTransformer != nil && t.inputTransformer.GetTemplate() != "" {
				tpl, err := transformer.NewTemplate(t.inputTransformer.GetTemplate(), "{{", "}}")
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

	var resp *http.Response
	work := func() error {
		var bodyForAttempt io.Reader
		if body != nil {
			if seeker, ok := body.(io.Seeker); ok {
				seeker.Seek(0, io.SeekStart)
				bodyForAttempt = body
			} else {
				return &resilience.PermanentError{Err: fmt.Errorf("cannot retry request with non-seekable body")}
			}
		}

		httpReq, err := http.NewRequestWithContext(ctx, method, urlString, bodyForAttempt)
		if err != nil {
			return &resilience.PermanentError{Err: fmt.Errorf("failed to create http request: %w", err)}
		}

		if contentType != "" {
			httpReq.Header.Set("Content-Type", contentType)
		}
		httpReq.Header.Set("User-Agent", "mcpany-e2e-test")
		httpReq.Header.Set("Accept", "*/*")

		if t.authenticator != nil {
			if err := t.authenticator.Authenticate(httpReq); err != nil {
				return &resilience.PermanentError{Err: fmt.Errorf("failed to authenticate request: %w", err)}
			}
		}

		if method == http.MethodGet || method == http.MethodDelete {
			q := httpReq.URL.Query()
			for key, value := range inputs {
				q.Add(key, fmt.Sprintf("%v", value))
			}
			httpReq.URL.RawQuery = q.Encode()
		}

		if logging.GetLogger().Enabled(ctx, slog.LevelDebug) {
			reqDump, err := httputil.DumpRequestOut(httpReq, true)
			if err != nil {
				logging.GetLogger().Error("failed to dump http request", "error", err)
			} else {
				logging.GetLogger().Debug("sending http request", "request", string(reqDump))
			}
		}

		resp, err = httpClient.Do(httpReq)
		if err != nil {
			return fmt.Errorf("failed to execute http request: %w", err)
		}

		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return &resilience.PermanentError{Err: fmt.Errorf("upstream HTTP request failed with status %d", resp.StatusCode)}
		}

		if resp.StatusCode >= 500 {
			return fmt.Errorf("upstream HTTP request failed with status %d", resp.StatusCode)
		}
		return nil
	}

	if err := t.resilienceManager.Execute(work); err != nil {
		metrics.IncrCounter([]string{"http", "request", "error"}, 1)
		return nil, err
	}

	defer resp.Body.Close()
	metrics.IncrCounter([]string{"http", "request", "success"}, 1)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read http response body: %w", err)
	}

	if logging.GetLogger().Enabled(ctx, slog.LevelDebug) {
		resp.Body = io.NopCloser(bytes.NewBuffer(respBody))
		respDump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			logging.GetLogger().Error("failed to dump http response", "error", err)
		} else {
			logging.GetLogger().Debug("received http response", "response", string(respDump))
		}
	}

	if t.outputTransformer != nil {
		if t.outputTransformer.GetFormat() == configv1.OutputTransformer_RAW_BYTES {
			return map[string]any{"raw": respBody}, nil
		}

		parser := transformer.NewTextParser()
		outputFormat := configv1.OutputTransformer_OutputFormat_name[int32(t.outputTransformer.GetFormat())]
		parsedResult, err := parser.Parse(outputFormat, respBody, t.outputTransformer.GetExtractionRules())
		if err != nil {
			return nil, fmt.Errorf("failed to parse output: %w", err)
		}

		if t.outputTransformer.GetTemplate() != "" {
			tpl, err := transformer.NewTemplate(t.outputTransformer.GetTemplate(), "{{", "}}")
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

	var result any
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
	cache             *configv1.CacheConfig
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
		cache:             callDefinition.GetCache(),
	}
}

// Tool returns the protobuf definition of the MCP tool.
func (t *MCPTool) Tool() *v1.Tool {
	return t.tool
}

// GetCacheConfig returns the cache configuration for the MCP tool.
func (t *MCPTool) GetCacheConfig() *configv1.CacheConfig {
	return t.cache
}

// Execute handles the execution of the MCP tool. It forwards the tool call,
// including its name and arguments, to the downstream MCP service using the
// configured client and applies any necessary transformations to the request
// and response.
func (t *MCPTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	// Use the tool name from the definition, as the request tool name might be sanitized/modified
	bareToolName := t.tool.GetName()

	var inputs map[string]any
	if err := json.Unmarshal(req.ToolInputs, &inputs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool inputs: %w", err)
	}

	var arguments json.RawMessage
	if t.inputTransformer != nil && t.inputTransformer.GetTemplate() != "" {
		tpl, err := transformer.NewTemplate(t.inputTransformer.GetTemplate(), "{{", "}}")
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
		if t.outputTransformer.GetFormat() == configv1.OutputTransformer_RAW_BYTES {
			return map[string]any{"raw": responseBytes}, nil
		}
		parser := transformer.NewTextParser()
		outputFormat := configv1.OutputTransformer_OutputFormat_name[int32(t.outputTransformer.GetFormat())]
		parsedResult, err := parser.Parse(outputFormat, responseBytes, t.outputTransformer.GetExtractionRules())
		if err != nil {
			return nil, fmt.Errorf("failed to parse output: %w", err)
		}

		if t.outputTransformer.GetTemplate() != "" {
			tpl, err := transformer.NewTemplate(t.outputTransformer.GetTemplate(), "{{", "}}")
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
	cache             *configv1.CacheConfig
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
		cache:             callDefinition.GetCache(),
	}
}

// Tool returns the protobuf definition of the OpenAPI tool.
func (t *OpenAPITool) Tool() *v1.Tool {
	return t.tool
}

// GetCacheConfig returns the cache configuration for the OpenAPI tool.
func (t *OpenAPITool) GetCacheConfig() *configv1.CacheConfig {
	return t.cache
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
			url = strings.ReplaceAll(url, "{{"+paramName+"}}", fmt.Sprintf("%v", paramValue))
			delete(inputs, paramName)
		}
	}

	var body io.Reader
	var contentType string
	if t.method == http.MethodPost || t.method == http.MethodPut {
		if t.inputTransformer != nil && t.inputTransformer.GetTemplate() != "" {
			tpl, err := transformer.NewTemplate(t.inputTransformer.GetTemplate(), "{{", "}}")
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
		if t.outputTransformer.GetFormat() == configv1.OutputTransformer_RAW_BYTES {
			return map[string]any{"raw": respBody}, nil
		}
		parser := transformer.NewTextParser()
		outputFormat := configv1.OutputTransformer_OutputFormat_name[int32(t.outputTransformer.GetFormat())]
		parsedResult, err := parser.Parse(outputFormat, respBody, t.outputTransformer.GetExtractionRules())
		if err != nil {
			return nil, fmt.Errorf("failed to parse output: %w", err)
		}

		if t.outputTransformer.GetTemplate() != "" {
			tpl, err := transformer.NewTemplate(t.outputTransformer.GetTemplate(), "{{", "}}")
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
	tool           *v1.Tool
	service        *configv1.CommandLineUpstreamService
	callDefinition *configv1.CommandLineCallDefinition
}

// NewCommandTool creates a new CommandTool.
//
// tool is the protobuf definition of the tool.
// command is the command to be executed.
func NewCommandTool(
	tool *v1.Tool,
	service *configv1.CommandLineUpstreamService,
	callDefinition *configv1.CommandLineCallDefinition,
) Tool {
	return &CommandTool{
		tool:           tool,
		service:        service,
		callDefinition: callDefinition,
	}
}

// LocalCommandTool implements the Tool interface for a tool that is executed as a
// local command-line process. It maps tool inputs to command-line arguments and
// environment variables.
type LocalCommandTool struct {
	tool           *v1.Tool
	service        *configv1.CommandLineUpstreamService
	callDefinition *configv1.CommandLineCallDefinition
}

// NewLocalCommandTool creates a new LocalCommandTool.
//
// tool is the protobuf definition of the tool.
// command is the command to be executed.
func NewLocalCommandTool(
	tool *v1.Tool,
	service *configv1.CommandLineUpstreamService,
	callDefinition *configv1.CommandLineCallDefinition,
) Tool {
	return &LocalCommandTool{
		tool:           tool,
		service:        service,
		callDefinition: callDefinition,
	}
}

// Tool returns the protobuf definition of the command-line tool.
func (t *LocalCommandTool) Tool() *v1.Tool {
	return t.tool
}

// GetCacheConfig returns the cache configuration for the command-line tool.
func (t *LocalCommandTool) GetCacheConfig() *configv1.CacheConfig {
	if t.callDefinition == nil {
		return nil
	}
	return t.callDefinition.GetCache()
}

// Execute handles the execution of the command-line tool. It constructs a command
// with arguments and environment variables derived from the tool inputs, runs
// the command, and returns its output.
func (t *LocalCommandTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	var inputs map[string]any
	if err := json.Unmarshal(req.ToolInputs, &inputs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool inputs: %w", err)
	}

	args := []string{}
	// Default to args from definition
	if t.callDefinition.GetArgs() != nil {
		args = append(args, t.callDefinition.GetArgs()...)
	}

	// If args are provided in the request, they override the definition's args.
	if inputs != nil {
		if argsVal, ok := inputs["args"]; ok {
			if argsList, ok := argsVal.([]any); ok {
				args = []string{} // Override
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
			delete(inputs, "args")
		}
	}

	timeout := t.service.GetTimeout()
	if timeout != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout.AsDuration())
		defer cancel()
	}

	executor := command.NewLocalExecutor()

	env := os.Environ()
	for key, value := range inputs {
		env = append(env, fmt.Sprintf("%s=%v", key, value))
	}

	for _, param := range t.callDefinition.GetParameters() {
		if secret := param.GetSecret(); secret != nil {
			secretValue, err := util.ResolveSecret(secret)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve secret for parameter %q: %w", param.GetSchema().GetName(), err)
			}
			env = append(env, fmt.Sprintf("%s=%s", param.GetSchema().GetName(), secretValue))
		}
	}

	startTime := time.Now()
	// Handle JSON communication protocol directly. This approach was chosen over a
	// separate JSONExecutor to keep the logic self-contained within the CommandTool,
	// simplifying the overall design and reducing dependencies between packages.
	if t.service.GetCommunicationProtocol() == configv1.CommandLineUpstreamService_COMMUNICATION_PROTOCOL_JSON {
		stdin, stdout, stderr, _, err := executor.ExecuteWithStdIO(ctx, t.service.GetCommand(), args, t.service.GetWorkingDirectory(), env)
		if err != nil {
			return nil, fmt.Errorf("failed to execute command with stdio: %w", err)
		}

		go func() {
			defer stderr.Close()
			io.Copy(io.Discard, stderr)
		}()

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer stdin.Close()
			json.NewEncoder(stdin).Encode(inputs)
		}()

		var result map[string]interface{}
		var decodeErr error
		go func() {
			defer wg.Done()
			decodeErr = json.NewDecoder(stdout).Decode(&result)
		}()

		wg.Wait()
		if decodeErr != nil {
			return nil, fmt.Errorf("failed to decode JSON from command output: %w", decodeErr)
		}
		return result, nil
	}

	var stdin io.Reader
	if stdinVal, ok := inputs["stdin"]; ok {
		if stdinStr, ok := stdinVal.(string); ok {
			stdin = strings.NewReader(stdinStr)
		}
		delete(inputs, "stdin")
	}

	stdout, stderr, exitCodeChan, err := executor.Execute(ctx, t.service.GetCommand(), args, t.service.GetWorkingDirectory(), env, stdin)
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	var stdoutBuf, stderrBuf, combinedBuf bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if closer, ok := stdout.(io.ReadCloser); ok {
			defer closer.Close()
		}
		io.Copy(io.MultiWriter(&stdoutBuf, &combinedBuf), stdout)
	}()
	go func() {
		defer wg.Done()
		if closer, ok := stderr.(io.ReadCloser); ok {
			defer closer.Close()
		}
		io.Copy(io.MultiWriter(&stderrBuf, &combinedBuf), stderr)
	}()

	wg.Wait()
	exitCode := <-exitCodeChan
	endTime := time.Now()

	status := consts.CommandStatusSuccess
	if ctx.Err() == context.DeadlineExceeded {
		status = consts.CommandStatusTimeout
		exitCode = -1 // Override exit code on timeout
	} else if exitCode != 0 {
		status = consts.CommandStatusError
	}

	result := map[string]interface{}{
		"command":         t.service.GetCommand(),
		"args":            args,
		"stdout":          stdoutBuf.String(),
		"stderr":          stderrBuf.String(),
		"combined_output": combinedBuf.String(),
		"start_time":      startTime,
		"end_time":        endTime,
		"return_code":     exitCode,
		"status":          status,
	}

	return result, nil
}

// Tool returns the protobuf definition of the command-line tool.
func (t *CommandTool) Tool() *v1.Tool {
	return t.tool
}

// GetCacheConfig returns the cache configuration for the command-line tool.
func (t *CommandTool) GetCacheConfig() *configv1.CacheConfig {
	if t.callDefinition == nil {
		return nil
	}
	return t.callDefinition.GetCache()
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
	// Default to args from definition
	if t.callDefinition.GetArgs() != nil {
		args = append(args, t.callDefinition.GetArgs()...)
	}

	// If args are provided in the request, they override the definition's args.
	if inputs != nil {
		if argsVal, ok := inputs["args"]; ok {
			if argsList, ok := argsVal.([]any); ok {
				args = []string{} // Override
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
			delete(inputs, "args")
		}
	}

	timeout := t.service.GetTimeout()
	if timeout != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout.AsDuration())
		defer cancel()
	}

	executor := command.NewExecutor(t.service.GetContainerEnvironment())

	env := os.Environ()
	for key, value := range inputs {
		env = append(env, fmt.Sprintf("%s=%v", key, value))
	}

	for _, param := range t.callDefinition.GetParameters() {
		if secret := param.GetSecret(); secret != nil {
			secretValue, err := util.ResolveSecret(secret)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve secret for parameter %q: %w", param.GetSchema().GetName(), err)
			}
			env = append(env, fmt.Sprintf("%s=%s", param.GetSchema().GetName(), secretValue))
		}
	}

	startTime := time.Now()
	// Handle JSON communication protocol directly. This approach was chosen over a
	// separate JSONExecutor to keep the logic self-contained within the CommandTool,
	// simplifying the overall design and reducing dependencies between packages.
	if t.service.GetCommunicationProtocol() == configv1.CommandLineUpstreamService_COMMUNICATION_PROTOCOL_JSON {
		stdin, stdout, stderr, _, err := executor.ExecuteWithStdIO(ctx, t.service.GetCommand(), args, t.service.GetWorkingDirectory(), env)
		if err != nil {
			return nil, fmt.Errorf("failed to execute command with stdio: %w", err)
		}

		go func() {
			defer stderr.Close()
			io.Copy(io.Discard, stderr)
		}()

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer stdin.Close()
			json.NewEncoder(stdin).Encode(inputs)
		}()

		var result map[string]interface{}
		var decodeErr error
		go func() {
			defer wg.Done()
			decodeErr = json.NewDecoder(stdout).Decode(&result)
		}()

		wg.Wait()
		if decodeErr != nil {
			return nil, fmt.Errorf("failed to decode JSON from command output: %w", decodeErr)
		}
		return result, nil
	}

	var stdin io.Reader
	if stdinVal, ok := inputs["stdin"]; ok {
		if stdinStr, ok := stdinVal.(string); ok {
			stdin = strings.NewReader(stdinStr)
		}
		delete(inputs, "stdin")
	}

	stdout, stderr, exitCodeChan, err := executor.Execute(ctx, t.service.GetCommand(), args, t.service.GetWorkingDirectory(), env, stdin)
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	var stdoutBuf, stderrBuf, combinedBuf bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if closer, ok := stdout.(io.ReadCloser); ok {
			defer closer.Close()
		}
		io.Copy(io.MultiWriter(&stdoutBuf, &combinedBuf), stdout)
	}()
	go func() {
		defer wg.Done()
		if closer, ok := stderr.(io.ReadCloser); ok {
			defer closer.Close()
		}
		io.Copy(io.MultiWriter(&stderrBuf, &combinedBuf), stderr)
	}()

	wg.Wait()
	exitCode := <-exitCodeChan
	endTime := time.Now()

	status := consts.CommandStatusSuccess
	if ctx.Err() == context.DeadlineExceeded {
		status = consts.CommandStatusTimeout
		exitCode = -1 // Override exit code on timeout
	} else if exitCode != 0 {
		status = consts.CommandStatusError
	}

	result := map[string]interface{}{
		"command":         t.service.GetCommand(),
		"args":            args,
		"stdout":          stdoutBuf.String(),
		"stderr":          stderrBuf.String(),
		"combined_output": combinedBuf.String(),
		"start_time":      startTime,
		"end_time":        endTime,
		"return_code":     exitCode,
		"status":          status,
	}

	return result, nil
}
