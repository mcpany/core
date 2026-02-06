// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"bytes"
	"context"
	stdjson "encoding/json" // Renamed to stdjson to avoid conflict
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	jsoniter "github.com/json-iterator/go"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/command"
	"github.com/mcpany/core/server/pkg/consts"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/metrics"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/resilience"
	"github.com/mcpany/core/server/pkg/transformer"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/pkg/validation"
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

var (
	metricGrpcRequestError   = []string{"grpc", "request", "error"}
	metricGrpcRequestSuccess = []string{"grpc", "request", "success"}
	metricGrpcRequestLatency = []string{"grpc", "request", "latency"}
	metricHTTPRequestError   = []string{"http", "request", "error"}
	metricHTTPRequestSuccess = []string{"http", "request", "success"}
	metricHTTPRequestLatency = []string{"http", "request", "latency"}
)

var fastJSON = jsoniter.ConfigCompatibleWithStandardLibrary

// Tool is the fundamental interface for any executable tool in the system.
//
// Summary: Abstract interface for tool implementations (gRPC, HTTP, CLI).
type Tool interface {
	// Tool returns the protobuf definition of the tool.
	//
	// Summary: Retrieves the internal protobuf definition.
	//
	// Returns:
	//   - *v1.Tool: The protobuf tool definition.
	Tool() *v1.Tool

	// MCPTool returns the MCP tool definition.
	//
	// Summary: Retrieves the MCP-compliant tool definition.
	//
	// Returns:
	//   - *mcp.Tool: The MCP tool definition.
	MCPTool() *mcp.Tool

	// Execute runs the tool with the provided context and request.
	//
	// Summary: Executes the tool logic.
	//
	// Parameters:
	//   - ctx: context.Context. The execution context.
	//   - req: *ExecutionRequest. The request containing tool name and inputs.
	//
	// Returns:
	//   - any: The execution result (usually a map or string).
	//   - error: An error if execution fails.
	Execute(ctx context.Context, req *ExecutionRequest) (any, error)

	// GetCacheConfig returns the cache configuration for the tool.
	//
	// Summary: Retrieves caching rules for the tool.
	//
	// Returns:
	//   - *configv1.CacheConfig: The cache configuration, or nil if disabled.
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

	// CompiledPolicies are the pre-compiled call policies for the service.
	CompiledPolicies []*CompiledCallPolicy

	// HealthStatus indicates the health of the service ("healthy", "unhealthy", "unknown").
	HealthStatus string
}

// ExecutionRequest represents a request to execute a specific tool.
//
// Summary: Payload for tool execution requests.
type ExecutionRequest struct {
	// ToolName is the name of the tool to be executed.
	ToolName string `json:"name"`
	// ToolInputs is the raw JSON message of the tool inputs. It is used by
	// tools that need to unmarshal the inputs into a specific struct.
	ToolInputs stdjson.RawMessage `json:"toolInputs"`
	// Arguments is a map of the tool inputs. It is used by tools that need to
	// access the inputs as a map.
	Arguments map[string]interface{} `json:"arguments"`
	// DryRun indicates whether the tool should be executed in dry-run mode.
	// In dry-run mode, the tool should validate inputs and return a preview
	// of the execution without performing any side effects.
	DryRun bool `json:"dryRun"`
	// Tool is the resolved tool instance. Populated internally to avoid re-lookup.
	Tool Tool `json:"-"`
}

// ServiceRegistry defines an interface for a component that can look up tools
// and service information. It is used for dependency injection to decouple
// components from the main service registry.
type ServiceRegistry interface {
	// GetTool retrieves a tool by name.
	//
	// Parameters:
	//   - toolName: The name of the tool to retrieve.
	//
	// Returns:
	//   - Tool: The tool instance if found.
	//   - bool: True if the tool exists, false otherwise.
	GetTool(toolName string) (Tool, bool)

	// GetServiceInfo retrieves metadata for a service.
	//
	// Parameters:
	//   - serviceID: The unique identifier of the service.
	//
	// Returns:
	//   - *ServiceInfo: The service metadata info if found.
	//   - bool: True if the service exists, false otherwise.
	GetServiceInfo(serviceID string) (*ServiceInfo, bool)
}

// ExecutionFunc represents the next middleware in the chain.
type ExecutionFunc func(ctx context.Context, req *ExecutionRequest) (any, error)

type contextKey string

const toolContextKey = contextKey("tool")

// NewContextWithTool creates a new context with the given tool embedded.
//
// Parameters:
//   - ctx: The context to extend.
//   - t: The tool instance to embed in the context.
//
// Returns:
//   - context.Context: A new context containing the tool.
func NewContextWithTool(ctx context.Context, t Tool) context.Context {
	return context.WithValue(ctx, toolContextKey, t)
}

// GetFromContext retrieves a tool from the context if present.
//
// Parameters:
//   - ctx: The context to search.
//
// Returns:
//   - Tool: The tool instance from the context.
//   - bool: True if a tool was found, false otherwise.
func GetFromContext(ctx context.Context) (Tool, bool) {
	t, ok := ctx.Value(toolContextKey).(Tool)
	return t, ok
}

// Callable is an interface that represents a callable tool.
type Callable interface {
	// Call executes the callable with the given request.
	//
	// Parameters:
	//   - ctx: The context for the request.
	//   - req: The execution request details.
	//
	// Returns:
	//   - any: The result of the execution.
	//   - error: An error if the operation fails.
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
//
// Parameters:
//   - ctx: The context to extend.
//   - cc: The CacheControl instance to embed.
//
// Returns:
//   - context.Context: A new context containing the CacheControl.
func NewContextWithCacheControl(ctx context.Context, cc *CacheControl) context.Context {
	return context.WithValue(ctx, cacheControlContextKey, cc)
}

// GetCacheControl retrieves the CacheControl from the context.
//
// Parameters:
//   - ctx: The context to search.
//
// Returns:
//   - *CacheControl: The CacheControl instance if found.
//   - bool: True if CacheControl exists, false otherwise.
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
	resilienceManager *resilience.Manager
}

// NewGRPCTool creates a new GRPCTool instance.
//
// Parameters:
//   - tool: The protobuf definition of the tool.
//   - poolManager: The connection pool manager for gRPC connections.
//   - serviceID: The identifier for the service.
//   - method: The gRPC method descriptor.
//   - callDefinition: The configuration for the gRPC call.
//   - resilienceConfig: The resilience configuration (retries, timeouts, etc.).
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
func (t *GRPCTool) Tool() *v1.Tool {
	return t.tool
}

// MCPTool returns the MCP-compliant tool definition.
//
// It lazily converts the internal protobuf definition to the MCP format on first access.
//
// Returns:
//   - *mcp.Tool: The MCP tool definition.
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
func (t *GRPCTool) GetCacheConfig() *configv1.CacheConfig {
	return t.cache
}

// Execute handles the execution of the gRPC tool. It retrieves a client from the
// pool, unmarshals the JSON input into a protobuf request message, invokes the
// gRPC method, and marshals the protobuf response back to JSON.
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
	allowedParams     map[string]bool

	// Cached fields for performance
	initError            error
	cachedMethod         string
	cachedURL            *url.URL
	pathSegments         []urlSegment
	querySegments        []urlSegment
	paramInPath          []bool
	paramInQuery         []bool
	cachedInputTemplate  *transformer.TextTemplate
	cachedOutputTemplate *transformer.TextTemplate
}

