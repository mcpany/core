// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
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

const (
	contentTypeJSON     = "application/json"
	redactedPlaceholder = "[REDACTED]"
)

// Tool is the fundamental interface for any executable tool in the system.
// Each implementation represents a different type of underlying service
// (e.g., gRPC, HTTP, command-line).
type Tool interface {
	// Tool returns the protobuf definition of the tool.
	Tool() *v1.Tool
	// MCPTool returns the MCP tool definition.
	MCPTool() *mcp.Tool
	// Execute runs the tool with the provided context and request, returning
	// the result or an error.
	Execute(ctx context.Context, req *ExecutionRequest) (any, error)
	// GetCacheConfig returns the cache configuration for the tool.
	GetCacheConfig() *configv1.CacheConfig
}

// ServiceInfo holds metadata about a registered upstream service, including its
// configuration and any associated protobuf file descriptors.
type ServiceInfo struct {
	// Name is the unique name of the service.
	Name string
	// Config is the configuration of the upstream service.
	Config *configv1.UpstreamServiceConfig
	// Fds is the FileDescriptorSet associated with the service (for gRPC/protobuf).
	Fds *descriptorpb.FileDescriptorSet

	// PreHooks are the cached pre-call hooks for the service.
	PreHooks []PreCallHook
	// PostHooks are the cached post-call hooks for the service.
	PostHooks []PostCallHook
}

// ExecutionRequest represents a request to execute a specific tool, including
// its name and input arguments as a raw JSON message.
type ExecutionRequest struct {
	// ToolName is the name of the tool to be executed.
	ToolName string
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
	// GetTool retrieves a tool by name.
	GetTool(toolName string) (Tool, bool)
	// GetServiceInfo retrieves metadata for a service.
	GetServiceInfo(serviceID string) (*ServiceInfo, bool)
}

// ExecutionFunc represents the next middleware in the chain.
type ExecutionFunc func(ctx context.Context, req *ExecutionRequest) (any, error)

type contextKey string

const toolContextKey = contextKey("tool")

// NewContextWithTool creates a new context with the given tool.
// ctx is the context.
// Returns the result.
func NewContextWithTool(ctx context.Context, t Tool) context.Context {
	return context.WithValue(ctx, toolContextKey, t)
}

// GetFromContext retrieves a tool from the context.
// ctx is the context.
// Returns the result, the result.
func GetFromContext(ctx context.Context) (Tool, bool) {
	t, ok := ctx.Value(toolContextKey).(Tool)
	return t, ok
}

// Callable is an interface that represents a callable tool.
type Callable interface {
	// Call executes the callable with the given request.
	Call(ctx context.Context, req *ExecutionRequest) (any, error)
}

// Action defines the decision made by a pre-call hook.
type Action int

const (
	// ActionAllow indicates that the action is allowed.
	ActionAllow Action = 0
	// ActionDeny indicates that the action is denied.
	ActionDeny Action = 1
	// ActionSaveCache indicates that the result should be cached.
	ActionSaveCache Action = 2
	// ActionDeleteCache indicates that the cache should be invalidated.
	ActionDeleteCache Action = 3
)

// CacheControl is a mutable struct to pass cache control instructions via context.
type CacheControl struct {
	Action Action
}

const cacheControlContextKey = contextKey("cache_control")

// NewContextWithCacheControl creates a new context with the given CacheControl.
// ctx is the context.
// cc is the cc.
// Returns the result.
func NewContextWithCacheControl(ctx context.Context, cc *CacheControl) context.Context {
	return context.WithValue(ctx, cacheControlContextKey, cc)
}

// GetCacheControl retrieves the CacheControl from the context.
// ctx is the context.
// Returns the result, the result.
func GetCacheControl(ctx context.Context) (*CacheControl, bool) {
	cc, ok := ctx.Value(cacheControlContextKey).(*CacheControl)
	return cc, ok
}

// PreCallHook defines the interface for hooks executed before a tool call.
type PreCallHook interface {
	// ExecutePre runs the hook. It returns an action (Allow/Deny),
	// a potentially modified request (or nil if unchanged), and an error.
	ExecutePre(ctx context.Context, req *ExecutionRequest) (Action, *ExecutionRequest, error)
}

// PostCallHook defines the interface for hooks executed after a tool call.
type PostCallHook interface {
	// ExecutePost runs the hook. It returns the potentially modified result
	// (or original if unchanged) and an error.
	ExecutePost(ctx context.Context, req *ExecutionRequest, result any) (any, error)
}

// GRPCTool implements the Tool interface for a tool that is exposed via a gRPC
// endpoint. It handles the marshalling of JSON inputs to protobuf messages and
// invoking the gRPC method.
type GRPCTool struct {
	tool           *v1.Tool
	mcpTool        *mcp.Tool
	mcpToolOnce    sync.Once
	poolManager    *pool.Manager
	serviceID      string
	method         protoreflect.MethodDescriptor
	requestMessage protoreflect.ProtoMessage
	cache          *configv1.CacheConfig
}

// NewGRPCTool creates a new GRPCTool.
//
// Parameters:
//
//	tool: The protobuf definition of the tool.
//	poolManager: The connection pool manager.
//	serviceID: The identifier for the service.
//	method: The gRPC method descriptor.
//	callDefinition: The configuration for the gRPC call.
//
// Returns:
//
//	*GRPCTool: The created GRPCTool.
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

// MCPTool returns the MCP tool definition.
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
func (t *GRPCTool) GetCacheConfig() *configv1.CacheConfig {
	return t.cache
}

// Execute handles the execution of the gRPC tool. It retrieves a client from the
// pool, unmarshals the JSON input into a protobuf request message, invokes the
// gRPC method, and marshals the protobuf response back to JSON.
// ctx is the context.
// Returns the result, an error.
func (t *GRPCTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	if logging.GetLogger().Enabled(ctx, slog.LevelDebug) {
		logging.GetLogger().Debug("executing tool", "tool", req.ToolName, "inputs", prettyPrint(req.ToolInputs, contentTypeJSON))
	}
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
	mcpTool           *mcp.Tool
	mcpToolOnce       sync.Once
	poolManager       *pool.Manager
	serviceID         string
	authenticator     auth.UpstreamAuthenticator
	parameters        []*configv1.HttpParameterMapping
	inputTransformer  *configv1.InputTransformer
	outputTransformer *configv1.OutputTransformer
	webhookClient     *WebhookClient
	cache             *configv1.CacheConfig
	resilienceManager *resilience.Manager
	policies          []*CompiledCallPolicy
	callID            string

	// Cached fields for performance
	initError          error
	cachedMethod       string
	cachedURL          *url.URL
	cachedPath         string // with %7B replaced
	cachedQuery        string // with %7B replaced
	cachedPlaceholders map[string]string
	paramInPath        []bool
	paramInQuery       []bool
}

