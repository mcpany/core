// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"bytes"
	"context"

	stdjson "encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/metrics"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/resilience"
	"github.com/mcpany/core/server/pkg/transformer"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// HTTPTool implements the Tool interface for a tool exposed via an HTTP endpoint.
//
// Summary: Tool implementation for HTTP services.
//
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
	secretParams      map[string]bool

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
// Summary: Initializes a new HTTPTool.
//
// Parameters:
//   - tool: *v1.Tool. The protobuf definition of the tool.
//   - poolManager: *pool.Manager. The connection pool manager for HTTP connections.
//   - serviceID: string. The identifier for the service.
//   - authenticator: auth.UpstreamAuthenticator. The authenticator for upstream requests.
//   - callDefinition: *configv1.HttpCallDefinition. The configuration for the HTTP call.
//   - cfg: *configv1.ResilienceConfig. The resilience configuration.
//   - policies: []*configv1.CallPolicy. The security policies for the call.
//   - callID: string. The unique identifier for the call.
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
		secretParams:      make(map[string]bool),
	}

	for _, param := range callDefinition.GetParameters() {
		if param.GetSecret() != nil {
			t.secretParams[param.GetSchema().GetName()] = true
		}
	}

	compiled, err := CompileCallPolicies(policies)
	if err != nil {
		t.initError = fmt.Errorf("failed to compile call policies: %w", err)
	}
	t.policies = compiled

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
func (t *HTTPTool) Tool() *v1.Tool {
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
func (t *HTTPTool) GetCacheConfig() *configv1.CacheConfig {
	return t.cache
}

// Execute handles the execution of the HTTP tool.
//
// Summary: Executes the HTTP tool call.
//
// It builds an HTTP request by mapping input parameters to the path, query, and body,
// applies any configured transformations, sends the request, and processes the response.
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
//   - Updates metrics.
//   - Logs execution details.
//
// IsStreaming returns true if the tool supports streaming.
//
// Summary: Checks if the tool supports streaming execution.
//
// Returns:
//   - bool: True if streaming is supported.
func (t *HTTPTool) IsStreaming() bool {
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
func (t *HTTPTool) StreamExecute(ctx context.Context, req *ExecutionRequest) (<-chan any, error) {
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

// Execute handles the execution of the HTTP tool.
//
// Summary: Executes the HTTP tool call.
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
//   - Returns an error if policy evaluation fails or blocks execution.
//   - Returns an error if the http pool is not found.
//   - Returns an error if getting a client from the pool fails.
//   - Returns an error if input validation or body preparation fails.
//   - Returns an error if the HTTP request fails.
//
// Side Effects:
//   - Makes an HTTP request to the upstream service.
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

	inputs, urlString, redactedURLString, inputsModified, err := t.prepareInputsAndURL(ctx, req)
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
			t.logRequest(ctx, httpReq, bodyForAttempt, redactedURLString)
		}

		attemptResp, err := httpClient.Do(httpReq)
		if err != nil {

			logging.GetLogger().ErrorContext(ctx, "Failed to execute HTTP request", "tool", t.tool.GetName(), "error", err)
			return fmt.Errorf("failed to execute http request")
		}

		if attemptResp.StatusCode == http.StatusTooManyRequests {
			_ = attemptResp.Body.Close()
			return fmt.Errorf("upstream HTTP request failed with status %d (Too Many Requests)", attemptResp.StatusCode)
		}

		if attemptResp.StatusCode >= 400 {

			bodyBytes, _ := io.ReadAll(io.LimitReader(attemptResp.Body, 1024))
			_ = attemptResp.Body.Close()

			bodyBytes = util.RedactJSON(bodyBytes)
			bodyStr := string(bodyBytes)

			logging.GetLogger().DebugContext(ctx, "Upstream HTTP error", "status", attemptResp.StatusCode, "body", bodyStr, "url", redactedURLString)

			displayBody := bodyStr
			const maxErrorBodyLen = 200
			if len(displayBody) > maxErrorBodyLen {
				displayBody = displayBody[:maxErrorBodyLen] + "... (truncated)"
			}

			isDebug := os.Getenv("MCPANY_DEBUG") == trueStr
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

func (t *HTTPTool) logRequest(ctx context.Context, httpReq *http.Request, body io.Reader, redactedURLString string) {
	parsedURL, err := url.Parse(redactedURLString)
	var pathAndQuery string
	if err == nil {
		pathAndQuery = parsedURL.Path
		if parsedURL.RawQuery != "" {
			pathAndQuery += "?" + parsedURL.RawQuery
		}
	} else {
		pathAndQuery = httpReq.URL.Path
	}

	// Log headers
	var headerBuf bytes.Buffer
	headerBuf.WriteString(fmt.Sprintf("%s %s %s\n", httpReq.Method, pathAndQuery, httpReq.Proto))
	headerBuf.WriteString(fmt.Sprintf("Host: %s\n", httpReq.Host))
	for k, v := range httpReq.Header {
		val := strings.Join(v, ", ")
		if isSensitiveHeader(k) {
			val = redactedPlaceholder
		}
		headerBuf.WriteString(fmt.Sprintf("%s: %s\n", k, val))
	}
	logging.GetLogger().DebugContext(ctx, "sending http request headers", "headers", headerBuf.String())

	if body != nil {
		contentType := httpReq.Header.Get("Content-Type")
		bodyBytes, _ := io.ReadAll(body)

		if seeker, ok := body.(io.Seeker); ok {
			_, _ = seeker.Seek(0, io.SeekStart)
		}
		logging.GetLogger().DebugContext(ctx, "sending http request body", "body", prettyPrint(bodyBytes, contentType))
	}
}

func (t *HTTPTool) prepareInputsAndURL(ctx context.Context, req *ExecutionRequest) (map[string]any, string, string, bool, error) {
	var inputs map[string]any
	if len(req.ToolInputs) > 0 {

		req.ToolInputs = bytes.TrimSpace(req.ToolInputs)
	}

	if len(req.ToolInputs) > 0 {

		if err := fastJSONNumber.Unmarshal(req.ToolInputs, &inputs); err != nil {
			return nil, "", "", false, fmt.Errorf("failed to unmarshal tool inputs: %w (inputs: %q)", err, string(req.ToolInputs))
		}
	}

	filtered := false
	for k := range inputs {
		if !t.allowedParams[k] {
			delete(inputs, k)
			filtered = true
		}
	}

	pathReplacements, queryReplacements, inputsModified, err := t.processParameters(ctx, inputs)
	if err != nil {
		return nil, "", "", false, err
	}
	inputsModified = inputsModified || filtered

	var pathBuf strings.Builder
	var redactedPathBuf strings.Builder

	for _, seg := range t.pathSegments {
		if seg.isParam {
			if val, ok := pathReplacements[seg.value]; ok {
				pathBuf.WriteString(val)
				if t.secretParams[seg.value] {
					redactedPathBuf.WriteString(redactedPlaceholder)
				} else {
					redactedPathBuf.WriteString(val)
				}
			} else {
				pathBuf.WriteString("{{" + seg.value + "}}")
				redactedPathBuf.WriteString("{{" + seg.value + "}}")
			}
		} else {
			pathBuf.WriteString(seg.value)
			redactedPathBuf.WriteString(seg.value)
		}
	}
	pathStr := pathBuf.String()
	redactedPathStr := redactedPathBuf.String()

	var queryBuf strings.Builder
	var redactedQueryBuf strings.Builder

	for _, seg := range t.querySegments {
		if seg.isParam {
			if val, ok := queryReplacements[seg.value]; ok {
				queryBuf.WriteString(val)
				if t.secretParams[seg.value] {
					redactedQueryBuf.WriteString(redactedPlaceholder)
				} else {
					redactedQueryBuf.WriteString(val)
				}
			} else {
				queryBuf.WriteString("{{" + seg.value + "}}")
				redactedQueryBuf.WriteString("{{" + seg.value + "}}")
			}
		} else {
			queryBuf.WriteString(seg.value)
			redactedQueryBuf.WriteString(seg.value)
		}
	}
	queryStr := queryBuf.String()
	redactedQueryStr := redactedQueryBuf.String()

	hadTrailingSlash := strings.HasSuffix(pathStr, "/")

	wasRootDoubleSlash := pathStr == "//"

	pathStr = cleanPathPreserveDoubleSlash(pathStr)
	if hadTrailingSlash && (pathStr != "/" || wasRootDoubleSlash) {
		pathStr += "/"
	}

	hadTrailingSlashRedacted := strings.HasSuffix(redactedPathStr, "/")
	redactedPathStr = cleanPathPreserveDoubleSlash(redactedPathStr)
	if hadTrailingSlashRedacted && (redactedPathStr != "/" || wasRootDoubleSlash) {
		redactedPathStr += "/"
	}

	buildURL := func(pStr, qStr string, redactUser bool) string {
		var buf strings.Builder
		if t.cachedURL.Scheme != "" {
			buf.WriteString(t.cachedURL.Scheme)
			buf.WriteString("://")
		}
		if t.cachedURL.User != nil {
			if redactUser {
				if t.cachedURL.User.Username() != "" {
					buf.WriteString(t.cachedURL.User.Username())
					buf.WriteString(":" + redactedPlaceholder)
					buf.WriteString("@")
				}
			} else {
				buf.WriteString(t.cachedURL.User.String())
				buf.WriteString("@")
			}
		}
		buf.WriteString(t.cachedURL.Host)
		if pStr != "" && !strings.HasPrefix(pStr, "/") {
			buf.WriteString("/")
		}
		buf.WriteString(pStr)
		if qStr != "" {
			buf.WriteString("?")
			buf.WriteString(qStr)
		}
		return buf.String()
	}

	urlString := buildURL(pathStr, queryStr, false)
	redactedURLString := buildURL(redactedPathStr, redactedQueryStr, true)

	return inputs, urlString, redactedURLString, inputsModified, nil
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

				val = ""
			} else if t.paramInPath[i] || t.paramInQuery[i] {

				delete(inputs, name)
				inputsModified = true
			}

			valStr := util.ToString(val)

			if t.paramInPath[i] {

				if err := checkForPathTraversal(valStr); err != nil {
					return nil, nil, false, fmt.Errorf("path traversal attempt detected in parameter %q: %w", name, err)
				}

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

		subparts := strings.SplitN(part, "}}", 2)

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

		data := map[string]any{
			"kind":      configv1.WebhookKind_WEBHOOK_KIND_TRANSFORM_INPUT,
			"tool_name": toolName,
			"inputs":    inputs,
		}
		respEvent, err := t.webhookClient.Call(ctx, "com.mcpany.tool.transform_input", data)
		if err != nil {
			return nil, "", fmt.Errorf("transformation webhook failed: %w", err)
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
			return nil, "", fmt.Errorf("failed to render input template: %w", err)
		}
		body = strings.NewReader(renderedBody)
		if t.cachedInputTemplate.IsJSON {
			contentType = contentTypeJSON
		}
	case t.inputTransformer != nil && t.inputTransformer.GetTemplate() != "":

		return nil, "", fmt.Errorf("input template configured but not cached (initialization error?)")
	default:

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
		return string(respBody), nil
	}

	return result, nil
}