// NewHTTPTool creates a new HTTPTool instance.
//
// Parameters:
//   - tool: The protobuf definition of the tool.
//   - poolManager: The connection pool manager for HTTP connections.
//   - serviceID: The identifier for the service.
//   - authenticator: The authenticator for upstream requests.
//   - callDefinition: The configuration for the HTTP call.
//   - cfg: The resilience configuration.
//   - policies: The security policies for the call.
//   - callID: The unique identifier for the call.
//
// Returns:
//   - *HTTPTool: The initialized HTTPTool.
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
		allowedParams:     make(map[string]bool, len(callDefinition.GetParameters())),
	}

	compiled, err := CompileCallPolicies(policies)
	if err != nil {
		t.initError = fmt.Errorf("failed to compile call policies: %w", err)
	}
	t.policies = compiled

	// Cache templates
	if it := t.inputTransformer; it != nil && it.GetTemplate() != "" { //nolint:staticcheck
		tpl, err := transformer.NewTemplate(it.GetTemplate(), "{{", "}}") //nolint:staticcheck
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

	// Pre-calculate URL components
	// Use SplitN to allow spaces in the URL (e.g. in query parameters with invalid encoding)
	methodAndURL := strings.SplitN(tool.GetUnderlyingMethodFqn(), " ", 2)
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

	queryStr := u.RawQuery
	queryStr = strings.ReplaceAll(queryStr, "%7B", "{")
	queryStr = strings.ReplaceAll(queryStr, "%7D", "}")

	t.pathSegments = parseURLSegments(pathStr)
	t.querySegments = parseURLSegments(queryStr)

	t.paramInPath = make([]bool, len(callDefinition.GetParameters()))
	t.paramInQuery = make([]bool, len(callDefinition.GetParameters()))

	for i, param := range callDefinition.GetParameters() {
		if schema := param.GetSchema(); schema != nil {
			name := schema.GetName()
			t.allowedParams[name] = true
			placeholder := "{{" + name + "}}"

			if strings.Contains(pathStr, placeholder) {
				t.paramInPath[i] = true
			}
			if strings.Contains(queryStr, placeholder) {
				t.paramInQuery[i] = true
			}
		}
	}

	return t
}

// Tool returns the protobuf definition of the HTTP tool.
//
// Returns:
//   - *v1.Tool: The underlying protobuf definition.
func (t *HTTPTool) Tool() *v1.Tool {
	return t.tool
}

// MCPTool returns the MCP-compliant tool definition.
//
// It lazily converts the internal protobuf definition to the MCP format on first access.
//
// Returns:
//   - *mcp.Tool: The MCP tool definition.
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
//
// Returns:
//   - *configv1.CacheConfig: The cache configuration, if any.
func (t *HTTPTool) GetCacheConfig() *configv1.CacheConfig {
	return t.cache
}

// Execute handles the execution of the HTTP tool. It builds an HTTP request by
// mapping input parameters to the path, query, and body, applies any
// configured transformations, sends the request, and processes the response.
func (t *HTTPTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	if logging.GetLogger().Enabled(ctx, slog.LevelDebug) {
		logging.GetLogger().Debug("executing tool", "tool", req.ToolName, "inputs", prettyPrint(req.ToolInputs, contentTypeJSON))
	}
	defer metrics.MeasureSince(metricHTTPRequestLatency, time.Now())

	if allowed, err := EvaluateCompiledCallPolicy(t.policies, t.tool.GetName(), t.callID, req.ToolInputs); err != nil {
		return nil, fmt.Errorf("failed to evaluate call policy: %w", err)
	} else if !allowed {
		return nil, fmt.Errorf("tool execution blocked by policy")
	}

	if t.initError != nil {
		return nil, t.initError
	}

	httpPool, ok := pool.Get[*client.HTTPClientWrapper](t.poolManager, t.serviceID)
	if !ok {
		metrics.IncrCounter(metricHTTPRequestError, 1)
		return nil, fmt.Errorf("no http pool found for service: %s", t.serviceID)
	}

	httpClient, err := httpPool.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get client from pool: %w", err)
	}
	defer httpPool.Put(httpClient)

	inputs, urlString, inputsModified, err := t.prepareInputsAndURL(ctx, req)
	if err != nil {
		return nil, err
	}

	if err := validation.IsSafeURL(urlString); err != nil {
		return nil, fmt.Errorf("unsafe url: %w", err)
	}

	body, contentType, err := t.prepareBody(ctx, inputs, t.cachedMethod, req.ToolName, req.ToolInputs, inputsModified)
	if err != nil {
		return nil, err
	}

	if req.DryRun {
		logging.GetLogger().Info("Dry run execution", "tool", req.ToolName)
		dryRunResult := map[string]any{
			"dry_run": true,
			"request": map[string]any{
				"method": t.cachedMethod,
				"url":    urlString,
				"headers": map[string]string{
					"Content-Type": contentType,
				},
			},
		}
		if body != nil {
			if seeker, ok := body.(io.Seeker); ok {
				_, _ = seeker.Seek(0, io.SeekStart)
			}
			bodyBytes, _ := io.ReadAll(body)
			dryRunResult["request"].(map[string]any)["body"] = string(bodyBytes)
		}
		return dryRunResult, nil
	}

	var resp *http.Response
	work := func(ctx context.Context) error {
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

		httpReq, err := t.createHTTPRequest(ctx, urlString, bodyForAttempt, contentType, inputs)
		if err != nil {
			return &resilience.PermanentError{Err: err}
		}

		if logging.GetLogger().Enabled(ctx, slog.LevelDebug) {
			t.logRequest(ctx, httpReq, bodyForAttempt)
		}

		attemptResp, err := httpClient.Do(httpReq)
		if err != nil {
			return fmt.Errorf("failed to execute http request: %w", err)
		}

		if attemptResp.StatusCode == http.StatusTooManyRequests {
			_ = attemptResp.Body.Close()
			return fmt.Errorf("upstream HTTP request failed with status %d (Too Many Requests)", attemptResp.StatusCode)
		}

		if attemptResp.StatusCode >= 400 {
			// Read body to include in error message
			// Limit to 1KB to avoid flooding logs/context
			bodyBytes, _ := io.ReadAll(io.LimitReader(attemptResp.Body, 1024))
			_ = attemptResp.Body.Close()

			// Try to redact JSON content to avoid leaking sensitive fields in error messages
			// (e.g. if the error response contains the request payload or secrets)
			bodyBytes = util.RedactJSON(bodyBytes)
			bodyStr := string(bodyBytes)

			// Sentinel Security Update: Redact secrets from URL in logs
			logURL := t.redactURL(httpReq.URL)
			logging.GetLogger().DebugContext(ctx, "Upstream HTTP error", "status", attemptResp.StatusCode, "body", bodyStr, "url", logURL)

			// Truncate body for the returned error message to prevent leaking large stack traces or extensive details to the user/LLM.
			// We keep enough to likely identify the issue (e.g. "invalid argument").
			displayBody := bodyStr
			const maxErrorBodyLen = 200
			if len(displayBody) > maxErrorBodyLen {
				displayBody = displayBody[:maxErrorBodyLen] + "... (truncated)"
			}

			// Security: Hide the body if it is not JSON (potential stack trace) unless debug is enabled.
			// util.RedactJSON returns the original input if it's not JSON.
			// If it was JSON, it is already redacted.
			isDebug := os.Getenv("MCPANY_DEBUG") == "true"
			if !isDebug && !stdjson.Valid(bodyBytes) {
				displayBody = "[Body hidden for security. Enable debug mode to view.]"
			}

			errMsg := fmt.Errorf("upstream HTTP request failed with status %d: %s", attemptResp.StatusCode, displayBody)

			if attemptResp.StatusCode < 500 {
				return &resilience.PermanentError{Err: errMsg}
			}
			return errMsg
		}

		resp = attemptResp
		return nil
	}

	if err := t.resilienceManager.Execute(ctx, work); err != nil {
		metrics.IncrCounter(metricHTTPRequestError, 1)
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()
	metrics.IncrCounter(metricHTTPRequestSuccess, 1)

	return t.processResponse(ctx, resp)
}