// NewHTTPTool creates a new HTTPTool.
//
// Parameters:
//
//	tool: The protobuf definition of the tool.
//	poolManager: The connection pool manager.
//	serviceID: The identifier for the service.
//	authenticator: The authenticator for upstream requests.
//	callDefinition: The configuration for the HTTP call.
//	cfg: The resilience configuration.
//	policies: The security policies for the call.
//	callID: The unique identifier for the call.
//
// Returns:
//
//	*HTTPTool: The created HTTPTool.
func NewHTTPTool(tool *v1.Tool, poolManager *pool.Manager, serviceID string, authenticator auth.UpstreamAuthenticator, callDefinition *configv1.HttpCallDefinition, cfg *configv1.ResilienceConfig, policies []*configv1.CallPolicy, callID string) *HTTPTool {
	var webhookClient *WebhookClient
	if it := callDefinition.GetInputTransformer(); it != nil && it.GetWebhook() != nil {
		webhookClient = NewWebhookClient(it.GetWebhook())
	}
	t := &HTTPTool{
		tool:              tool,
		poolManager:       poolManager,
		serviceID:         serviceID,
		authenticator:     authenticator,
		parameters:        callDefinition.GetParameters(),
		inputTransformer:  callDefinition.GetInputTransformer(),
		outputTransformer: callDefinition.GetOutputTransformer(),
		webhookClient:     webhookClient,
		cache:             callDefinition.GetCache(),
		resilienceManager: resilience.NewManager(cfg),
		callID:            callID,
	}

	compiled, err := CompileCallPolicies(policies)
	if err != nil {
		t.initError = fmt.Errorf("failed to compile call policies: %w", err)
	}
	t.policies = compiled

	// Pre-calculate URL components
	methodAndURL := strings.Fields(tool.GetUnderlyingMethodFqn())
	if len(methodAndURL) != 2 {
		t.initError = fmt.Errorf("invalid http tool definition: expected method and URL, got %q", tool.GetUnderlyingMethodFqn())
		return t
	}
	t.cachedMethod = methodAndURL[0]
	rawURL := methodAndURL[1]

	u, err := url.Parse(rawURL)
	if err != nil {
		t.initError = fmt.Errorf("failed to parse url: %w", err)
		return t
	}
	t.cachedURL = u

	pathStr := u.EscapedPath()
	pathStr = strings.ReplaceAll(pathStr, "%7B", "{")
	pathStr = strings.ReplaceAll(pathStr, "%7D", "}")
	t.cachedPath = pathStr

	queryStr := u.RawQuery
	queryStr = strings.ReplaceAll(queryStr, "%7B", "{")
	queryStr = strings.ReplaceAll(queryStr, "%7D", "}")
	t.cachedQuery = queryStr

	t.cachedPlaceholders = make(map[string]string)
	t.paramInPath = make([]bool, len(callDefinition.GetParameters()))
	t.paramInQuery = make([]bool, len(callDefinition.GetParameters()))

	for i, param := range callDefinition.GetParameters() {
		if schema := param.GetSchema(); schema != nil {
			name := schema.GetName()
			placeholder := "{{" + name + "}}"
			t.cachedPlaceholders[name] = placeholder

			if strings.Contains(t.cachedPath, placeholder) {
				t.paramInPath[i] = true
			}
			if strings.Contains(t.cachedQuery, placeholder) {
				t.paramInQuery[i] = true
			}
		}
	}

	return t
}

// Tool returns the protobuf definition of the HTTP tool.
func (t *HTTPTool) Tool() *v1.Tool {
	return t.tool
}

// MCPTool returns the MCP tool definition.
func (t *HTTPTool) MCPTool() *mcp.Tool {
	t.mcpToolOnce.Do(func() {
		var err error
		t.mcpTool, err = ConvertProtoToMCPTool(t.tool)
		if err != nil {
			logging.GetLogger().Error("Failed to convert tool to MCP tool", "toolName", t.tool.GetName(), "error", err)
		}
	})
	return t.mcpTool
}

// GetCacheConfig returns the cache configuration for the HTTP tool.
func (t *HTTPTool) GetCacheConfig() *configv1.CacheConfig {
	return t.cache
}