func (t *HTTPTool) createHTTPRequest(ctx context.Context, urlString string, body io.Reader, contentType string, inputs map[string]interface{}) (*http.Request, error) {
	httpReq, err := http.NewRequestWithContext(ctx, t.cachedMethod, urlString, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	if contentType != "" {
		httpReq.Header.Set("Content-Type", contentType)
	}
	httpReq.Header.Set("Accept", "*/*")
	httpReq.Header.Set("User-Agent", "MCPAny/1.0 (https://github.com/mcpany/core; contact@mcpany.org)")

	if t.authenticator != nil {
		if err := t.authenticator.Authenticate(httpReq); err != nil {
			return nil, fmt.Errorf("failed to authenticate request: %w", err)
		}
		logging.GetLogger().Debug("Applied authentication", "user_agent", httpReq.Header.Get("User-Agent"))
	} else {
		logging.GetLogger().Debug("No authenticator configured")
	}

	if t.cachedMethod == http.MethodGet || t.cachedMethod == http.MethodDelete {
		q := httpReq.URL.Query()
		for key, value := range inputs {
			q.Add(key, util.ToString(value))
		}
		httpReq.URL.RawQuery = q.Encode()
	}
	return httpReq, nil
}

func (t *HTTPTool) logRequest(ctx context.Context, httpReq *http.Request, body io.Reader) {
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
	if body != nil {
		contentType := httpReq.Header.Get("Content-Type")
		bodyBytes, _ := io.ReadAll(body)
		// Restore body
		if seeker, ok := body.(io.Seeker); ok {
			_, _ = seeker.Seek(0, io.SeekStart)
		}
		logging.GetLogger().DebugContext(ctx, "sending http request body", "body", prettyPrint(bodyBytes, contentType))
	}
}

func (t *HTTPTool) prepareInputsAndURL(ctx context.Context, req *ExecutionRequest) (map[string]any, string, bool, error) {
	var inputs map[string]any
	if len(req.ToolInputs) > 0 {
		// Trim whitespace to avoid EOF errors on empty/whitespace-only inputs
		req.ToolInputs = bytes.TrimSpace(req.ToolInputs)
	}

	// ⚡ Bolt: Use json-iterator
	if len(req.ToolInputs) > 0 {
		decoder := fastJSON.NewDecoder(bytes.NewReader(req.ToolInputs))
		decoder.UseNumber()
		if err := decoder.Decode(&inputs); err != nil {
			return nil, "", false, fmt.Errorf("failed to unmarshal tool inputs: %w (inputs: %q)", err, string(req.ToolInputs))
		}
	}

	// Filter undefined parameters from inputs to prevent mass assignment/pollution
	filtered := false
	for k := range inputs {
		if !t.allowedParams[k] {
			delete(inputs, k)
			filtered = true
		}
	}

	pathReplacements, queryReplacements, inputsModified, err := t.processParameters(ctx, inputs)
	if err != nil {
		return nil, "", false, err
	}
	inputsModified = inputsModified || filtered

	var pathBuf strings.Builder
	for _, seg := range t.pathSegments {
		if seg.isParam {
			if val, ok := pathReplacements[seg.value]; ok {
				pathBuf.WriteString(val)
			} else {
				pathBuf.WriteString("{{" + seg.value + "}}")
			}
		} else {
			pathBuf.WriteString(seg.value)
		}
	}
	pathStr := pathBuf.String()

	var queryBuf strings.Builder
	for _, seg := range t.querySegments {
		if seg.isParam {
			if val, ok := queryReplacements[seg.value]; ok {
				queryBuf.WriteString(val)
			} else {
				queryBuf.WriteString("{{" + seg.value + "}}")
			}
		} else {
			queryBuf.WriteString(seg.value)
		}
	}
	queryStr := queryBuf.String()

	// Clean the path to resolve . and .. and //
	// We do this on the encoded string to treat %2F as opaque characters
	// This prevents path.Clean from treating encoded slashes as separators
	// and messing up the re-encoding later (which would convert %2F to /).
	hadTrailingSlash := strings.HasSuffix(pathStr, "/")
	// Detect if it was specifically root double slash which cleans to / but should be restored to //
	wasRootDoubleSlash := pathStr == "//"

	pathStr = cleanPathPreserveDoubleSlash(pathStr)
	if hadTrailingSlash && (pathStr != "/" || wasRootDoubleSlash) {
		pathStr += "/"
	}

	// Reconstruct URL string manually to avoid re-encoding
	var buf strings.Builder
	if t.cachedURL.Scheme != "" {
		buf.WriteString(t.cachedURL.Scheme)
		buf.WriteString("://")
	}
	if t.cachedURL.User != nil {
		buf.WriteString(t.cachedURL.User.String())
		buf.WriteString("@")
	}
	buf.WriteString(t.cachedURL.Host)
	if pathStr != "" && !strings.HasPrefix(pathStr, "/") {
		buf.WriteString("/")
	}
	buf.WriteString(pathStr)
	if queryStr != "" {
		buf.WriteString("?")
		buf.WriteString(queryStr)
	}
	urlString := buf.String()

	return inputs, urlString, inputsModified, nil
}

func (t *HTTPTool) processParameters(ctx context.Context, inputs map[string]any) (map[string]string, map[string]string, bool, error) {
	pathReplacements := make(map[string]string, len(t.parameters))
	queryReplacements := make(map[string]string, len(t.parameters))
	inputsModified := false

	for i, param := range t.parameters {
		name := param.GetSchema().GetName()
		if secret := param.GetSecret(); secret != nil {
			secretValue, err := util.ResolveSecret(ctx, secret)
			if err != nil {
				return nil, nil, false, fmt.Errorf("failed to resolve secret for parameter %q: %w", name, err)
			}
			if t.paramInPath[i] {
				pathReplacements[name] = secretValue
			}
			if t.paramInQuery[i] {
				queryReplacements[name] = secretValue
			}
		} else if schema := param.GetSchema(); schema != nil {
			val, ok := inputs[name]
			if !ok {
				if schema.GetIsRequired() {
					return nil, nil, false, fmt.Errorf("missing required parameter: %s", name)
				}
				// If optional and missing, treat as empty string
				val = ""
			} else if t.paramInPath[i] || t.paramInQuery[i] {
				// Only delete from inputs if it was present AND used in path/query replacement
				delete(inputs, name)
				inputsModified = true
			}

			valStr := util.ToString(val)

			if t.paramInPath[i] {
				// ALWAYS check for path traversal in path parameters, regardless of escaping settings.
				// Even if escaped, some servers might decode and normalize the path, leading to traversal.
				if err := checkForPathTraversal(valStr); err != nil {
					return nil, nil, false, fmt.Errorf("path traversal attempt detected in parameter %q: %w", name, err)
				}
				// Also check decoded value just in case the input was already encoded
				if decodedVal, err := url.QueryUnescape(valStr); err == nil && decodedVal != valStr {
					if err := checkForPathTraversal(decodedVal); err != nil {
						return nil, nil, false, fmt.Errorf("path traversal attempt detected in parameter %q (decoded): %w", name, err)
					}
				}
			}

			if param.GetDisableEscape() {
				if t.paramInPath[i] {
					pathReplacements[name] = valStr
				}
				if t.paramInQuery[i] {
					queryReplacements[name] = valStr
				}
			} else {
				if t.paramInPath[i] {
					// Even if we escape, we should check for ".." in the input because
					// url.PathEscape does NOT escape dots, so ".." remains ".."
					// and path.Clean will resolve it, potentially allowing path traversal.
					if err := checkForPathTraversal(valStr); err != nil {
						return nil, nil, false, fmt.Errorf("path traversal attempt detected in parameter %q: %w", name, err)
					}
					pathReplacements[name] = url.PathEscape(valStr)
				}
				if t.paramInQuery[i] {
					queryReplacements[name] = url.QueryEscape(valStr)
				}
			}
		}
	}
	return pathReplacements, queryReplacements, inputsModified, nil
}

type urlSegment struct {
	isParam bool
	value   string // Literal text or parameter name
}

func parseURLSegments(template string) []urlSegment {
	parts := strings.Split(template, "{{")
	segments := make([]urlSegment, 0, len(parts)*2)
	for i, part := range parts {
		if i == 0 {
			if part != "" {
				segments = append(segments, urlSegment{isParam: false, value: part})
			}
			continue
		}
		// part looks like "param}}suffix"
		subparts := strings.SplitN(part, "}}", 2)
		// If "}}" is missing, it's not a valid parameter replacement, so treat as literal "{{" + part
		if len(subparts) == 1 {
			segments = append(segments, urlSegment{isParam: false, value: "{{" + part})
			continue
		}

		paramName := subparts[0]
		segments = append(segments, urlSegment{isParam: true, value: paramName})
		if len(subparts) > 1 && subparts[1] != "" {
			segments = append(segments, urlSegment{isParam: false, value: subparts[1]})
		}
	}
	return segments
}

func (t *HTTPTool) prepareBody(ctx context.Context, inputs map[string]any, method string, toolName string, originalInputs []byte, inputsModified bool) (io.Reader, string, error) {
	if inputs == nil {
		return nil, "", nil
	}
	if method != http.MethodPost && method != http.MethodPut {
		return nil, "", nil
	}

	// ⚡ Bolt: Use json-iterator
	var body io.Reader
	var contentType string

	switch {
	case t.webhookClient != nil:
		// Use webhook for transformation
		data := map[string]any{
			"kind":      configv1.WebhookKind_WEBHOOK_KIND_TRANSFORM_INPUT,
			"tool_name": toolName,
			"inputs":    inputs,
		}
		respEvent, err := t.webhookClient.Call(ctx, "com.mcpany.tool.transform_input", data)
		if err != nil {
			return nil, "", fmt.Errorf("transformation webhook failed: %w", err)
		}
		// We expect the data to be the transformed body
		respData := respEvent.Data()
		if len(respData) > 0 {
			body = bytes.NewReader(respData)
			// Verify if it looks like JSON?
			if fastJSON.Valid(respData) {
				contentType = contentTypeJSON
			}
		}
	case t.cachedInputTemplate != nil:
		renderedBody, err := t.cachedInputTemplate.Render(inputs)
		if err != nil {
			return nil, "", fmt.Errorf("failed to render input template: %w", err)
		}
		body = strings.NewReader(renderedBody)
	case t.inputTransformer != nil && t.inputTransformer.GetTemplate() != "": //nolint:staticcheck
		// Fallback for unexpected case
		return nil, "", fmt.Errorf("input template configured but not cached (initialization error?)")
	default:
		// Optimization: if inputs were not modified, reuse the original bytes
		if !inputsModified && len(originalInputs) > 0 {
			body = bytes.NewReader(originalInputs)
		} else {
			jsonBytes, err := fastJSON.Marshal(inputs)
			if err != nil {
				return nil, "", fmt.Errorf("failed to marshal tool inputs to json: %w", err)
			}
			body = bytes.NewReader(jsonBytes)
		}
		contentType = contentTypeJSON
	}

	return body, contentType, nil
}

func (t *HTTPTool) processResponse(ctx context.Context, resp *http.Response) (any, error) {
	maxSize := getMaxHTTPResponseSize()
	// Read up to maxSize + 1 to detect if it exceeds the limit
	reader := io.LimitReader(resp.Body, maxSize+1)
	respBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read http response body: %w", err)
	}
	if int64(len(respBody)) > maxSize {
		return nil, fmt.Errorf("response body exceeds maximum size of %d bytes", maxSize)
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

	// ⚡ Bolt: Use json-iterator
	var result any
	if err := fastJSON.Unmarshal(respBody, &result); err != nil {
		return string(respBody), nil //nolint:nilerr
	}

	return result, nil
}

// redactURL removes sensitive query parameters from the URL for logging.
func (t *HTTPTool) redactURL(u *url.URL) string {
	// Check if we have any secret parameters mapped
	hasSecretParams := false
	for _, param := range t.parameters {
		if param.GetSecret() != nil {
			hasSecretParams = true
			break
		}
	}

	if !hasSecretParams {
		return u.String()
	}

	// Create a copy of the URL to avoid modifying the original
	redactedURL := *u
	q := redactedURL.Query()

	for _, param := range t.parameters {
		if param.GetSecret() != nil {
			name := param.GetSchema().GetName()
			if q.Has(name) {
				q.Set(name, redactedPlaceholder)
			}
		}
	}
	redactedURL.RawQuery = q.Encode()
	return redactedURL.String()
}

// MCPTool implements the Tool interface for a tool that is exposed via another
// MCP-compliant service. It acts as a proxy, forwarding the tool call to the
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
// Parameters:
//   - tool: The protobuf definition of the tool.
//   - client: The MCP client for downstream communication.
//   - callDefinition: The configuration for the MCP call.
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

	// Cache templates
	if it := t.inputTransformer; it != nil && it.GetTemplate() != "" { //nolint:staticcheck
		tpl, err := transformer.NewTemplate(it.GetTemplate(), "{{", "}}") //nolint:staticcheck
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
func (t *MCPTool) Tool() *v1.Tool {
	return t.tool
}

// MCPTool returns the MCP-compliant tool definition.
//
// It lazily converts the internal protobuf definition to the MCP format on first access.
//
// Returns:
//   - *mcp.Tool: The MCP tool definition.
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
func (t *MCPTool) GetCacheConfig() *configv1.CacheConfig {
	return t.cache
}

// Execute handles the execution of the MCP tool. It forwards the tool call,
// including its name and arguments, to the downstream MCP service using the
// configured client and applies any necessary transformations to the request
// and response.
func (t *MCPTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	if t.initError != nil {
		return nil, t.initError
	}
	if logging.GetLogger().Enabled(ctx, slog.LevelDebug) {
		logging.GetLogger().Debug("executing tool", "tool", req.ToolName, "inputs", prettyPrint(req.ToolInputs, contentTypeJSON))
	}
	// Use the tool name from the definition, as the request tool name might be sanitized/modified
	bareToolName := t.tool.GetName()

	var inputs map[string]any
	if len(bytes.TrimSpace(req.ToolInputs)) == 0 {
		req.ToolInputs = []byte("{}")
	}

	// ⚡ Bolt: Use json-iterator
	decoder := fastJSON.NewDecoder(bytes.NewReader(req.ToolInputs))
	decoder.UseNumber()
	if err := decoder.Decode(&inputs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool inputs: %w", err)
	}

	var arguments stdjson.RawMessage // Use stdjson for compatibility with SDK or struct? mcp.CallToolParams expects json.RawMessage (from encoding/json)
	// mcp.CallToolParams.Arguments is json.RawMessage (standard).
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
	case t.inputTransformer != nil && t.inputTransformer.GetTemplate() != "": //nolint:staticcheck
		// Fallback for unexpected case
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
		return nil, nil // No content to parse
	}

	var responseBytes []byte
	if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
		responseBytes = []byte(textContent.Text)
	} else {
		// Fallback for other content types - marshal the whole content part
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
		// If unmarshalling to a map fails, return the raw string content
		return string(responseBytes), nil //nolint:nilerr // intentional fallback for non-JSON responses
	}

	return resultMap, nil
}