// Execute handles the execution of the HTTP tool. It builds an HTTP request by
// mapping input parameters to the path, query, and body, applies any
// configured transformations, sends the request, and processes the response.
//
// ctx is the context.
// Returns the result, an error.
//
//nolint:gocyclo
func (t *HTTPTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	if logging.GetLogger().Enabled(ctx, slog.LevelDebug) {
		logging.GetLogger().Debug("executing tool", "tool", req.ToolName, "inputs", prettyPrint(req.ToolInputs, contentTypeJSON))
	}
	defer metrics.MeasureSince([]string{"http", "request", "latency"}, time.Now())

	if allowed, err := EvaluateCompiledCallPolicy(t.policies, t.tool.GetName(), t.callID, req.ToolInputs); err != nil {
		return nil, fmt.Errorf("failed to evaluate call policy: %w", err)
	} else if !allowed {
		return nil, fmt.Errorf("tool execution blocked by policy")
	}

	httpPool, ok := pool.Get[*client.HTTPClientWrapper](t.poolManager, t.serviceID)
	if !ok {
		metrics.IncrCounter([]string{"http", "request", "error"}, 1)
		return nil, fmt.Errorf("no http pool found for service: %s", t.serviceID)
	}

	httpClient, err := httpPool.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get client from pool: %w", err)
	}
	defer httpPool.Put(httpClient)

	if t.initError != nil {
		return nil, t.initError
	}

	method := t.cachedMethod
	u := t.cachedURL
	pathStr := t.cachedPath
	queryStr := t.cachedQuery

	var inputs map[string]any
	if len(req.ToolInputs) > 0 {
		decoder := json.NewDecoder(bytes.NewReader(req.ToolInputs))
		decoder.UseNumber()
		if err := decoder.Decode(&inputs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tool inputs: %w", err)
		}
	}

	for i, param := range t.parameters {
		if secret := param.GetSecret(); secret != nil {
			secretValue, err := util.ResolveSecret(ctx, secret)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve secret for parameter %q: %w", param.GetSchema().GetName(), err)
			}
			placeholder := t.cachedPlaceholders[param.GetSchema().GetName()]
			if t.paramInPath[i] {
				pathStr = strings.ReplaceAll(pathStr, placeholder, secretValue)
			}
			if t.paramInQuery[i] {
				queryStr = strings.ReplaceAll(queryStr, placeholder, secretValue)
			}
		} else if schema := param.GetSchema(); schema != nil {
			placeholder := t.cachedPlaceholders[schema.GetName()]
			if val, ok := inputs[schema.GetName()]; ok {
				var valStr string
				// Optimization: avoid fmt.Sprintf for strings
				switch v := val.(type) {
				case string:
					valStr = v
				default:
					valStr = fmt.Sprintf("%v", v)
				}

				if param.GetDisableEscape() {
					// Check for path traversal if in path
					if t.paramInPath[i] {
						// Check the raw value first
						if err := checkForPathTraversal(valStr); err != nil {
							return nil, fmt.Errorf("path traversal attempt detected in parameter %q: %w", schema.GetName(), err)
						}

						// Check the decoded value to prevent bypasses like %2e%2e
						if decodedVal, err := url.QueryUnescape(valStr); err == nil {
							if err := checkForPathTraversal(decodedVal); err != nil {
								return nil, fmt.Errorf("path traversal attempt detected in parameter %q (decoded): %w", schema.GetName(), err)
							}
						}

						pathStr = strings.ReplaceAll(pathStr, placeholder, valStr)
					}
					if t.paramInQuery[i] {
						queryStr = strings.ReplaceAll(queryStr, placeholder, valStr)
					}
				} else {
					if t.paramInPath[i] {
						pathStr = strings.ReplaceAll(pathStr, placeholder, url.PathEscape(valStr))
					}
					if t.paramInQuery[i] {
						queryStr = strings.ReplaceAll(queryStr, placeholder, url.QueryEscape(valStr))
					}
				}
				delete(inputs, schema.GetName())
			} else {
				if t.paramInPath[i] {
					pathStr = strings.ReplaceAll(pathStr, "/"+placeholder, "")
				}
				if t.paramInQuery[i] {
					queryStr = strings.ReplaceAll(queryStr, "/"+placeholder, "")
				}
			}
		}
	}

	// Clean the path to resolve . and .. and //
	// We do this on the encoded string to treat %2F as opaque characters
	// This prevents path.Clean from treating encoded slashes as separators
	// and messing up the re-encoding later (which would convert %2F to /).
	hadTrailingSlash := strings.HasSuffix(pathStr, "/")
	pathStr = path.Clean(pathStr)
	if hadTrailingSlash && pathStr != "/" {
		pathStr += "/"
	}

	// Reconstruct URL string manually to avoid re-encoding
	var buf strings.Builder
	buf.WriteString(u.Scheme)
	buf.WriteString("://")
	if u.User != nil {
		buf.WriteString(u.User.String())
		buf.WriteString("@")
	}
	buf.WriteString(u.Host)
	if pathStr != "" && !strings.HasPrefix(pathStr, "/") {
		buf.WriteString("/")
	}
	buf.WriteString(pathStr)
	if queryStr != "" {
		buf.WriteString("?")
		buf.WriteString(queryStr)
	}
	urlString := buf.String()

	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}
	urlString = parsedURL.String()

	var body io.Reader
	var contentType string
	if inputs != nil {
		if method == http.MethodPost || method == http.MethodPut {
			switch {
			case t.webhookClient != nil:
				// Use webhook for transformation
				data := map[string]any{
					"kind":      configv1.WebhookKind_WEBHOOK_KIND_TRANSFORM_INPUT,
					"tool_name": req.ToolName,
					"inputs":    inputs,
				}
				respEvent, err := t.webhookClient.Call(ctx, "com.mcpany.tool.transform_input", data)
				if err != nil {
					return nil, fmt.Errorf("transformation webhook failed: %w", err)
				}
				// We expect the data to be the transformed body
				respData := respEvent.Data()
				if len(respData) > 0 {
					body = bytes.NewReader(respData)
					// Verify if it looks like JSON?
					if json.Valid(respData) {
						contentType = contentTypeJSON
					}
				}
			case t.inputTransformer != nil && t.inputTransformer.GetTemplate() != "": //nolint:staticcheck
				tpl, err := transformer.NewTemplate(t.inputTransformer.GetTemplate(), "{{", "}}") //nolint:staticcheck
				if err != nil {
					return nil, fmt.Errorf("failed to create input template: %w", err)
				}
				renderedBody, err := tpl.Render(inputs)
				if err != nil {
					return nil, fmt.Errorf("failed to render input template: %w", err)
				}
				body = strings.NewReader(renderedBody)
			default:
				jsonBytes, err := json.Marshal(inputs)
				if err != nil {
					return "", fmt.Errorf("failed to marshal tool inputs to json: %w", err)
				}
				body = bytes.NewReader(jsonBytes)
				contentType = contentTypeJSON
			}
		}
	}

	var resp *http.Response
	work := func() error {
		var bodyForAttempt io.Reader
		if body != nil {
			if seeker, ok := body.(io.Seeker); ok {
				if _, err := seeker.Seek(0, io.SeekStart); err != nil {
					return &resilience.PermanentError{Err: fmt.Errorf("failed to seek body: %w", err)}
				}
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
		// httpReq.Header.Set("User-Agent", "mcpany-e2e-test")
		httpReq.Header.Set("Accept", "*/*")

		if t.authenticator != nil {
			if err := t.authenticator.Authenticate(httpReq); err != nil {
				return &resilience.PermanentError{Err: fmt.Errorf("failed to authenticate request: %w", err)}
			}
			logging.GetLogger().Debug("Applied authentication", "user_agent", httpReq.Header.Get("User-Agent"))
		} else {
			logging.GetLogger().Debug("No authenticator configured")
		}

		if method == http.MethodGet || method == http.MethodDelete {
			q := httpReq.URL.Query()
			for key, value := range inputs {
				q.Add(key, fmt.Sprintf("%v", value))
			}
			httpReq.URL.RawQuery = q.Encode()
		}

		if logging.GetLogger().Enabled(ctx, slog.LevelDebug) {
			// Log headers
			var headerBuf bytes.Buffer
			headerBuf.WriteString(fmt.Sprintf("%s %s %s\n", httpReq.Method, httpReq.URL.Path, httpReq.Proto))
			headerBuf.WriteString(fmt.Sprintf("Host: %s\n", httpReq.Host))
			for k, v := range httpReq.Header {
				val := strings.Join(v, ", ")
				if isSensitiveHeader(k) {
					val = redactedPlaceholder
				}
				headerBuf.WriteString(fmt.Sprintf("%s: %s\n", k, val))
			}
			logging.GetLogger().DebugContext(ctx, "sending http request headers", "headers", headerBuf.String())

			// Log body
			if bodyForAttempt != nil {
				contentType := httpReq.Header.Get("Content-Type")
				bodyBytes, _ := io.ReadAll(bodyForAttempt)
				// Restore body
				if seeker, ok := bodyForAttempt.(io.Seeker); ok {
					_, _ = seeker.Seek(0, io.SeekStart)
				}
				logging.GetLogger().DebugContext(ctx, "sending http request body", "body", prettyPrint(bodyBytes, contentType))
			}
		}

		attemptResp, err := httpClient.Do(httpReq)
		if err != nil {
			return fmt.Errorf("failed to execute http request: %w", err)
		}

		if attemptResp.StatusCode >= 400 && attemptResp.StatusCode < 500 {
			bodyBytes, _ := io.ReadAll(attemptResp.Body)
			_ = attemptResp.Body.Close()
			logging.GetLogger().DebugContext(ctx, "Upstream HTTP 4xx", "status", attemptResp.StatusCode, "body", string(bodyBytes), "url", httpReq.URL.String())
			return &resilience.PermanentError{Err: fmt.Errorf("upstream HTTP request failed with status %d", attemptResp.StatusCode)}
		}

		if attemptResp.StatusCode >= 500 {
			_ = attemptResp.Body.Close()
			return fmt.Errorf("upstream HTTP request failed with status %d", attemptResp.StatusCode)
		}

		resp = attemptResp
		return nil
	}

	if err := t.resilienceManager.Execute(work); err != nil {
		metrics.IncrCounter([]string{"http", "request", "error"}, 1)
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()
	metrics.IncrCounter([]string{"http", "request", "success"}, 1)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read http response body: %w", err)
	}

	if logging.GetLogger().Enabled(ctx, slog.LevelDebug) {
		// Log headers
		var headerBuf bytes.Buffer
		headerBuf.WriteString(fmt.Sprintf("%s %s\n", resp.Proto, resp.Status))
		for k, v := range resp.Header {
			val := strings.Join(v, ", ")
			if isSensitiveHeader(k) {
				val = redactedPlaceholder
			}
			headerBuf.WriteString(fmt.Sprintf("%s: %s\n", k, val))
		}
		logging.GetLogger().DebugContext(ctx, "received http response headers", "headers", headerBuf.String())

		// Log body
		contentType := resp.Header.Get("Content-Type")
		logging.GetLogger().DebugContext(ctx, "received http response body", "body", prettyPrint(respBody, contentType))

		// Restore body for subsequent processing
		resp.Body = io.NopCloser(bytes.NewBuffer(respBody))
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
			tpl, err := transformer.NewTemplate(t.outputTransformer.GetTemplate(), "{{", "}}")
			if err != nil {
				return nil, fmt.Errorf("failed to create output template: %w", err)
			}
			resultMap, ok := parsedResult.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("output must be a map to be used with a template, got %T", parsedResult)
			}
			renderedOutput, err := tpl.Render(resultMap)
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
	mcpTool           *mcp.Tool
	mcpToolOnce       sync.Once
	client            client.MCPClient
	inputTransformer  *configv1.InputTransformer
	outputTransformer *configv1.OutputTransformer
	webhookClient     *WebhookClient
	cache             *configv1.CacheConfig
}

// NewMCPTool creates a new MCPTool.
//
// Parameters:
//
//	tool: The protobuf definition of the tool.
//	client: The MCP client for downstream communication.
//	callDefinition: The configuration for the MCP call.
//
// Returns:
//
//	*MCPTool: The created MCPTool.
func NewMCPTool(tool *v1.Tool, client client.MCPClient, callDefinition *configv1.MCPCallDefinition) *MCPTool {
	var webhookClient *WebhookClient
	if it := callDefinition.GetInputTransformer(); it != nil && it.GetWebhook() != nil {
		webhookClient = NewWebhookClient(it.GetWebhook())
	}
	return &MCPTool{
		tool:              tool,
		client:            client,
		inputTransformer:  callDefinition.GetInputTransformer(),
		outputTransformer: callDefinition.GetOutputTransformer(),
		webhookClient:     webhookClient,
		cache:             callDefinition.GetCache(),
	}
}

// Tool returns the protobuf definition of the MCP tool.
func (t *MCPTool) Tool() *v1.Tool {
	return t.tool
}

// MCPTool returns the MCP tool definition.
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
func (t *MCPTool) GetCacheConfig() *configv1.CacheConfig {
	return t.cache
}

// Execute handles the execution of the MCP tool. It forwards the tool call,
// including its name and arguments, to the downstream MCP service using the
// configured client and applies any necessary transformations to the request
// and response.
// ctx is the context.
// Returns the result, an error.
func (t *MCPTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	if logging.GetLogger().Enabled(ctx, slog.LevelDebug) {
		logging.GetLogger().Debug("executing tool", "tool", req.ToolName, "inputs", prettyPrint(req.ToolInputs, contentTypeJSON))
	}
	// Use the tool name from the definition, as the request tool name might be sanitized/modified
	bareToolName := t.tool.GetName()

	var inputs map[string]any
	decoder := json.NewDecoder(bytes.NewReader(req.ToolInputs))
	decoder.UseNumber()
	if err := decoder.Decode(&inputs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool inputs: %w", err)
	}

	var arguments json.RawMessage
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
			arguments = json.RawMessage(respData)
		}
	case t.inputTransformer != nil && t.inputTransformer.GetTemplate() != "": //nolint:staticcheck
		tpl, err := transformer.NewTemplate(t.inputTransformer.GetTemplate(), "{{", "}}") //nolint:staticcheck
		if err != nil {
			return nil, fmt.Errorf("failed to create input template: %w", err)
		}
		rendered, err := tpl.Render(inputs)
		if err != nil {
			return nil, fmt.Errorf("failed to render input template: %w", err)
		}
		arguments = []byte(rendered)
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
		parsedResult, err := parser.Parse(outputFormat, responseBytes, t.outputTransformer.GetExtractionRules(), t.outputTransformer.GetJqQuery())
		if err != nil {
			return nil, fmt.Errorf("failed to parse output: %w", err)
		}

		if t.outputTransformer.GetTemplate() != "" {
			tpl, err := transformer.NewTemplate(t.outputTransformer.GetTemplate(), "{{", "}}")
			if err != nil {
				return nil, fmt.Errorf("failed to create output template: %w", err)
			}
			resultMap, ok := parsedResult.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("output must be a map to be used with a template, got %T", parsedResult)
			}
			renderedOutput, err := tpl.Render(resultMap)
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
	mcpTool           *mcp.Tool
	mcpToolOnce       sync.Once
	client            client.HTTPClient
	parameterDefs     map[string]string
	method            string
	url               string
	authenticator     auth.UpstreamAuthenticator
	inputTransformer  *configv1.InputTransformer
	outputTransformer *configv1.OutputTransformer
	webhookClient     *WebhookClient
	cache             *configv1.CacheConfig
}