// OpenAPITool implements the Tool interface for a tool defined in an OpenAPI
// specification. It constructs and sends an HTTP request based on the OpenAPI
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
// Parameters:
//   - tool: The protobuf definition of the tool.
//   - client: The HTTP client for requests.
//   - parameterDefs: Mapping of parameter names to their locations (path, query, etc.).
//   - method: The HTTP method (GET, POST, etc.).
//   - url: The URL template.
//   - authenticator: The authenticator for upstream requests.
//   - callDefinition: The configuration for the OpenAPI call.
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

	// Cache templates
	if it := t.inputTransformer; it != nil && it.GetTemplate() != "" { //nolint:staticcheck
		tpl, err := transformer.NewTemplate(it.GetTemplate(), "{{", "}}") //nolint:staticcheck
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
func (t *OpenAPITool) Tool() *v1.Tool {
	return t.tool
}

// MCPTool returns the MCP-compliant tool definition.
//
// It lazily converts the internal protobuf definition to the MCP format on first access.
//
// Returns:
//   - *mcp.Tool: The MCP tool definition.
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
func (t *OpenAPITool) GetCacheConfig() *configv1.CacheConfig {
	return t.cache
}

// Execute handles the execution of the OpenAPI tool. It constructs an HTTP
// request based on the operation's method, URL, and parameter definitions,
// sends the request, and processes the response, applying transformations as
// needed.
func (t *OpenAPITool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) { //nolint:gocyclo
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

	// ⚡ Bolt: Use json-iterator
	decoder := fastJSON.NewDecoder(bytes.NewReader(req.ToolInputs))
	decoder.UseNumber()
	if err := decoder.Decode(&inputs); err != nil {
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
		case t.inputTransformer != nil && t.inputTransformer.GetTemplate() != "": //nolint:staticcheck
			// Fallback for unexpected case
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
		return nil, fmt.Errorf("failed to execute http request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	maxSize := getMaxHTTPResponseSize()
	// Read up to maxSize + 1 to detect if it exceeds the limit
	reader := io.LimitReader(resp.Body, maxSize+1)
	respBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read http response body: %w", err)
	}
	if int64(len(respBody)) > maxSize {
		return nil, fmt.Errorf("response body exceeds maximum size of %d bytes", maxSize)
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

// NewCommandTool creates a new CommandTool instance.
//
// Parameters:
//   - tool: The protobuf definition of the tool.
//   - service: The configuration of the command-line service.
//   - callDefinition: The configuration for the command-line call.
//   - policies: The security policies.
//   - callID: The unique identifier for the call.
//
// Returns:
//   - Tool: The created CommandTool.
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
	sandboxArgs    []string
	initError      error
}

// NewLocalCommandTool creates a new LocalCommandTool instance.
//
// Parameters:
//   - tool: The protobuf definition of the tool.
//   - service: The configuration of the command-line service.
//   - callDefinition: The configuration for the command-line call.
//   - policies: The security policies.
//   - callID: The unique identifier for the call.
//
// Returns:
//   - Tool: The created LocalCommandTool.
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

	// Check if the command is sed and supports sandbox
	cmd := service.GetCommand()
	base := filepath.Base(cmd)
	if base == "sed" || base == "gsed" {
		// Check if sed supports --sandbox by running `sed --sandbox --version`
		// We use exec.Command directly here.
		// Use --version because it exits successfully if supported.
		// If --sandbox is not supported, sed usually exits with error "illegal option"
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		checkCmd := exec.CommandContext(ctx, cmd, "--sandbox", "--version") //nolint:gosec // Trusted command from config
		if err := checkCmd.Run(); err == nil {
			t.sandboxArgs = []string{"--sandbox"}
			logging.GetLogger().Info("Enabled sandbox mode for sed tool", "tool", tool.GetName())
		} else {
			t.initError = fmt.Errorf("sed tool %q detected but --sandbox is not supported (error: %v); execution blocked for security", tool.GetName(), err)
			logging.GetLogger().Error("Failed to enable sandbox for sed", "tool", tool.GetName(), "error", err)
		}
	}

	return t
}

// Tool returns the protobuf definition of the command-line tool.
//
// Returns:
//   - *v1.Tool: The underlying protobuf definition.
func (t *LocalCommandTool) Tool() *v1.Tool {
	return t.tool
}

// MCPTool returns the MCP-compliant tool definition.
//
// It lazily converts the internal protobuf definition to the MCP format on first access.
//
// Returns:
//   - *mcp.Tool: The MCP tool definition.
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
//
// Returns:
//   - *configv1.CacheConfig: The cache configuration, if any.
func (t *LocalCommandTool) GetCacheConfig() *configv1.CacheConfig {
	if t.callDefinition == nil {
		return nil
	}
	return t.callDefinition.GetCache()
}

// Execute handles the execution of the command-line tool. It constructs a command
// with arguments and environment variables derived from the tool inputs, runs
// the command, and returns its output.
func (t *LocalCommandTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) { //nolint:gocyclo
	if t.initError != nil {
		return nil, t.initError
	}
	if logging.GetLogger().Enabled(ctx, slog.LevelDebug) {
		logging.GetLogger().Debug("executing tool", "tool", req.ToolName, "inputs", prettyPrint(req.ToolInputs, contentTypeJSON))
	}

	if allowed, err := EvaluateCompiledCallPolicy(t.policies, t.tool.GetName(), t.callID, req.ToolInputs); err != nil {
		return nil, fmt.Errorf("failed to evaluate call policy: %w", err)
	} else if !allowed {
		return nil, fmt.Errorf("tool execution blocked by policy")
	}
	var inputs map[string]any
	// Handle empty inputs by treating them as empty JSON object
	if len(bytes.TrimSpace(req.ToolInputs)) == 0 {
		req.ToolInputs = []byte("{}")
	}

	// ⚡ Bolt: Use json-iterator
	decoder := fastJSON.NewDecoder(bytes.NewReader(req.ToolInputs))
	decoder.UseNumber()
	if err := decoder.Decode(&inputs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool inputs: %w", err)
	}

	args := []string{}
	if len(t.sandboxArgs) > 0 {
		args = append(args, t.sandboxArgs...)
	}

	if t.callDefinition.GetArgs() != nil {
		args = append(args, t.callDefinition.GetArgs()...)
	}

	// Determine execution environment early for validation
	isDocker := t.service.GetContainerEnvironment() != nil && t.service.GetContainerEnvironment().GetImage() != ""

	// Substitute placeholders in args with input values
	if inputs != nil {
		for i, arg := range args {
			for k, v := range inputs {
				placeholder := "{{" + k + "}}"
				if strings.Contains(arg, placeholder) {
					val := util.ToString(v)
					if err := validateSafePathAndInjection(val, isDocker); err != nil {
						return nil, fmt.Errorf("parameter %q: %w", k, err)
					}
					// If running a shell, validate that inputs are safe for shell execution
					if isShellCommand(t.service.GetCommand()) {
						if err := checkForShellInjection(val, arg, placeholder, t.service.GetCommand()); err != nil {
							return nil, fmt.Errorf("parameter %q: %w", k, err)
						}
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
						if err := validateSafePathAndInjection(argStr, isDocker); err != nil {
							return nil, fmt.Errorf("args parameter: %w", err)
						}
						// If running a shell, validate that inputs are safe for shell execution
						if isShellCommand(t.service.GetCommand()) {
							if err := checkForShellInjection(argStr, "", "", t.service.GetCommand()); err != nil {
								return nil, fmt.Errorf("args parameter: %w", err)
							}
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
	// Only inherit safe environment variables from host for local execution
	// For Docker execution, we should not inherit host environment variables at all
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

	// Collect secrets for redaction
	secrets := make([]string, 0, len(resolvedServiceEnv))
	for _, v := range resolvedServiceEnv {
		secrets = append(secrets, v)
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
			secrets = append(secrets, secretValue)
			env = append(env, fmt.Sprintf("%s=%s", name, secretValue))
		} else if val, ok := inputs[name]; ok {
			valStr := util.ToString(val)
			if err := validateSafePathAndInjection(valStr, isDocker); err != nil {
				return nil, fmt.Errorf("parameter %q: %w", name, err)
			}
			env = append(env, fmt.Sprintf("%s=%s", name, valStr))
		}
	}

	if req.DryRun {
		logging.GetLogger().Info("Dry run execution", "tool", req.ToolName)
		return map[string]any{
			"dry_run": true,
			"request": map[string]any{
				"command": t.service.GetCommand(),
				"args":    args,
				"env":     env,
			},
		}, nil
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
		decoder := fastJSON.NewDecoder(bytes.NewReader(req.ToolInputs))
		decoder.UseNumber()
		if err := decoder.Decode(&unmarshaledInputs); err != nil {
			_ = stdin.Close()
			return nil, fmt.Errorf("failed to unmarshal tool inputs: %w", err)
		}

		// Write inputs to stdin in a separate goroutine to avoid deadlock if the command crashes
		go func() {
			defer func() { _ = stdin.Close() }()
			if err := fastJSON.NewEncoder(stdin).Encode(unmarshaledInputs); err != nil {
				logging.GetLogger().Warn("Failed to encode inputs to stdin", "error", err)
			}
		}()

		var result map[string]interface{}
		if err := fastJSON.NewDecoder(io.LimitReader(stdout, limit)).Decode(&result); err != nil {
			<-stderrDone
			return nil, fmt.Errorf("failed to execute JSON CLI command: %w. Stderr: %s", err, stderrBuf.String())
		}
		return result, nil
	}

	stdout, stderr, exitCodeChan, err := executor.Execute(ctx, t.service.GetCommand(), args, t.service.GetWorkingDirectory(), env)
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	var combinedBuf threadSafeBuffer
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
		"stdout":          util.RedactSecrets(stdoutBuf.String(), secrets),
		"stderr":          util.RedactSecrets(stderrBuf.String(), secrets),
		"combined_output": util.RedactSecrets(combinedBuf.String(), secrets),
		"start_time":      startTime,
		"end_time":        endTime,
		"return_code":     exitCode,
		"status":          status,
	}

	return result, nil
}

// Tool returns the protobuf definition of the command-line tool.
//
// Returns:
//   - *v1.Tool: The underlying protobuf definition.
func (t *CommandTool) Tool() *v1.Tool {
	return t.tool
}

// MCPTool returns the MCP-compliant tool definition.
//
// It lazily converts the internal protobuf definition to the MCP format on first access.
//
// Returns:
//   - *mcp.Tool: The MCP tool definition.
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
//
// Returns:
//   - *configv1.CacheConfig: The cache configuration, if any.
func (t *CommandTool) GetCacheConfig() *configv1.CacheConfig {
	if t.callDefinition == nil {
		return nil
	}
	return t.callDefinition.GetCache()
}

// Execute handles the execution of the command-line tool. It constructs a command
// with arguments and environment variables derived from the tool inputs, runs
// the command, and returns its output.
func (t *CommandTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) { //nolint:gocyclo
	if t.initError != nil {
		return nil, t.initError
	}
	if logging.GetLogger().Enabled(ctx, slog.LevelDebug) {
		logging.GetLogger().Debug("executing tool", "tool", req.ToolName, "inputs", prettyPrint(req.ToolInputs, contentTypeJSON))
	}

	if allowed, err := EvaluateCompiledCallPolicy(t.policies, t.tool.GetName(), t.callID, req.ToolInputs); err != nil {
		return nil, fmt.Errorf("failed to evaluate call policy: %w", err)
	} else if !allowed {
		return nil, fmt.Errorf("tool execution blocked by policy")
	}
	var inputs map[string]any
	// Handle empty inputs by treating them as empty JSON object
	if len(bytes.TrimSpace(req.ToolInputs)) == 0 {
		req.ToolInputs = []byte("{}")
	}

	// ⚡ Bolt: Use json-iterator
	decoder := fastJSON.NewDecoder(bytes.NewReader(req.ToolInputs))
	decoder.UseNumber()
	if err := decoder.Decode(&inputs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool inputs: %w", err)
	}

	args := []string{}
	if t.callDefinition.GetArgs() != nil {
		args = append(args, t.callDefinition.GetArgs()...)
	}

	// Determine execution environment early for validation
	isDocker := t.service.GetContainerEnvironment() != nil && t.service.GetContainerEnvironment().GetImage() != ""

	// Substitute placeholders in args with input values
	if inputs != nil {
		for i, arg := range args {
			for k, v := range inputs {
				placeholder := "{{" + k + "}}"
				if strings.Contains(arg, placeholder) {
					val := util.ToString(v)
					if err := checkForPathTraversal(val); err != nil {
						return nil, fmt.Errorf("parameter %q: %w", k, err)
					}
					if !isDocker {
						if err := checkForLocalFileAccess(val); err != nil {
							return nil, fmt.Errorf("parameter %q: %w", k, err)
						}
					}
					if err := checkForArgumentInjection(val); err != nil {
						return nil, fmt.Errorf("parameter %q: %w", k, err)
					}
					// If running a shell, validate that inputs are safe for shell execution
					if isShellCommand(t.service.GetCommand()) {
						if err := checkForShellInjection(val, arg, placeholder, t.service.GetCommand()); err != nil {
							return nil, fmt.Errorf("parameter %q: %w", k, err)
						}
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
						if !isDocker {
							if err := checkForLocalFileAccess(argStr); err != nil {
								return nil, fmt.Errorf("args parameter: %w", err)
							}
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

	// Collect secrets for redaction
	secrets := make([]string, 0, len(resolvedServiceEnv))
	for _, v := range resolvedServiceEnv {
		secrets = append(secrets, v)
	}

	for k, v := range resolvedServiceEnv {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	if ce := t.service.GetContainerEnvironment(); ce != nil {
		resolvedContainerEnv, err := util.ResolveSecretMap(ctx, ce.GetEnv(), nil)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve container env: %w", err)
		}
		for _, v := range resolvedContainerEnv {
			secrets = append(secrets, v)
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
			secrets = append(secrets, secretValue)
			env = append(env, fmt.Sprintf("%s=%s", name, secretValue))
		} else if val, ok := inputs[name]; ok {
			valStr := util.ToString(val)
			if err := checkForPathTraversal(valStr); err != nil {
				return nil, fmt.Errorf("parameter %q: %w", name, err)
			}
			if !isDocker {
				if err := checkForLocalFileAccess(valStr); err != nil {
					return nil, fmt.Errorf("parameter %q: %w", name, err)
				}
			}
			env = append(env, fmt.Sprintf("%s=%s", name, valStr))
		}
	}

	if req.DryRun {
		logging.GetLogger().Info("Dry run execution", "tool", req.ToolName)
		return map[string]any{
			"dry_run": true,
			"request": map[string]any{
				"command": t.service.GetCommand(),
				"args":    args,
				"env":     env,
			},
		}, nil
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
		decoder := fastJSON.NewDecoder(bytes.NewReader(req.ToolInputs))
		decoder.UseNumber()
		if err := decoder.Decode(&unmarshaledInputs); err != nil {
			_ = stdin.Close()
			return nil, fmt.Errorf("failed to unmarshal tool inputs: %w", err)
		}

		// Write inputs to stdin in a separate goroutine to avoid deadlock if the command crashes
		go func() {
			defer func() { _ = stdin.Close() }()
			if err := fastJSON.NewEncoder(stdin).Encode(unmarshaledInputs); err != nil {
				logging.GetLogger().Warn("Failed to encode inputs to stdin", "error", err)
			}
		}()

		var result map[string]interface{}
		if err := fastJSON.NewDecoder(io.LimitReader(stdout, limit)).Decode(&result); err != nil {
			<-stderrDone
			return nil, fmt.Errorf("failed to execute JSON CLI command: %w. Stderr: %s", err, stderrBuf.String())
		}
		return result, nil
	}

	stdout, stderr, exitCodeChan, err := executor.Execute(ctx, t.service.GetCommand(), args, t.service.GetWorkingDirectory(), env)
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	var combinedBuf threadSafeBuffer
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
		"stdout":          util.RedactSecrets(stdoutBuf.String(), secrets),
		"stderr":          util.RedactSecrets(stderrBuf.String(), secrets),
		"combined_output": util.RedactSecrets(combinedBuf.String(), secrets),
		"start_time":      startTime,
		"end_time":        endTime,
		"return_code":     exitCode,
		"status":          status,
	}

	return result, nil
}

type threadSafeBuffer struct {
	b  bytes.Buffer
	mu sync.Mutex
}

// Write writes bytes to the buffer in a thread-safe manner.
//
// Parameters:
//   - p: The slice of bytes to write.
//
// Returns:
//   - n: The number of bytes written.
//   - err: An error if one occurred.
func (tsb *threadSafeBuffer) Write(p []byte) (n int, err error) {
	tsb.mu.Lock()
	defer tsb.mu.Unlock()
	return tsb.b.Write(p)
}

// String returns the contents of the buffer as a string in a thread-safe manner.
//
// Returns:
//   - string: The contents of the buffer.
func (tsb *threadSafeBuffer) String() string {
	tsb.mu.Lock()
	defer tsb.mu.Unlock()
	return tsb.b.String()
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
		// Use stdjson for Indent
		if err := stdjson.Indent(&prettyJSON, input, "", "  "); err == nil {
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

		var stack []string

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

			switch t := token.(type) {
			case xml.StartElement:
				// Redact attributes
				for i := range t.Attr {
					if util.IsSensitiveKey(t.Attr[i].Name.Local) {
						t.Attr[i].Value = redactedPlaceholder
					}
				}
				token = t
				stack = append(stack, t.Name.Local)
			case xml.EndElement:
				if len(stack) > 0 {
					stack = stack[:len(stack)-1]
				}
			case xml.CharData:
				if len(stack) > 0 {
					currentTag := stack[len(stack)-1]
					if util.IsSensitiveKey(currentTag) {
						token = xml.CharData([]byte(redactedPlaceholder))
					}
				}
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

// getMaxHTTPResponseSize returns the maximum size of the HTTP response body in bytes.
// It checks the MCPANY_MAX_HTTP_RESPONSE_SIZE environment variable.
func getMaxHTTPResponseSize() int64 {
	val := os.Getenv("MCPANY_MAX_HTTP_RESPONSE_SIZE")
	if val != "" {
		if size, err := strconv.ParseInt(val, 10, 64); err == nil {
			return size
		}
	}
	return consts.DefaultMaxHTTPResponseBytes
}

func isSensitiveHeader(key string) bool {
	k := strings.ToLower(key)
	if k == "authorization" || k == "proxy-authorization" || k == "cookie" || k == "set-cookie" || k == "x-api-key" {
		return true
	}
	if strings.Contains(k, "token") || strings.Contains(k, "secret") || strings.Contains(k, "password") {
		return true
	}
	if strings.Contains(k, "access-token") || strings.Contains(k, "x-auth") || strings.Contains(k, "csrf") || strings.Contains(k, "xsrf") {
		return true
	}
	if strings.Contains(k, "signature") {
		return true
	}
	return false
}

func checkForPathTraversal(val string) error {
	if val == ".." {
		return fmt.Errorf("path traversal attempt detected")
	}
	// Check for standard traversal sequences
	if strings.HasPrefix(val, "../") || strings.HasPrefix(val, "..\\") {
		return fmt.Errorf("path traversal attempt detected")
	}
	if strings.HasSuffix(val, "/..") || strings.HasSuffix(val, "\\..") {
		return fmt.Errorf("path traversal attempt detected")
	}
	if strings.Contains(val, "/../") || strings.Contains(val, "\\..\\") || strings.Contains(val, "/..\\") || strings.Contains(val, "\\../") {
		return fmt.Errorf("path traversal attempt detected")
	}

	// Also check for encoded traversal sequences often used to bypass filters
	// %2e%2e is ..
	// ⚡ Bolt Optimization: Manual scan to avoid strings.ToLower allocation
	for i := 0; i < len(val); {
		idx := strings.IndexByte(val[i:], '%')
		if idx == -1 {
			break
		}
		i += idx
		if i+5 < len(val) {
			if val[i+1] == '2' && (val[i+2]|0x20 == 'e') &&
				val[i+3] == '%' &&
				val[i+4] == '2' && (val[i+5]|0x20 == 'e') {
				return fmt.Errorf("path traversal attempt detected (encoded)")
			}
		}
		i++
	}
	return nil
}

// cleanPathPreserveDoubleSlash cleans the path like path.Clean but preserves double slashes.
// It resolves . and .. segments.
// It also trims the trailing slash (by removing the final empty segment if present),
// assuming the caller will handle restoring it if necessary.
func cleanPathPreserveDoubleSlash(p string) string {
	if p == "" {
		return "."
	}

	rooted := strings.HasPrefix(p, "/")
	parts := strings.Split(p, "/")

	var out []string

	for i, part := range parts {
		// Preserve root empty part if rooted
		if i == 0 && rooted {
			out = append(out, part) // part is ""
			continue
		}

		// Skip last empty part (trailing slash removal)
		// We only skip it if it is strictly the last part and is empty.
		if i == len(parts)-1 && part == "" {
			continue
		}

		if part == "." {
			continue
		}
		if part == ".." {
			if len(out) > 0 {
				last := out[len(out)-1]
				if rooted {
					// At root, ignore ..
					// Root is [""] or ["", "", ...]
					// If out is just [""], we are at /.
					if len(out) == 1 && out[0] == "" {
						continue
					}
					// If out ends in "", it might be //
					// If we are at // (i.e. ["", ""]), we should stay at // (ignore ..)
					if len(out) == 2 && out[1] == "" {
						continue
					}
					// Pop it
					out = out[:len(out)-1]
				} else {
					if last == ".." {
						out = append(out, part)
					} else {
						// Pop previous segment
						out = out[:len(out)-1]
					}
				}
			} else if !rooted {
				out = append(out, part)
			}
		} else {
			out = append(out, part)
		}
	}

	if len(out) == 0 {
		return "."
	}

	res := strings.Join(out, "/")
	// If rooted and result is empty string, return "/"
	if rooted && res == "" {
		return "/"
	}
	// If result is empty string but not rooted, return "."
	if res == "" {
		return "."
	}

	return res
}

func checkForLocalFileAccess(val string) error {
	if filepath.IsAbs(val) {
		return fmt.Errorf("absolute path detected: %s (only relative paths are allowed for local execution)", val)
	}
	// Also block "file:" scheme to prevent SSRF/LFI (e.g. curl file:///etc/passwd)
	// We check for "file:" prefix case-insensitively.
	if strings.HasPrefix(strings.ToLower(val), "file:") {
		return fmt.Errorf("file: scheme detected: %s (local file access is not allowed)", val)
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

func isShellCommand(cmd string) bool {
	// Shells and commands that invoke a shell or execute code remotely/locally via shell.
	// We treat them as shells to enforce strict argument validation.
	shells := []string{
		"sh", "bash", "zsh", "dash", "ash", "ksh", "csh", "tcsh", "fish",
		"pwsh", "powershell", "powershell.exe", "pwsh.exe", "cmd", "cmd.exe",
		"ssh", "scp", "su", "sudo", "env",
		"busybox", "expect", "watch", "tmux", "screen",
		// Common interpreters and runners that can execute code
		"python", "python2", "python3",
		"ruby", "perl", "php",
		"node", "nodejs", "bun", "deno",
		"lua", "awk", "gawk", "nawk", "mawk", "sed",
		"jq",
		"psql", "mysql", "sqlite3",
		"docker",
		// Additional shells/runners found missing
		"busybox", "expect", "tclsh", "wish",
		"irb", "php-cgi", "perl5",
		"openssl", "git", "hg", "svn",
		"wget", "curl", "nc", "netcat", "ncat",
		"socat", "telnet",
		// Editors and pagers that can execute commands
		"vi", "vim", "nvim", "emacs", "nano",
		"less", "more", "man",
		// Build tools and others that can execute commands
		"tar", "find", "xargs", "tee",
		"make", "rake", "ant", "mvn", "gradle",
		"npm", "yarn", "pnpm", "npx", "bunx", "go", "cargo", "pip",
		// Cloud/DevOps tools that can execute commands or have sensitive flags
		"kubectl", "helm", "aws", "gcloud", "az", "terraform", "ansible", "ansible-playbook",
		// Additional interpreters and compilers that can execute code
		"R", "Rscript", "julia", "groovy", "jshell",
		"scala", "kotlin", "swift",
		"elixir", "iex", "erl", "escript",
		"ghci", "clisp", "sbcl", "lisp", "scheme", "racket",
		"lua5.1", "lua5.2", "lua5.3", "lua5.4", "luajit",
		"gcc", "g++", "clang", "java",
		// Additional dangerous tools
		"zip", "unzip", "rsync", "nmap", "tcpdump", "gdb", "lldb",
	}
	base := filepath.Base(cmd)
	for _, shell := range shells {
		if base == shell {
			return true
		}
		// Check for versioned binaries (e.g. python3.10, ruby2.7)
		if strings.HasPrefix(base, shell) {
			suffix := base[len(shell):]
			if isVersionSuffix(suffix) {
				return true
			}
		}
	}

	// Check for script extensions that indicate shell execution or interpretation
	ext := strings.ToLower(filepath.Ext(base))
	scriptExts := []string{
		".sh", ".bash", ".zsh", ".ash", ".ksh", ".csh", ".tcsh", ".fish",
		".bat", ".cmd", ".ps1", ".vbs", ".js", ".mjs", ".ts",
		".py", ".pyc", ".pyo", ".pyd",
		".rb", ".pl", ".pm", ".php",
		".lua", ".r",
	}
	for _, scriptExt := range scriptExts {
		if ext == scriptExt {
			return true
		}
	}

	return false
}

func isVersionSuffix(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if (r < '0' || r > '9') && r != '.' && r != '-' {
			return false
		}
	}
	return true
}

func checkForShellInjection(val string, template string, placeholder string, command string) error {
	// Determine the quoting context of the placeholder in the template
	quoteLevel := analyzeQuoteContext(template, placeholder)

	base := strings.ToLower(filepath.Base(command))
	isWindowsCmd := base == "cmd.exe" || base == "cmd"
	if isWindowsCmd && quoteLevel == 2 {
		quoteLevel = 0
	}

	// Sentinel Security Update: Interpreter Injection Protection
	if isInterpreter(command) {
		if err := checkInterpreterInjection(val, template, base, quoteLevel); err != nil {
			return err
		}
	}

	if quoteLevel == 3 { // Backticked
		return checkBacktickInjection(val, command)
	}

	if quoteLevel == 2 { // Single Quoted
		if strings.Contains(val, "'") {
			return fmt.Errorf("shell injection detected: value contains single quote which breaks out of single-quoted argument")
		}

		// Sentinel Security Update:
		// Even if single-quoted, if the shell command invokes an interpreter (like awk, perl, ruby, python),
		// the content inside the quotes might be interpreted as code.
		// Since we cannot know if the inner command is an interpreter, we explicitly block common RCE patterns.

		// Block backticks (used by Perl, Ruby, PHP for execution)
		if strings.Contains(val, "`") {
			return fmt.Errorf("shell injection detected: value contains backtick inside single-quoted argument (potential interpreter abuse)")
		}

		// Block dangerous function calls (system, exec, popen, eval) followed by open parenthesis
		// We use a case-insensitive check for robustness, although most interpreters are case-sensitive.
		// We normalize by removing whitespace to detect "system (" or "system\t(".
		var b strings.Builder
		b.Grow(len(val))
		for _, r := range val {
			if !unicode.IsSpace(r) {
				b.WriteRune(r)
			}
		}
		cleanVal := strings.ToLower(b.String())

		dangerousCalls := []string{"system(", "exec(", "popen(", "eval("}
		for _, call := range dangerousCalls {
			if strings.Contains(cleanVal, call) {
				return fmt.Errorf("shell injection detected: value contains dangerous function call %q inside single-quoted argument (potential interpreter abuse)", call)
			}
		}

		return nil
	}

	if quoteLevel == 1 { // Double Quoted
		// In double quotes, dangerous characters are double quote, $, and backtick
		// We also need to block backslash because it can be used to escape the closing quote
		// % is also dangerous in Windows CMD inside double quotes
		if idx := strings.IndexAny(val, "\"$`\\%"); idx != -1 {
			return fmt.Errorf("shell injection detected: value contains dangerous character %q inside double-quoted argument", val[idx])
		}
		return nil
	}

	return checkUnquotedInjection(val, command)
}

func checkInterpreterInjection(val, template, base string, quoteLevel int) error {
	// Python: Check for f-string prefix in template
	if strings.HasPrefix(base, "python") {
		// Scan template to find the prefix of the quote containing the placeholder
		// Given complexity, we use a heuristic: if template contains f" or f', enforce checks.
		hasFString := false
		for i := 0; i < len(template)-1; i++ {
			if template[i+1] == '\'' || template[i+1] == '"' {
				prefix := strings.ToLower(getPrefix(template, i+1))
				if prefix == "f" || prefix == "fr" || prefix == "rf" {
					hasFString = true
					break
				}
			}
		}
		if hasFString {
			if strings.ContainsAny(val, "{}") {
				return fmt.Errorf("python f-string injection detected: value contains '{' or '}'")
			}
		}
	}

	// Ruby: #{...} works in double quotes AND backticks
	if strings.HasPrefix(base, "ruby") && (quoteLevel == 1 || quoteLevel == 3) { // Double Quoted or Backticked
		if strings.Contains(val, "#{") {
			return fmt.Errorf("ruby interpolation injection detected: value contains '#{'")
		}
	}

	// Node/JS/Perl/PHP: ${...} works in backticks (JS) or double quotes (Perl/PHP)
	isNode := strings.HasPrefix(base, "node") || base == "bun" || base == "deno"
	isPerl := strings.HasPrefix(base, "perl")
	isPhp := strings.HasPrefix(base, "php")

	if isNode && quoteLevel == 3 { // Backtick
		if strings.Contains(val, "${") {
			return fmt.Errorf("javascript template literal injection detected: value contains '${'")
		}
	}
	// Perl and PHP interpolate variables in both double quotes and backticks
	if (isPerl || isPhp) && (quoteLevel == 1 || quoteLevel == 3) { // Double Quoted or Backticked
		if strings.Contains(val, "${") {
			return fmt.Errorf("variable interpolation injection detected: value contains '${'")
		}
	}

	// Awk: Block pipe | to prevent external command execution
	isAwk := strings.HasPrefix(base, "awk") || strings.HasPrefix(base, "gawk") || strings.HasPrefix(base, "nawk") || strings.HasPrefix(base, "mawk")
	if isAwk {
		if strings.Contains(val, "|") {
			return fmt.Errorf("awk injection detected: value contains '|'")
		}
	}

	return nil
}

func checkBacktickInjection(val, command string) error {
	// Backticks in Shell are command substitution (Level 0 danger).
	// Unless it is a known interpreter that uses backticks safely (like JS template literals),
	// we must enforce strict checks.
	if !isInterpreter(command) {
		const dangerousChars = ";|&$`(){}!<>\"\n\r\t\v\f*?[]~#%^'\\ "
		if idx := strings.IndexAny(val, dangerousChars); idx != -1 {
			return fmt.Errorf("shell injection detected: value contains dangerous character %q inside backticks", val[idx])
		}
	}
	// For interpreters (like JS), we already handled specific injections above.
	// We should still prevent breaking out of backticks.
	if strings.Contains(val, "`") {
		return fmt.Errorf("backtick injection detected")
	}
	return nil
}

func checkUnquotedInjection(val, command string) error {
	// Unquoted (or unknown quoting): strict check
	// Block common shell metacharacters and globbing/expansion characters
	// % and ^ are Windows CMD metacharacters
	// We also block quotes and backslashes to prevent argument splitting and interpretation abuse
	// We also block control characters that could act as separators or cause confusion (\r, \t, \v, \f)
	// Sentinel Security Update: Added space (' ') to block list to prevent argument injection in shell commands
	const dangerousChars = ";|&$`(){}!<>\"\n\r\t\v\f*?[]~#%^'\\ "

	charsToCheck := dangerousChars
	// For 'env' command, '=' is dangerous as it allows setting arbitrary environment variables
	if filepath.Base(command) == "env" {
		charsToCheck += "="
	}

	if idx := strings.IndexAny(val, charsToCheck); idx != -1 {
		return fmt.Errorf("shell injection detected: value contains dangerous character %q", val[idx])
	}
	return nil
}

func isInterpreter(command string) bool {
	base := strings.ToLower(filepath.Base(command))
	interpreters := []string{"python", "ruby", "perl", "php", "node", "nodejs", "bun", "deno", "lua", "java", "R", "julia", "elixir", "go", "awk", "gawk", "nawk", "mawk"}
	for _, interp := range interpreters {
		if base == interp || strings.HasPrefix(base, interp) {
			return true
		}
	}
	return false
}

func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

func getPrefix(s string, idx int) string {
	// idx is index of quote char
	start := idx - 1
	for start >= 0 {
		c := s[start]
		if !isWordChar(c) {
			break
		}
		start--
	}
	return s[start+1 : idx]
}

func analyzeQuoteContext(template, placeholder string) int {
	if template == "" || placeholder == "" {
		return 0
	}

	// Levels: 0 = Unquoted (Strict), 1 = Double, 2 = Single, 3 = Backtick
	minLevel := 3

	inSingle := false
	inDouble := false
	inBacktick := false
	escaped := false

	foundAny := false

	for i := 0; i < len(template); i++ {
		// Check if we match placeholder at current position
		if strings.HasPrefix(template[i:], placeholder) {
			foundAny = true
			currentLevel := 0
			switch {
			case inSingle:
				currentLevel = 2
			case inDouble:
				currentLevel = 1
			case inBacktick:
				currentLevel = 3
			}

			if currentLevel < minLevel {
				minLevel = currentLevel
			}

			// Advance past placeholder
			i += len(placeholder) - 1
			continue
		}

		char := template[i]

		if escaped {
			escaped = false
			continue
		}

		if char == '\\' && !inSingle {
			escaped = true
			continue
		}

		switch char {
		case '\'':
			if !inDouble && !inBacktick {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle && !inBacktick {
				inDouble = !inDouble
			}
		case '`':
			if !inSingle && !inDouble {
				inBacktick = !inBacktick
			}
		}
	}

	if !foundAny {
		return 0 // Should not happen if called correctly, fallback to strict
	}

	return minLevel
}

func validateSafePathAndInjection(val string, isDocker bool) error {
	if err := checkForPathTraversal(val); err != nil {
		return err
	}
	// Also check decoded value just in case the input was already encoded
	if decodedVal, err := url.QueryUnescape(val); err == nil && decodedVal != val {
		if err := checkForPathTraversal(decodedVal); err != nil {
			return fmt.Errorf("%w (decoded)", err)
		}
	}

	if !isDocker {
		if err := checkForLocalFileAccess(val); err != nil {
			return err
		}
		// Also check decoded value for local file access (e.g. %66ile://)
		if decodedVal, err := url.QueryUnescape(val); err == nil && decodedVal != val {
			if err := checkForLocalFileAccess(decodedVal); err != nil {
				return fmt.Errorf("%w (decoded)", err)
			}
		}
	}

	if err := checkForArgumentInjection(val); err != nil {
		return err
	}
	// Also check decoded value for argument injection (e.g. %2drf)
	if decodedVal, err := url.QueryUnescape(val); err == nil && decodedVal != val {
		if err := checkForArgumentInjection(decodedVal); err != nil {
			return fmt.Errorf("%w (decoded)", err)
		}
	}
	return nil
}