// NewOpenAPITool creates a new OpenAPITool.
//
// Parameters:
//
//	tool: The protobuf definition of the tool.
//	client: The HTTP client for requests.
//	parameterDefs: Mapping of parameter names to their locations (path, query, etc.).
//	method: The HTTP method.
//	url: The URL template.
//	authenticator: The authenticator for requests.
//	callDefinition: The configuration for the OpenAPI call.
//
// Returns:
//
//	*OpenAPITool: The created OpenAPITool.
func NewOpenAPITool(tool *v1.Tool, client client.HTTPClient, parameterDefs map[string]string, method, url string, authenticator auth.UpstreamAuthenticator, callDefinition *configv1.OpenAPICallDefinition) *OpenAPITool {
	var webhookClient *WebhookClient
	if it := callDefinition.GetInputTransformer(); it != nil && it.GetWebhook() != nil {
		webhookClient = NewWebhookClient(it.GetWebhook())
	}
	return &OpenAPITool{
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
}

// Tool returns the protobuf definition of the OpenAPI tool.
func (t *OpenAPITool) Tool() *v1.Tool {
	return t.tool
}

// MCPTool returns the MCP tool definition.
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
func (t *OpenAPITool) GetCacheConfig() *configv1.CacheConfig {
	return t.cache
}

// Execute handles the execution of the OpenAPI tool. It constructs an HTTP
// request based on the operation's method, URL, and parameter definitions,
// sends the request, and processes the response, applying transformations as
// needed.
// ctx is the context.
// Returns the result, an error.
func (t *OpenAPITool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) { //nolint:gocyclo
	if logging.GetLogger().Enabled(ctx, slog.LevelDebug) {
		logging.GetLogger().Debug("executing tool", "tool", req.ToolName, "inputs", prettyPrint(req.ToolInputs, contentTypeJSON))
	}
	var inputs map[string]any
	decoder := json.NewDecoder(bytes.NewReader(req.ToolInputs))
	decoder.UseNumber()
	if err := decoder.Decode(&inputs); err != nil {
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
				if json.Valid(respData) {
					contentType = contentTypeJSON
				}
			}
		case t.inputTransformer != nil && t.inputTransformer.GetTemplate() != "": //nolint:staticcheck
			tpl, err := transformer.NewTemplate(t.inputTransformer.GetTemplate(), "{{", "}}") //nolint:staticcheck
			if err != nil {
				return nil, fmt.Errorf("failed to create input template: %w", err)
			}
			renderedBody, err := tpl.Render(inputs)
			if err != nil {
				return nil, fmt.Errorf("failed to render input template: %w", err)
			}
			body = strings.NewReader(renderedBody)
		default:
			jsonBytes, err := json.Marshal(inputs)
			if err != nil {
				return "", fmt.Errorf("failed to marshal tool inputs to json: %w", err)
			}
			body = bytes.NewReader(jsonBytes)
			contentType = contentTypeJSON
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
				if slice, ok := paramValue.([]interface{}); ok {
					for _, v := range slice {
						q.Add(paramName, fmt.Sprintf("%v", v))
					}
				} else {
					q.Add(paramName, fmt.Sprintf("%v", paramValue))
				}
			}
		}
		httpReq.URL.RawQuery = q.Encode()
	}

	resp, err := t.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute http request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

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
		parsedResult, err := parser.Parse(outputFormat, respBody, t.outputTransformer.GetExtractionRules(), t.outputTransformer.GetJqQuery())
		if err != nil {
			return nil, fmt.Errorf("failed to parse output: %w", err)
		}

		if t.outputTransformer.GetTemplate() != "" {
			tpl, err := transformer.NewTemplate(t.outputTransformer.GetTemplate(), "{{", "}}")
			if err != nil {
				return nil, fmt.Errorf("failed to create output template: %w", err)
			}
			resultMap, ok := parsedResult.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("output must be a map to be used with a template, got %T", parsedResult)
			}
			renderedOutput, err := tpl.Render(resultMap)
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
	tool            *v1.Tool
	mcpTool         *mcp.Tool
	mcpToolOnce     sync.Once
	service         *configv1.CommandLineUpstreamService
	callDefinition  *configv1.CommandLineCallDefinition
	executorFactory func(*configv1.ContainerEnvironment) command.Executor
	policies        []*CompiledCallPolicy
	callID          string
	initError       error
}

// NewCommandTool creates a new CommandTool.
//
// Parameters:
//
//	tool: The protobuf definition of the tool.
//	service: The configuration of the command-line service.
//	callDefinition: The configuration for the command-line call.
//	policies: The security policies.
//	callID: The unique identifier for the call.
//
// Returns:
//
//	Tool: The created CommandTool.
func NewCommandTool(
	tool *v1.Tool,
	service *configv1.CommandLineUpstreamService,
	callDefinition *configv1.CommandLineCallDefinition,
	policies []*configv1.CallPolicy,
	callID string,
) Tool {
	compiled, err := CompileCallPolicies(policies)
	t := &CommandTool{
		tool:           tool,
		service:        service,
		callDefinition: callDefinition,
		policies:       compiled,
		callID:         callID,
	}
	if err != nil {
		t.initError = fmt.Errorf("failed to compile call policies: %w", err)
	}
	return t
}

// LocalCommandTool implements the Tool interface for a tool that is executed as a
// local command-line process. It maps tool inputs to command-line arguments and
// environment variables.
type LocalCommandTool struct {
	tool           *v1.Tool
	mcpTool        *mcp.Tool
	mcpToolOnce    sync.Once
	service        *configv1.CommandLineUpstreamService
	callDefinition *configv1.CommandLineCallDefinition
	policies       []*CompiledCallPolicy
	callID         string
	initError      error
}

// NewLocalCommandTool creates a new LocalCommandTool.
//
// Parameters:
//
//	tool: The protobuf definition of the tool.
//	service: The configuration of the command-line service.
//	callDefinition: The configuration for the command-line call.
//	policies: The security policies.
//	callID: The unique identifier for the call.
//
// Returns:
//
//	Tool: The created LocalCommandTool.
func NewLocalCommandTool(
	tool *v1.Tool,
	service *configv1.CommandLineUpstreamService,
	callDefinition *configv1.CommandLineCallDefinition,
	policies []*configv1.CallPolicy,
	callID string,
) Tool {
	compiled, err := CompileCallPolicies(policies)
	t := &LocalCommandTool{
		tool:           tool,
		service:        service,
		callDefinition: callDefinition,
		policies:       compiled,
		callID:         callID,
	}
	if err != nil {
		t.initError = fmt.Errorf("failed to compile call policies: %w", err)
	}
	return t
}

// Tool returns the protobuf definition of the command-line tool.
func (t *LocalCommandTool) Tool() *v1.Tool {
	return t.tool
}

// MCPTool returns the MCP tool definition.
func (t *LocalCommandTool) MCPTool() *mcp.Tool {
	t.mcpToolOnce.Do(func() {
		var err error
		t.mcpTool, err = ConvertProtoToMCPTool(t.tool)
		if err != nil {
			logging.GetLogger().Error("Failed to convert tool to MCP tool", "toolName", t.tool.GetName(), "error", err)
		}
	})
	return t.mcpTool
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
// ctx is the context.
// req is the req.
func (t *LocalCommandTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) { //nolint:gocyclo
	if t.initError != nil {
		return nil, t.initError
	}
	if logging.GetLogger().Enabled(ctx, slog.LevelDebug) {
		logging.GetLogger().Debug("executing tool", "tool", req.ToolName, "inputs", string(req.ToolInputs))
	}

	if allowed, err := EvaluateCompiledCallPolicy(t.policies, t.tool.GetName(), t.callID, req.ToolInputs); err != nil {
		return nil, fmt.Errorf("failed to evaluate call policy: %w", err)
	} else if !allowed {
		return nil, fmt.Errorf("tool execution blocked by policy")
	}
	var inputs map[string]any
	decoder := json.NewDecoder(bytes.NewReader(req.ToolInputs))
	decoder.UseNumber()
	if err := decoder.Decode(&inputs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool inputs: %w", err)
	}

	args := []string{}
	if t.callDefinition.GetArgs() != nil {
		args = append(args, t.callDefinition.GetArgs()...)
	}

	// Substitute placeholders in args with input values
	if inputs != nil {
		for i, arg := range args {
			for k, v := range inputs {
				placeholder := "{{" + k + "}}"
				if strings.Contains(arg, placeholder) {
					val := fmt.Sprintf("%v", v)
					if err := checkForPathTraversal(val); err != nil {
						return nil, fmt.Errorf("parameter %q: %w", k, err)
					}
					if err := checkForArgumentInjection(val); err != nil {
						return nil, fmt.Errorf("parameter %q: %w", k, err)
					}
					args[i] = strings.ReplaceAll(arg, placeholder, val)
				}
			}
		}
	}
	if inputs != nil {
		if argsVal, ok := inputs["args"]; ok {
			// Check if args is allowed in the schema
			argsAllowed := false
			if inputSchema := t.tool.GetInputSchema(); inputSchema != nil {
				if props := inputSchema.Fields["properties"].GetStructValue(); props != nil {
					if _, ok := props.Fields["args"]; ok {
						argsAllowed = true
					}
				}
			}

			if !argsAllowed {
				return nil, fmt.Errorf("'args' parameter is not allowed for this tool")
			}

			if argsList, ok := argsVal.([]any); ok {
				for _, arg := range argsList {
					if argStr, ok := arg.(string); ok {
						if err := checkForPathTraversal(argStr); err != nil {
							return nil, fmt.Errorf("args parameter: %w", err)
						}
						if err := checkForArgumentInjection(argStr); err != nil {
							return nil, fmt.Errorf("args parameter: %w", err)
						}
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

	env := []string{}
	// Inherit only safe environment variables from host
	// We strictly only preserve PATH and system identifiers to avoid leaking secrets
	allowedEnvVars := []string{"PATH", "HOME", "USER", "SHELL", "TMPDIR", "SYSTEMROOT", "WINDIR"}
	for _, key := range allowedEnvVars {
		if val, ok := os.LookupEnv(key); ok {
			env = append(env, fmt.Sprintf("%s=%s", key, val))
		}
	}

	resolvedServiceEnv, err := util.ResolveSecretMap(ctx, t.service.GetEnv(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve service env: %w", err)
	}
	for k, v := range resolvedServiceEnv {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	for _, param := range t.callDefinition.GetParameters() {
		name := param.GetSchema().GetName()
		if secret := param.GetSecret(); secret != nil {
			secretValue, err := util.ResolveSecret(ctx, secret)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve secret for parameter %q: %w", name, err)
			}
			env = append(env, fmt.Sprintf("%s=%s", name, secretValue))
		} else if val, ok := inputs[name]; ok {
			valStr := fmt.Sprintf("%v", val)
			if err := checkForPathTraversal(valStr); err != nil {
				return nil, fmt.Errorf("parameter %q: %w", name, err)
			}
			env = append(env, fmt.Sprintf("%s=%s", name, valStr))
		}
	}

	startTime := time.Now()
	limit := getMaxCommandOutputSize()

	// Differentiate between JSON and environment variable-based communication
	if t.service.GetCommunicationProtocol() == configv1.CommandLineUpstreamService_COMMUNICATION_PROTOCOL_JSON {
		stdin, stdout, stderr, _, err := executor.ExecuteWithStdIO(ctx, t.service.GetCommand(), args, t.service.GetWorkingDirectory(), env)
		if err != nil {
			return nil, fmt.Errorf("failed to execute command with stdio: %w", err)
		}
		// We don't defer stdin.Close() here because we close it in the writer goroutine

		var stderrBuf bytes.Buffer
		stderrDone := make(chan struct{})
		go func() {
			defer close(stderrDone)
			defer func() { _ = stderr.Close() }()
			_, _ = io.Copy(&stderrBuf, io.LimitReader(stderr, limit))
		}()

		var unmarshaledInputs map[string]interface{}
		decoder := json.NewDecoder(bytes.NewReader(req.ToolInputs))
		decoder.UseNumber()
		if err := decoder.Decode(&unmarshaledInputs); err != nil {
			_ = stdin.Close()
			return nil, fmt.Errorf("failed to unmarshal tool inputs: %w", err)
		}

		// Write inputs to stdin in a separate goroutine to avoid deadlock if the command crashes
		go func() {
			defer func() { _ = stdin.Close() }()
			if err := json.NewEncoder(stdin).Encode(unmarshaledInputs); err != nil {
				logging.GetLogger().Warn("Failed to encode inputs to stdin", "error", err)
			}
		}()

		var result map[string]interface{}
		if err := json.NewDecoder(io.LimitReader(stdout, limit)).Decode(&result); err != nil {
			<-stderrDone
			return nil, fmt.Errorf("failed to execute JSON CLI command: %w. Stderr: %s", err, stderrBuf.String())
		}
		return result, nil
	}

	stdout, stderr, exitCodeChan, err := executor.Execute(ctx, t.service.GetCommand(), args, t.service.GetWorkingDirectory(), env)
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	var stdoutBuf, stderrBuf, combinedBuf bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		defer func() { _ = stdout.Close() }()
		_, _ = io.Copy(io.MultiWriter(&stdoutBuf, &combinedBuf), io.LimitReader(stdout, limit))
	}()
	go func() {
		defer wg.Done()
		defer func() { _ = stderr.Close() }()
		_, _ = io.Copy(io.MultiWriter(&stderrBuf, &combinedBuf), io.LimitReader(stderr, limit))
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

// MCPTool returns the MCP tool definition.
func (t *CommandTool) MCPTool() *mcp.Tool {
	t.mcpToolOnce.Do(func() {
		var err error
		t.mcpTool, err = ConvertProtoToMCPTool(t.tool)
		if err != nil {
			logging.GetLogger().Error("Failed to convert tool to MCP tool", "toolName", t.tool.GetName(), "error", err)
		}
	})
	return t.mcpTool
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
// ctx is the context.
// req is the req.
func (t *CommandTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) { //nolint:gocyclo
	if t.initError != nil {
		return nil, t.initError
	}
	if logging.GetLogger().Enabled(ctx, slog.LevelDebug) {
		logging.GetLogger().Debug("executing tool", "tool", req.ToolName, "inputs", string(req.ToolInputs))
	}

	if allowed, err := EvaluateCompiledCallPolicy(t.policies, t.tool.GetName(), t.callID, req.ToolInputs); err != nil {
		return nil, fmt.Errorf("failed to evaluate call policy: %w", err)
	} else if !allowed {
		return nil, fmt.Errorf("tool execution blocked by policy")
	}
	var inputs map[string]any
	decoder := json.NewDecoder(bytes.NewReader(req.ToolInputs))
	decoder.UseNumber()
	if err := decoder.Decode(&inputs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool inputs: %w", err)
	}

	args := []string{}
	if t.callDefinition.GetArgs() != nil {
		args = append(args, t.callDefinition.GetArgs()...)
	}

	// Substitute placeholders in args with input values
	if inputs != nil {
		for i, arg := range args {
			for k, v := range inputs {
				placeholder := "{{" + k + "}}"
				if strings.Contains(arg, placeholder) {
					val := fmt.Sprintf("%v", v)
					if err := checkForPathTraversal(val); err != nil {
						return nil, fmt.Errorf("parameter %q: %w", k, err)
					}
					if err := checkForArgumentInjection(val); err != nil {
						return nil, fmt.Errorf("parameter %q: %w", k, err)
					}
					args[i] = strings.ReplaceAll(arg, placeholder, val)
				}
			}
		}
	}

	if inputs != nil {
		if argsVal, ok := inputs["args"]; ok {
			// Check if args is allowed in the schema
			argsAllowed := false
			if inputSchema := t.tool.GetInputSchema(); inputSchema != nil {
				if props := inputSchema.Fields["properties"].GetStructValue(); props != nil {
					if _, ok := props.Fields["args"]; ok {
						argsAllowed = true
					}
				}
			}

			if !argsAllowed {
				return nil, fmt.Errorf("'args' parameter is not allowed for this tool")
			}

			if argsList, ok := argsVal.([]any); ok {
				for _, arg := range argsList {
					if argStr, ok := arg.(string); ok {
						if err := checkForPathTraversal(argStr); err != nil {
							return nil, fmt.Errorf("args parameter: %w", err)
						}
						if err := checkForArgumentInjection(argStr); err != nil {
							return nil, fmt.Errorf("args parameter: %w", err)
						}
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

	executor := t.getExecutor(t.service.GetContainerEnvironment())

	env := []string{}
	// Only inherit safe environment variables from host for local execution
	// For Docker execution, we should not inherit host environment variables at all
	isDocker := t.service.GetContainerEnvironment() != nil && t.service.GetContainerEnvironment().GetImage() != ""
	if !isDocker {
		// We strictly only preserve PATH and system identifiers to avoid leaking secrets
		allowedEnvVars := []string{"PATH", "HOME", "USER", "SHELL", "TMPDIR", "SYSTEMROOT", "WINDIR"}
		for _, key := range allowedEnvVars {
			if val, ok := os.LookupEnv(key); ok {
				env = append(env, fmt.Sprintf("%s=%s", key, val))
			}
		}
	}

	resolvedServiceEnv, err := util.ResolveSecretMap(ctx, t.service.GetEnv(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve service env: %w", err)
	}
	for k, v := range resolvedServiceEnv {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	if ce := t.service.GetContainerEnvironment(); ce != nil {
		resolvedContainerEnv, err := util.ResolveSecretMap(ctx, ce.GetEnv(), nil)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve container env: %w", err)
		}
		for k, v := range resolvedContainerEnv {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	for _, param := range t.callDefinition.GetParameters() {
		name := param.GetSchema().GetName()
		if secret := param.GetSecret(); secret != nil {
			secretValue, err := util.ResolveSecret(ctx, secret)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve secret for parameter %q: %w", name, err)
			}
			env = append(env, fmt.Sprintf("%s=%s", name, secretValue))
		} else if val, ok := inputs[name]; ok {
			valStr := fmt.Sprintf("%v", val)
			if err := checkForPathTraversal(valStr); err != nil {
				return nil, fmt.Errorf("parameter %q: %w", name, err)
			}
			env = append(env, fmt.Sprintf("%s=%s", name, valStr))
		}
	}

	startTime := time.Now()
	limit := getMaxCommandOutputSize()

	// Differentiate between JSON and environment variable-based communication
	if t.service.GetCommunicationProtocol() == configv1.CommandLineUpstreamService_COMMUNICATION_PROTOCOL_JSON {
		stdin, stdout, stderr, _, err := executor.ExecuteWithStdIO(ctx, t.service.GetCommand(), args, t.service.GetWorkingDirectory(), env)
		if err != nil {
			return nil, fmt.Errorf("failed to execute command with stdio: %w", err)
		}
		// We don't defer stdin.Close() here because we close it in the writer goroutine

		var stderrBuf bytes.Buffer
		stderrDone := make(chan struct{})
		go func() {
			defer close(stderrDone)
			defer func() { _ = stderr.Close() }()
			_, _ = io.Copy(&stderrBuf, io.LimitReader(stderr, limit))
		}()

		var unmarshaledInputs map[string]interface{}
		decoder := json.NewDecoder(bytes.NewReader(req.ToolInputs))
		decoder.UseNumber()
		if err := decoder.Decode(&unmarshaledInputs); err != nil {
			_ = stdin.Close()
			return nil, fmt.Errorf("failed to unmarshal tool inputs: %w", err)
		}

		// Write inputs to stdin in a separate goroutine to avoid deadlock if the command crashes
		go func() {
			defer func() { _ = stdin.Close() }()
			if err := json.NewEncoder(stdin).Encode(unmarshaledInputs); err != nil {
				logging.GetLogger().Warn("Failed to encode inputs to stdin", "error", err)
			}
		}()

		var result map[string]interface{}
		if err := json.NewDecoder(io.LimitReader(stdout, limit)).Decode(&result); err != nil {
			<-stderrDone
			return nil, fmt.Errorf("failed to execute JSON CLI command: %w. Stderr: %s", err, stderrBuf.String())
		}
		return result, nil
	}

	stdout, stderr, exitCodeChan, err := executor.Execute(ctx, t.service.GetCommand(), args, t.service.GetWorkingDirectory(), env)
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	var stdoutBuf, stderrBuf, combinedBuf bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		defer func() { _ = stdout.Close() }()
		_, _ = io.Copy(io.MultiWriter(&stdoutBuf, &combinedBuf), io.LimitReader(stdout, limit))
	}()
	go func() {
		defer wg.Done()
		defer func() { _ = stderr.Close() }()
		_, _ = io.Copy(io.MultiWriter(&stderrBuf, &combinedBuf), io.LimitReader(stderr, limit))
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

// prettyPrint formats the input based on content type for better readability.
func prettyPrint(input []byte, contentType string) string {
	if len(input) == 0 {
		return ""
	}

	contentType = strings.ToLower(contentType)
	// Handle binary data
	if strings.HasPrefix(contentType, "image/") ||
		strings.HasPrefix(contentType, "audio/") ||
		strings.HasPrefix(contentType, "video/") ||
		contentType == "application/octet-stream" {
		return fmt.Sprintf("[Binary Data: %d bytes]", len(input))
	}

	// Try JSON
	if strings.Contains(contentType, "json") {
		// Redact sensitive keys
		input = util.RedactJSON(input)

		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, input, "", "  "); err == nil {
			return prettyJSON.String()
		}
		// If JSON parsing fails, fall through to return string(input)
	}

	// Try XML
	if strings.Contains(contentType, "xml") {
		decoder := xml.NewDecoder(bytes.NewReader(input))
		var buf bytes.Buffer
		encoder := xml.NewEncoder(&buf)
		encoder.Indent("", "  ")

		// Attempt to decode and re-encode to format
		for {
			token, err := decoder.Token()
			if err == io.EOF {
				_ = encoder.Flush()
				return buf.String()
			}
			if err != nil {
				// XML parsing failed, return raw string
				return string(input)
			}
			if err := encoder.EncodeToken(token); err != nil {
				return string(input)
			}
		}
	}

	return string(input)
}

func (t *CommandTool) getExecutor(env *configv1.ContainerEnvironment) command.Executor {
	if t.executorFactory != nil {
		return t.executorFactory(env)
	}
	return command.NewExecutor(env)
}

// getMaxCommandOutputSize returns the maximum size of the command output (stdout + stderr) in bytes.
// It checks the MCPANY_MAX_COMMAND_OUTPUT_SIZE environment variable.
func getMaxCommandOutputSize() int64 {
	val := os.Getenv("MCPANY_MAX_COMMAND_OUTPUT_SIZE")
	if val != "" {
		if size, err := strconv.ParseInt(val, 10, 64); err == nil {
			return size
		}
		// Log error? For now just fallback.
	}
	return consts.DefaultMaxCommandOutputBytes
}

func isSensitiveHeader(key string) bool {
	k := strings.ToLower(key)
	if k == "authorization" || k == "proxy-authorization" || k == "cookie" || k == "set-cookie" || k == "x-api-key" {
		return true
	}
	if strings.Contains(k, "token") || strings.Contains(k, "secret") || strings.Contains(k, "password") {
		return true
	}
	return false
}

func checkForPathTraversal(val string) error {
	if val == ".." {
		return fmt.Errorf("path traversal attempt detected")
	}
	if strings.HasPrefix(val, "../") || strings.HasPrefix(val, "..\\") {
		return fmt.Errorf("path traversal attempt detected")
	}
	if strings.HasSuffix(val, "/..") || strings.HasSuffix(val, "\\..") {
		return fmt.Errorf("path traversal attempt detected")
	}
	if strings.Contains(val, "/../") || strings.Contains(val, "\\..\\") || strings.Contains(val, "/..\\") || strings.Contains(val, "\\../") {
		return fmt.Errorf("path traversal attempt detected")
	}
	return nil
}

func checkForArgumentInjection(val string) error {
	if strings.HasPrefix(val, "-") {
		// Allow negative numbers
		if _, err := strconv.ParseFloat(val, 64); err == nil {
			return nil
		}
		return fmt.Errorf("argument injection detected: value starts with '-'")
	}
	return nil
}
