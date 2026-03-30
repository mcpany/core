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
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	jsoniter "github.com/json-iterator/go"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/consts"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	contentTypeJSON     = "application/json"
	redactedPlaceholder = "[REDACTED]"

	// HealthStatusUnhealthy indicates that a service is in an unhealthy state.
	// Summary: Defines HealthStatusUnhealthy.
	HealthStatusUnhealthy = "unhealthy"

	gitCommand = "git"
	trueStr    = "true"
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

// ⚡ Bolt: Global JSON decoder configuration with UseNumber enabled.
// Randomized Selection from Top 5 High-Impact Targets
// This prevents creating a new decoder on every tool execution (allocation reduction)
// while preserving the UseNumber behavior required for accurate number handling.
// Note: This config is frozen and thread-safe.
var fastJSONNumber = jsoniter.Config{
	EscapeHTML:             true,
	SortMapKeys:            true,
	ValidateJsonRawMessage: true,
	UseNumber:              true,
}.Froze()

// Tool is the fundamental interface for any executable tool in the system.
//
// Summary: Interface for defining and executing tools.
//
// Each implementation represents a different type of underlying service
// (e.g., gRPC, HTTP, command-line).
type Tool interface {
	// Tool returns the protobuf definition of the tool.
	//
	// Returns:
	//   - *v1.Tool: The protobuf tool definition.
	Tool() *v1.Tool

	// MCPTool returns the MCP tool definition.
	//
	// Returns:
	//   - *mcp.Tool: The MCP tool definition.
	MCPTool() *mcp.Tool

	// Execute runs the tool with the provided context and request, returning
	// the result or an error.
	//
	// Summary: Executes the tool.
	//
	// Parameters:
	//   - ctx: context.Context. The execution context.
	//   - req: *ExecutionRequest. The request payload.
	//
	// Returns:
	//   - any: The execution result.
	//   - error: An error if execution fails.
	//
	// Side Effects:
	//   - Executes the underlying service logic (network calls, command execution, etc.).
	Execute(ctx context.Context, req *ExecutionRequest) (any, error)

	// IsStreaming returns true if the tool supports streaming execution.
	//
	// Summary: Checks if the tool supports streaming execution.
	//
	// Returns:
	//   - bool: True if streaming is supported.
	IsStreaming() bool

	// StreamExecute runs the tool in streaming mode, returning a channel of results.
	//
	// Summary: Executes the tool in streaming mode.
	//
	// Parameters:
	//   - ctx: context.Context. The execution context.
	//   - req: *ExecutionRequest. The request payload.
	//
	// Returns:
	//   - <-chan any: A channel that emits streaming results.
	//   - error: An error if the operation fails or streaming is not supported.
	//
	// Side Effects:
	//   - Executes the underlying service logic in a streaming manner.
	StreamExecute(ctx context.Context, req *ExecutionRequest) (<-chan any, error)

	// GetCacheConfig returns the cache configuration for the tool.
	//
	// Summary: Retrieves cache configuration.
	//
	// Returns:
	//   - *configv1.CacheConfig: The cache configuration, or nil if none.
	GetCacheConfig() *configv1.CacheConfig
}

// ServiceInfo holds metadata about a registered upstream service, including its
// configuration and any associated protobuf file descriptors.
//
// Summary: Metadata for a registered service.
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

// ExecutionRequest represents a request to execute a specific tool, including
// its name and input arguments as a raw JSON message.
//
// Summary: Request payload for tool execution.
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
//
// Summary: Interface for tool and service lookup.
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
//
// Summary: Function signature for tool execution middleware.
//
// Parameters:
//   - ctx: context.Context. The execution context.
//   - req: *ExecutionRequest. The request payload.
//
// Returns:
//   - any: The execution result.
//   - error: An error if execution fails.
type ExecutionFunc func(ctx context.Context, req *ExecutionRequest) (any, error)

type contextKey string

const toolContextKey = contextKey("tool")

// NewContextWithTool creates a new context with the given tool embedded.
//
// Summary: Embeds a tool into the context.
//
// Parameters:
//   - ctx: context.Context. The context to extend.
//   - t: Tool. The tool instance to embed in the context.
//
// Returns:
//   - context.Context: A new context containing the tool.
func NewContextWithTool(ctx context.Context, t Tool) context.Context {
	return context.WithValue(ctx, toolContextKey, t)
}

// GetFromContext retrieves a tool from the context if present.
//
// Summary: Retrieves a tool from the context.
//
// Parameters:
//   - ctx: context.Context. The context to search.
//
// Returns:
//   - Tool: The tool instance from the context.
//   - bool: True if a tool was found, false otherwise.
func GetFromContext(ctx context.Context) (Tool, bool) {
	t, ok := ctx.Value(toolContextKey).(Tool)
	return t, ok
}

// Callable is an interface that represents a callable tool.
//
// Summary: Interface for executing a tool.
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

// StreamingCallable is an interface that represents a callable tool that can stream output.
//
// Summary: Interface for executing a tool with streaming output.
type StreamingCallable interface {
	Callable

	// StreamCall executes the callable with the given request, emitting updates to the channel.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - req: *ExecutionRequest. The execution request details.
	//
	// Returns:
	//   - <-chan any: A channel that emits streaming results.
	//   - error: An error if the initial operation fails.
	StreamCall(ctx context.Context, req *ExecutionRequest) (<-chan any, error)
}

// Action defines the decision made by a pre-call hook.
//
// Summary: Enumeration of possible hook actions.
type Action int

const (
	// ActionAllow indicates that the action is allowed to proceed.
	//
	// Summary: Allow action.
	ActionAllow Action = 0
	// ActionDeny indicates that the action is denied and should be blocked.
	//
	// Summary: Deny action.
	ActionDeny Action = 1
	// ActionSaveCache indicates that the result should be saved to the cache.
	//
	// Summary: Save to cache action.
	ActionSaveCache Action = 2
	// ActionDeleteCache indicates that the associated cache entry should be invalidated.
	//
	// Summary: Delete from cache action.
	ActionDeleteCache Action = 3
)

// CacheControl is a mutable struct to pass cache control instructions via context.
//
// Summary: Context-based cache control instructions.
type CacheControl struct {
	Action Action
}

const cacheControlContextKey = contextKey("cache_control")

// NewContextWithCacheControl creates a new context with the given CacheControl.
//
// Summary: Embeds CacheControl into the context.
//
// Parameters:
//   - ctx: context.Context. The context to extend.
//   - cc: *CacheControl. The CacheControl instance to embed.
//
// Returns:
//   - context.Context: A new context containing the CacheControl.
func NewContextWithCacheControl(ctx context.Context, cc *CacheControl) context.Context {
	return context.WithValue(ctx, cacheControlContextKey, cc)
}

// GetCacheControl retrieves the CacheControl from the context.
//
// Summary: Retrieves CacheControl from the context.
//
// Parameters:
//   - ctx: context.Context. The context to search.
//
// Returns:
//   - *CacheControl: The CacheControl instance if found.
//   - bool: True if CacheControl exists, false otherwise.
func GetCacheControl(ctx context.Context) (*CacheControl, bool) {
	cc, ok := ctx.Value(cacheControlContextKey).(*CacheControl)
	return cc, ok
}

// PreCallHook defines the interface for hooks executed before a tool call.
//
// Summary: Interface for pre-execution hooks.
type PreCallHook interface {
	// ExecutePre runs the hook. It returns an action (Allow/Deny),
	// a potentially modified request (or nil if unchanged), and an error.
	ExecutePre(ctx context.Context, req *ExecutionRequest) (Action, *ExecutionRequest, error)
}

// PostCallHook defines the interface for hooks executed after a tool call.
//
// Summary: Interface for post-execution hooks.
type PostCallHook interface {
	// ExecutePost runs the hook. It returns the potentially modified result
	// (or original if unchanged) and an error.
	ExecutePost(ctx context.Context, req *ExecutionRequest, result any) (any, error)
}

// Write writes bytes to the buffer in a thread-safe manner.
//
// Parameters:
//   - p: The slice of bytes to write.
//
// Returns:
//   - n: The number of bytes written.
//   - err: An error if one occurred.
//
// Summary: Updates Write operation.
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

// String returns the contents of the buffer as a string in a thread-safe manner.
//
// Returns:
//   - string: The contents of the buffer.
//
// Summary: Executes String operation.
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

// prettyPrint formats the input based on content type for better readability.
func prettyPrint(input []byte, contentType string) string {
	if len(input) == 0 {
		return ""
	}

	contentType = strings.ToLower(contentType)

	if strings.HasPrefix(contentType, "image/") ||
		strings.HasPrefix(contentType, "audio/") ||
		strings.HasPrefix(contentType, "video/") ||
		contentType == "application/octet-stream" {
		return fmt.Sprintf("[Binary Data: %d bytes]", len(input))
	}

	if strings.Contains(contentType, "json") {

		input = util.RedactJSON(input)

		var prettyJSON bytes.Buffer

		if err := stdjson.Indent(&prettyJSON, input, "", "  "); err == nil {
			return prettyJSON.String()
		}

	}

	if strings.Contains(contentType, "xml") {
		decoder := xml.NewDecoder(bytes.NewReader(input))
		var buf bytes.Buffer
		encoder := xml.NewEncoder(&buf)
		encoder.Indent("", "  ")

		var stack []string

		for {
			token, err := decoder.Token()
			if err == io.EOF {
				_ = encoder.Flush()
				return buf.String()
			}
			if err != nil {

				return string(input)
			}

			switch t := token.(type) {
			case xml.StartElement:

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

// getMaxCommandOutputSize returns the maximum size of the command output (stdout + stderr) in bytes.
// It checks the MCPANY_MAX_COMMAND_OUTPUT_SIZE environment variable.
func getMaxCommandOutputSize() int64 {
	val := os.Getenv("MCPANY_MAX_COMMAND_OUTPUT_SIZE")
	if val != "" {
		if size, err := strconv.ParseInt(val, 10, 64); err == nil {
			return size
		}

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

	if strings.HasPrefix(val, "../") || strings.HasPrefix(val, "..\\") {
		return fmt.Errorf("path traversal attempt detected")
	}
	if strings.HasSuffix(val, "/..") || strings.HasSuffix(val, "\\..") {
		return fmt.Errorf("path traversal attempt detected")
	}
	if strings.Contains(val, "/../") || strings.Contains(val, "\\..\\") || strings.Contains(val, "/..\\") || strings.Contains(val, "\\../") {
		return fmt.Errorf("path traversal attempt detected")
	}

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

var dangerousEnvVars = map[string]bool{

	"GIT_SSH": true, "GIT_SSH_COMMAND": true, "GIT_ASKPASS": true,
	"GIT_PAGER": true, "GIT_EDITOR": true, "GIT_EXTERNAL_DIFF": true,
	"GIT_MAN_VIEWER": true, "GIT_SEQUENCE_EDITOR": true,
	"GIT_CONFIG_PARAMETERS": true, "GIT_CONFIG_COUNT": true,

	"PYTHONPATH": true, "PYTHONSTARTUP": true, "PYTHONHOME": true,
	"PERL5LIB": true, "PERLIB": true, "PERL5OPT": true,
	"RUBYLIB": true, "RUBYOPT": true,
	"NODE_OPTIONS": true, "NODE_PATH": true,
	"JAVA_TOOL_OPTIONS": true, "JDK_JAVA_OPTIONS": true, "_JAVA_OPTIONS": true,
	"R_PROFILE_USER": true, "R_ENVIRON_USER": true,

	"BASH_ENV": true, "ENV": true, "PS4": true, "SHELLOPTS": true, "PROMPT_COMMAND": true, "IFS": true,

	"GCONV_PATH": true, "SHELL": true, "HOME": true,
	"XDG_CONFIG_HOME": true, "XDG_DATA_HOME": true, "XDG_CACHE_HOME": true,
}

// isDangerousEnvVar checks if the environment variable name is potentially dangerous.
// This prevents users from injecting variables that could lead to RCE or configuration overrides
// in common tools (Git, Python, Node, etc.).
func isDangerousEnvVar(name string) bool {
	name = strings.ToUpper(name)

	if dangerousEnvVars[name] {
		return true
	}

	if strings.HasPrefix(name, "LD_") || strings.HasPrefix(name, "DYLD_") {
		return true
	}

	if strings.HasPrefix(name, "GIT_CONFIG_KEY_") || strings.HasPrefix(name, "GIT_CONFIG_VALUE_") {
		return true
	}

	return false
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

		if i == 0 && rooted {
			out = append(out, part)
			continue
		}

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

					if len(out) == 1 && out[0] == "" {
						continue
					}

					if len(out) == 2 && out[1] == "" {
						continue
					}

					out = out[:len(out)-1]
				} else {
					if last == ".." {
						out = append(out, part)
					} else {

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

	if rooted && res == "" {
		return "/"
	}

	if res == "" {
		return "."
	}

	return res
}

func checkForLocalFileAccess(val string) error {
	if filepath.IsAbs(val) {
		return fmt.Errorf("absolute path detected: %s (only relative paths are allowed for local execution)", val)
	}

	if strings.HasPrefix(strings.ToLower(val), "file:") {
		return fmt.Errorf("file: scheme detected: %s (local file access is not allowed)", val)
	}

	if err := validation.IsAllowedPath(val); err != nil {
		return fmt.Errorf("path access denied: %w", err)
	}
	return nil
}

func checkForArgumentInjection(val string) error {
	if strings.HasPrefix(val, "-") {

		if _, err := strconv.ParseFloat(val, 64); err == nil {
			return nil
		}
		return fmt.Errorf("argument injection detected: value starts with '-'")
	}
	if strings.HasPrefix(val, "+") {

		if _, err := strconv.ParseFloat(val, 64); err == nil {
			return nil
		}
		return fmt.Errorf("argument injection detected: value starts with '+'")
	}
	return nil
}

func isShellCommand(cmd string) bool {
	cmd = strings.TrimSpace(cmd)
	return isShell(cmd) || isInterpreter(cmd)
}

func isShell(cmd string) bool {
	cmd = strings.TrimSpace(cmd)

	shells := []string{
		"sh", "bash", "zsh", "dash", "ash", "ksh", "csh", "tcsh", "fish",
		"pwsh", "powershell", "powershell.exe", "pwsh.exe", "cmd", "cmd.exe",
		"ssh", "scp", "su", "sudo",
		"busybox", "expect", "watch", "tmux", "screen",
	}
	base := filepath.Base(cmd)
	for _, shell := range shells {
		if base == shell {
			return true
		}
	}

	ext := strings.ToLower(filepath.Ext(base))
	scriptExts := []string{
		".sh", ".bash", ".zsh", ".ash", ".ksh", ".csh", ".tcsh", ".fish",
		".bat", ".cmd", ".ps1", ".vbs",
	}
	for _, scriptExt := range scriptExts {
		if ext == scriptExt {
			return true
		}
	}
	return false
}

func checkForShellInjection(val string, template string, placeholder string, command string, isShell bool) error {

	quoteLevel := analyzeQuoteContext(template, placeholder)

	base := strings.ToLower(filepath.Base(command))
	isWindowsCmd := base == "cmd.exe" || base == "cmd"
	if isWindowsCmd && quoteLevel == 2 {
		quoteLevel = 0
	}

	if isInterpreter(command) {
		if err := checkInterpreterInjection(val, template, base, quoteLevel); err != nil {
			return err
		}

		if quoteLevel == 0 || quoteLevel == 1 || quoteLevel == 2 {
			if err := checkInterpreterFunctionCalls(val, base); err != nil {
				return err
			}
		}
	}

	if err := checkArgumentInterpreterInjection(val, template, base, quoteLevel, isShell); err != nil {
		return err
	}

	if quoteLevel == 3 {
		return checkBacktickInjection(val, command)
	}

	if quoteLevel == 2 {
		if strings.Contains(val, "'") {
			return fmt.Errorf("shell injection detected: value contains single quote which breaks out of single-quoted argument")
		}

		if !isShell && strings.Contains(val, "\\") {
			return fmt.Errorf("interpreter injection detected: value contains backslash inside single-quoted argument")
		}

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

	if quoteLevel == 1 {

		if idx := strings.IndexAny(val, "\"$`\\%"); idx != -1 {
			return fmt.Errorf("shell injection detected: value contains dangerous character %q inside double-quoted argument", val[idx])
		}
		return nil
	}

	return checkUnquotedInjection(val, command, isShell)
}

//nolint:gocyclo
func stripInterpreterComments(val, language string) string {
	var b strings.Builder
	b.Grow(len(val))

	inLineComment := false
	inBlockComment := false
	inSingle := false
	inDouble := false
	inBacktick := false
	escaped := false

	supportsHash := false
	supportsSlash := false
	supportsBlock := false

	switch language {
	case "python", "ruby", "perl", "sh", "bash", "zsh", "dash", "ash", "ksh", "csh", "tcsh", "fish":
		supportsHash = true
	case "node", "nodejs", "bun", "deno", "java", "c", "cpp", "go", "rust", "swift", "kotlin", "scala", "groovy":
		supportsSlash = true
		supportsBlock = true
	case "php":
		supportsHash = true
		supportsSlash = true
		supportsBlock = true
	default:

		supportsHash = true
		supportsSlash = true
		supportsBlock = true
	}

	for i := 0; i < len(val); i++ {
		char := val[i]

		if inLineComment {
			if char == '\n' {
				inLineComment = false
				b.WriteByte(char)
			}
			continue
		}
		if inBlockComment {
			if char == '*' && i+1 < len(val) && val[i+1] == '/' {
				inBlockComment = false
				i++
			}
			continue
		}

		if escaped {
			escaped = false
			b.WriteByte(char)
			continue
		}

		if char == '\'' && !inDouble && !inBacktick {
			inSingle = !inSingle
			b.WriteByte(char)
			continue
		}
		if char == '"' && !inSingle && !inBacktick {
			inDouble = !inDouble
			b.WriteByte(char)
			continue
		}
		if char == '`' && !inSingle && !inDouble {
			inBacktick = !inBacktick
			b.WriteByte(char)
			continue
		}

		if inSingle || inDouble || inBacktick {
			if char == '\\' {
				escaped = true

				b.WriteByte(char)
				continue
			}
			b.WriteByte(char)
			continue
		}

		if supportsHash && char == '#' {
			inLineComment = true
			continue
		}
		if (supportsSlash || supportsBlock) && char == '/' && i+1 < len(val) {
			if supportsSlash && val[i+1] == '/' {
				inLineComment = true
				i++
				continue
			}
			if supportsBlock && val[i+1] == '*' {
				inBlockComment = true
				i++
				continue
			}
		}

		if char == '\\' {

			continue
		}

		b.WriteByte(char)
	}
	return b.String()
}

func checkInterpreterFunctionCalls(val, language string) error {

	val = stripInterpreterComments(val, language)

	base := strings.ToLower(language)
	isStrict := strings.HasPrefix(base, "ruby") || strings.HasPrefix(base, "perl") ||
		strings.HasPrefix(base, "php") || strings.HasPrefix(base, "expect") ||
		strings.HasPrefix(base, "tclsh") || strings.HasPrefix(base, "wish") ||
		strings.HasPrefix(base, "lua") || strings.HasPrefix(base, "luajit")

	var statementKeywords []string
	var objectKeywords []string
	var functionKeywords []string

	universal := []string{
		"system", "exec", "popen", "eval", "spawn", "fork",
		"import", "require",
		"subprocess", "child_process", "os", "sys",
		"open", "read", "write",
		"phpinfo",
		"syscall", "dlopen", "fiddle", "send", "__send__", "public_send",
	}

	if isStrict {

		statementKeywords = universal
	} else {

		statementKeywords = []string{"import", "require"}

		objectKeywords = []string{
			"subprocess", "child_process", "os", "sys",
			"__builtins__", "__class__", "__base__", "__subclasses__",
		}

		functionKeywords = []string{
			"system", "exec", "popen", "eval", "spawn", "fork",
			"open", "read", "write",
			"getattr", "setattr", "delattr",
			"compile", "globals", "locals", "vars",
		}
	}

	if len(statementKeywords) > 0 {
		if err := checkUnquotedKeywords(val, statementKeywords); err != nil {
			return err
		}
	}

	if len(objectKeywords) > 0 {

		if err := checkContextualKeywords(val, objectKeywords, []rune{'.', '[', '(', '=', ':'}); err != nil {
			return err
		}
	}

	if len(functionKeywords) > 0 {

		if err := checkContextualKeywords(val, functionKeywords, []rune{'(', '=', ':', ' ', '\'', '"'}); err != nil {
			return err
		}
	}

	if strings.Contains(strings.ToLower(val), "__import__") {
		return fmt.Errorf("interpreter injection detected: value contains '__import__'")
	}

	return nil
}

//nolint:gocyclo

// Last word seen before current word (separated only by whitespace)

func checkKeyword(word []byte, keywords []string, lastChar rune, lastWord []byte) error {

	for _, kw := range keywords {

		if string(word) == kw {

			if lastChar == '$' || lastChar == '@' || lastChar == '%' || lastChar == '>' {
				return nil
			}

			if string(lastWord) == "sub" || string(lastWord) == "package" || string(lastWord) == "use" || string(lastWord) == "class" {
				return nil
			}
			return fmt.Errorf("interpreter injection detected: dangerous keyword %q found (unquoted)", kw)
		}
	}
	return nil
}

func checkInterpreterInjection(val, template, base string, quoteLevel int) error {
	if err := checkTarInjection(val, base); err != nil {
		return err
	}
	if err := checkFindInjection(val, base); err != nil {
		return err
	}
	if err := checkPythonInjection(val, template, base); err != nil {
		return err
	}
	if err := checkRubyInjection(val, base, quoteLevel); err != nil {
		return err
	}
	if err := checkNodePerlPhpInjection(val, base, quoteLevel); err != nil {
		return err
	}
	if err := checkAwkInjection(val, base); err != nil {
		return err
	}
	if err := checkJqInjection(val, base, quoteLevel); err != nil {
		return err
	}
	if err := checkSQLInjection(val, base, quoteLevel); err != nil {
		return err
	}
	if err := checkGdbInjection(val, base, quoteLevel); err != nil {
		return err
	}
	return nil
}

func checkGdbInjection(val, base string, quoteLevel int) error {
	isGdb := base == "gdb"
	if !isGdb {
		return nil
	}

	if quoteLevel > 1 {
		return nil
	}

	val = strings.TrimSpace(val)
	if val == "" {
		return nil
	}

	parts := strings.Fields(val)
	if len(parts) == 0 {
		return nil
	}
	firstWord := strings.ToLower(parts[0])

	dangerousKeywords := map[string]bool{
		"shell":  true,
		"system": true,
		"pipe":   true,
		"make":   true,
	}

	if dangerousKeywords[firstWord] {
		return fmt.Errorf("gdb injection detected: dangerous command %q at start", firstWord)
	}

	return nil
}

func checkJqInjection(val, base string, quoteLevel int) error {

	if base != "jq" {
		return nil
	}

	if quoteLevel == 1 {
		return nil
	}

	dangerousKeywords := []string{
		"env",
		"input", "inputs",
		"module", "import", "include",
		"halt", "halt_error",
		"stderr",
	}

	if err := checkUnquotedKeywords(val, dangerousKeywords); err != nil {
		return fmt.Errorf("jq injection detected: %w", err)
	}

	return nil
}

func checkTarInjection(val, base string) error {

	isTar := base == "tar" || base == "gtar" || base == "bsdtar"
	if isTar {
		valLower := strings.ToLower(val)

		if strings.Contains(valLower, "exec=") || strings.Contains(valLower, "command=") {
			return fmt.Errorf("tar injection detected: value contains execution directive")
		}

		if strings.Contains(valLower, "checkpoint-action") {
			return fmt.Errorf("tar injection detected: value contains 'checkpoint-action'")
		}
	}
	return nil
}

func checkFindInjection(val, base string) error {

	isFind := base == "find"
	if !isFind {
		return nil
	}

	parts := strings.Fields(val)
	for _, part := range parts {
		lowerPart := strings.ToLower(part)
		if lowerPart == "-exec" || lowerPart == "-execdir" || lowerPart == "-ok" || lowerPart == "-okdir" || lowerPart == "-delete" {
			return fmt.Errorf("find injection detected: value contains dangerous flag %q", part)
		}
	}
	return nil
}

func checkSQLiteInjection(valLower string) error {
	valTrimmed := strings.TrimSpace(valLower)
	valLower = strings.ToLower(valTrimmed)
	if strings.HasPrefix(valLower, ".") {
		if strings.HasPrefix(valLower, ".shell") || strings.HasPrefix(valLower, ".system") {
			return fmt.Errorf("sqlite3 injection detected: .shell/.system command")
		}
		if strings.HasPrefix(valLower, ".open") || strings.HasPrefix(valLower, ".output") || strings.HasPrefix(valLower, ".once") {
			return fmt.Errorf("sqlite3 injection detected: file manipulation command")
		}
		if strings.HasPrefix(valLower, ".read") || strings.HasPrefix(valLower, ".import") || strings.HasPrefix(valLower, ".load") {
			return fmt.Errorf("sqlite3 injection detected: dangerous meta-command")
		}
	}
	return nil
}

func checkMySQLInjection(val string) error {
	if err := checkUnquotedKeywords(val, []string{"system", "source"}); err != nil {
		return fmt.Errorf("mysql injection detected: %w", err)
	}
	valUpper := strings.ToUpper(val)
	if strings.Contains(valUpper, "INFILE") || strings.Contains(valUpper, "OUTFILE") {
		return fmt.Errorf("mysql injection detected: file access")
	}
	return nil
}

func checkPSQLInjection(val, valTrimmed string) error {
	if strings.HasPrefix(valTrimmed, "\\!") || strings.HasPrefix(valTrimmed, "\\o") || strings.HasPrefix(valTrimmed, "\\copy") {
		return fmt.Errorf("psql injection detected: dangerous meta-command")
	}
	valUpper := strings.ToUpper(val)
	if strings.Contains(valUpper, "COPY") && strings.Contains(valUpper, "PROGRAM") {
		return fmt.Errorf("psql injection detected: COPY TO PROGRAM")
	}
	return nil
}

func checkSQLKeywords(val string) error {
	upperVal := strings.ToUpper(val)
	keywords := []string{
		"OR", "AND", "UNION", "SELECT", "FROM", "WHERE", "JOIN",
		"DROP", "ALTER", "CREATE", "INSERT", "UPDATE", "DELETE",
		"--",
	}

	isBoundary := func(r byte) bool {
		return !isWordChar(r)
	}

	for _, kw := range keywords {
		if kw == "--" {
			if strings.Contains(upperVal, "--") {
				return fmt.Errorf("SQL injection detected: value contains '--'")
			}
			continue
		}

		idx := strings.Index(upperVal, kw)
		for idx != -1 {
			startOk := idx == 0 || isBoundary(upperVal[idx-1])
			endOk := idx+len(kw) == len(upperVal) || isBoundary(upperVal[idx+len(kw)])

			if startOk && endOk {
				return fmt.Errorf("SQL injection detected: value contains SQL keyword %q in unquoted context", kw)
			}
			nextIdx := strings.Index(upperVal[idx+1:], kw)
			if nextIdx == -1 {
				break
			}
			idx += 1 + nextIdx
		}
	}
	return nil
}

func checkPythonInjection(val, template, base string) error {

	if strings.HasPrefix(base, "python") {

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
	return nil
}

func checkRubyInjection(val, base string, quoteLevel int) error {
	if !strings.HasPrefix(base, "ruby") {
		return nil
	}

	if quoteLevel == 1 || quoteLevel == 3 {
		if strings.Contains(val, "#{") {
			return fmt.Errorf("ruby interpolation injection detected: value contains '#{'")
		}

		if strings.Contains(val, "#@") {
			return fmt.Errorf("ruby variable interpolation injection detected: value contains '#@'")
		}
	}

	if quoteLevel == 0 {

		if strings.Contains(val, "%x") {
			return fmt.Errorf("ruby execution injection detected: value contains '%%x'")
		}
		if strings.Contains(val, "@") {
			return fmt.Errorf("ruby variable injection detected: value contains '@'")
		}
	}

	if quoteLevel >= 1 && quoteLevel <= 3 {
		if strings.HasPrefix(strings.TrimSpace(val), "|") {
			return fmt.Errorf("ruby open injection detected: value starts with '|'")
		}
	}

	return nil
}

func checkNodePerlPhpInjection(val, base string, quoteLevel int) error {

	isNode := strings.HasPrefix(base, "node") || base == "bun" || base == "deno"
	isPerl := strings.HasPrefix(base, "perl")
	isPhp := strings.HasPrefix(base, "php")

	if isNode && quoteLevel == 3 {
		if strings.Contains(val, "${") {
			return fmt.Errorf("javascript template literal injection detected: value contains '${'")
		}
	}

	if (isPerl || isPhp) && (quoteLevel == 1 || quoteLevel == 3) {
		if strings.Contains(val, "${") {
			return fmt.Errorf("variable interpolation injection detected: value contains '${'")
		}
	}

	if isPerl {

		if quoteLevel == 0 || quoteLevel == 1 || quoteLevel == 3 {
			if strings.Contains(val, "@") {
				return fmt.Errorf("perl array interpolation injection detected: value contains '@'")
			}
			if strings.Contains(val, "%") {
				return fmt.Errorf("perl hash interpolation injection detected: value contains '%%'")
			}
		}

		if strings.Contains(val, "qx") {

			if quoteLevel == 0 || quoteLevel == 1 || quoteLevel == 3 {

				idx := strings.Index(val, "qx")
				for idx != -1 {

					isPrecededByWordChar := false
					if idx > 0 {
						if isWordChar(val[idx-1]) {
							isPrecededByWordChar = true
						}
					}

					if !isPrecededByWordChar {

						return fmt.Errorf("shell injection detected: perl qx execution")
					}

					next := strings.Index(val[idx+2:], "qx")
					if next == -1 {
						break
					}
					idx += 2 + next
				}
			}
		}

		if quoteLevel >= 1 && quoteLevel <= 3 {
			if strings.HasPrefix(strings.TrimSpace(val), "|") {
				return fmt.Errorf("perl open injection detected: value starts with '|'")
			}
		}
	}
	return nil
}

func checkAwkInjection(val, base string) error {

	isAwk := strings.HasPrefix(base, "awk") || strings.HasPrefix(base, "gawk") || strings.HasPrefix(base, "nawk") || strings.HasPrefix(base, "mawk")
	if !isAwk {
		return nil
	}

	inDouble := false
	inComment := false
	escaped := false

	for i := 0; i < len(val); i++ {
		char := val[i]

		if inComment {
			if char == '\n' {
				inComment = false
			}
			continue
		}

		if escaped {
			escaped = false
			continue
		}

		if char == '\\' {
			escaped = true
			continue
		}

		if char == '"' {
			inDouble = !inDouble
			continue
		}

		if inDouble {
			continue
		}

		if char == '#' {
			inComment = true
			continue
		}

		if char == '|' {
			return fmt.Errorf("awk injection detected: value contains '|' (pipe)")
		}
		if char == '>' {
			return fmt.Errorf("awk injection detected: value contains '>' (redirection)")
		}
		if char == '<' {
			return fmt.Errorf("awk injection detected: value contains '<' (redirection)")
		}
		if char == '@' {
			return fmt.Errorf("awk injection detected: value contains '@' (indirect call/extension)")
		}

		if char == 'g' {
			if strings.HasPrefix(val[i:], "getline") {

				end := i + 7
				if end >= len(val) || !isWordChar(val[end]) {

					if i == 0 || !isWordChar(val[i-1]) {
						return fmt.Errorf("awk injection detected: value contains 'getline'")
					}
				}
			}
		}

		if char == 's' {
			if strings.HasPrefix(val[i:], "system") {

				end := i + 6
				if end >= len(val) || !isWordChar(val[end]) {

					if i == 0 || !isWordChar(val[i-1]) {
						return fmt.Errorf("awk injection detected: value contains 'system'")
					}
				}
			}
		}

		if strings.Contains(val, "@") {
			return fmt.Errorf("awk injection detected: value contains '@'")
		}
	}
	return nil
}

func checkArgumentInterpreterInjection(val string, template string, base string, quoteLevel int, isShell bool) error {

	args := strings.Fields(template)
	for _, arg := range args {
		argBase := strings.ToLower(filepath.Base(arg))

		if argBase != base && isInterpreter(argBase) {
			effectiveQuoteLevel := quoteLevel
			if isShell {

				effectiveQuoteLevel = 0
			}

			if err := checkInterpreterInjection(val, template, argBase, effectiveQuoteLevel); err != nil {
				return fmt.Errorf("argument interpreter injection detected (%s): %w", argBase, err)
			}

			if effectiveQuoteLevel == 0 || effectiveQuoteLevel == 1 || effectiveQuoteLevel == 2 {
				if err := checkInterpreterFunctionCalls(val, argBase); err != nil {
					return fmt.Errorf("argument interpreter injection detected (%s): %w", argBase, err)
				}
			}
		}
	}
	return nil
}

func checkBacktickInjection(val, command string) error {

	if !isSafeBacktickLanguage(command) {
		const dangerousChars = ";|&$`(){}!<>\"\n\r\t\v\f*?[]~#%^'\\ "
		if idx := strings.IndexAny(val, dangerousChars); idx != -1 {
			return fmt.Errorf("shell injection detected: value contains dangerous character %q inside backticks", val[idx])
		}
	}

	if strings.Contains(val, "`") {
		return fmt.Errorf("backtick injection detected")
	}
	return nil
}

func isSafeBacktickLanguage(command string) bool {
	base := strings.ToLower(filepath.Base(command))

	safe := []string{"node", "nodejs", "bun", "deno"}
	for _, s := range safe {
		if base == s || strings.HasPrefix(base, s) {
			return true
		}
	}
	return false
}

func checkUnquotedInjection(val, command string, isShell bool) error {

	dangerousChars := ";|&$`(){}!<>\"\n\r\t\v\f*?[]~#%^'\\"
	if isShell {
		dangerousChars += " "
	}

	charsToCheck := dangerousChars

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
	interpreters := []string{

		"python", "ruby", "perl", "php",
		"node", "nodejs", "bun", "deno",
		"lua", "awk", "gawk", "nawk", "mawk", "sed",
		"jq",
		"psql", "mysql", "sqlite3",
		"docker",
		"env",
		"tclsh", "wish",
		"irb", "php-cgi",

		"vi", "vim", "nvim", "emacs", "nano",
		"less", "more", "man",

		"find", "xargs", "tee",
		"make", "rake", "ant", "mvn", "gradle",
		"npm", "yarn", "pnpm", "npx", "bunx", "go", "cargo", "pip",

		"kubectl", "helm", "aws", "gcloud", "az", "terraform", "ansible", "ansible-playbook",

		"r", "rscript", "julia", "groovy", "jshell",
		"scala", "kotlin", "swift",
		"elixir", "iex", "erl", "escript",
		"ghci", "clisp", "sbcl", "lisp", "scheme", "racket",
		"lua", "luajit",
		"gcc", "g++", "clang", "java",

		"zip", "unzip", "rsync", "nmap", "tcpdump", "gdb", "lldb",
		"tar", "gtar", "bsdtar",
	}
	for _, interp := range interpreters {
		if base == interp || strings.HasPrefix(base, interp) {
			return true
		}
	}

	ext := strings.ToLower(filepath.Ext(base))
	scriptExts := []string{
		".js", ".mjs", ".ts",
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

func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

func getPrefix(s string, idx int) string {

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

	minLevel := 3

	inSingle := false
	inDouble := false
	inBacktick := false
	escaped := false

	foundAny := false

	for i := 0; i < len(template); i++ {

		if strings.HasPrefix(template[i:], placeholder) {
			foundAny = true
			currentLevel := 0
			switch {
			case inSingle:
				currentLevel = 2
			case inBacktick:
				currentLevel = 3
			case inDouble:
				currentLevel = 1
			}

			if currentLevel < minLevel {
				minLevel = currentLevel
			}

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
			if !inSingle {
				inBacktick = !inBacktick
			}

			if inDouble && !inSingle {
				minLevel = 0
			}
		case '$':

			if inDouble && !inSingle && !inBacktick && i+1 < len(template) {
				next := template[i+1]
				if next == '(' || next == '{' {
					minLevel = 0
				}
			}
		}
	}

	if !foundAny {
		return 0
	}

	return minLevel
}

func checkEnvInjection(val string) error {
	// Relaxed check for environment variables.
	// Allows spaces, but blocks shell metacharacters.
	// We rely on validateSafePathAndInjection to prevent argument injection (flags starting with -).
	const dangerousChars = ";|&$`(){}!<>\"\n\r\t\v\f*?[]~#%^'\\" // Space removed
	if idx := strings.IndexAny(val, dangerousChars); idx != -1 {
		return fmt.Errorf("shell injection detected: value contains dangerous character %q", val[idx])
	}
	return nil
}

func validateSafePathAndInjection(val string, isDocker bool, commandName string) error {

	val = strings.TrimSpace(val)

	if strings.Contains(val, "://") {
		if err := validation.IsSafeURL(val); err != nil {
			return fmt.Errorf("unsafe url argument: %w", err)
		}
	} else {

		if strings.EqualFold(val, "localhost") {
			allowLoopback := os.Getenv("MCPANY_ALLOW_LOOPBACK_RESOURCES") == trueStr
			if !allowLoopback {
				return fmt.Errorf("unsafe argument: localhost is not allowed")
			}
		} else if validation.IsSafeIP != nil {

			if err := validation.IsSafeIP(val); err != nil && err.Error() != "invalid IP address" {
				return fmt.Errorf("unsafe IP argument: %w", err)
			}
		}
	}

	if err := checkForPathTraversal(val); err != nil {
		return err
	}

	if decodedVal, err := url.QueryUnescape(val); err == nil && decodedVal != val {
		if err := checkForPathTraversal(decodedVal); err != nil {
			return fmt.Errorf("%w (decoded)", err)
		}
	}

	if !isDocker {
		if err := checkForLocalFileAccess(val); err != nil {
			return err
		}

		if decodedVal, err := url.QueryUnescape(val); err == nil && decodedVal != val {
			if err := checkForLocalFileAccess(decodedVal); err != nil {
				return fmt.Errorf("%w (decoded)", err)
			}
		}
	}

	if err := checkForArgumentInjection(val); err != nil {
		return err
	}

	if decodedVal, err := url.QueryUnescape(val); err == nil && decodedVal != val {
		if err := checkForArgumentInjection(decodedVal); err != nil {
			return fmt.Errorf("%w (decoded)", err)
		}
	}

	if isVulnerableToSchemes(commandName) {
		if err := checkForDangerousSchemes(val); err != nil {
			return err
		}

		if decodedVal, err := url.QueryUnescape(val); err == nil && decodedVal != val {
			if err := checkForDangerousSchemes(decodedVal); err != nil {
				return fmt.Errorf("%w (decoded)", err)
			}
		}
	}

	return nil
}

func isVulnerableToSchemes(command string) bool {
	base := strings.ToLower(filepath.Base(command))

	magickTools := []string{
		"convert", "mogrify", "identify", "composite", "compare", "stream",
		"montage", "display", "animate", "import", "conjure", "magick",
	}
	for _, tool := range magickTools {
		if base == tool {
			return true
		}
	}

	ffmpegTools := []string{"ffmpeg", "ffprobe", "ffplay"}
	for _, tool := range ffmpegTools {
		if base == tool {
			return true
		}
	}

	if base == gitCommand {
		return true
	}

	return false
}

func checkForDangerousSchemes(val string) error {

	idx := strings.Index(val, ":")
	if idx == -1 {
		return nil
	}

	scheme := strings.ToLower(val[:idx])

	for _, r := range scheme {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '+' || r == '-' || r == '.') {
			return nil
		}
	}

	dangerous := map[string]bool{

		"file": true, "gopher": true, "expect": true, "php": true,
		"zip": true, "jar": true, "war": true,

		"mvg": true, "msl": true, "vid": true, "ephemeral": true,
		"label": true, "text": true, "info": true, "pango": true,
		"caption": true, "plasma": true, "xc": true, "inline": true,
		"gradient": true, "pattern": true, "tile": true, "read": true,

		"concat": true, "subfile": true, "crypto": true, "data": true,
		"hls": true, "http": false, "https": false,
		"ftp": true, "rtmp": true, "rtsp": true,

		"ext": true,
	}

	if dangerous[scheme] {
		return fmt.Errorf("dangerous scheme detected: %s", scheme)
	}

	return nil
}

func checkWordSuffix(word []byte, keywords []string, runes []rune, nextIdx int, isSuffix func(rune) bool) error {
	found := false
	for _, kw := range keywords {
		if string(word) == kw {
			found = true
			break
		}
	}
	if !found {
		return nil
	}
	for k := nextIdx; k < len(runes); k++ {
		r := runes[k]
		if unicode.IsSpace(r) {
			continue
		}
		if isSuffix(r) {
			return fmt.Errorf("interpreter injection detected: dangerous keyword %q followed by %q", word, r)
		}
		return nil
	}
	return nil
}
func checkContextualKeywords(val string, keywords []string, suffixes []rune) error {
	var state quoteState

	wordBuf := make([]byte, 0, 64)
	inWord := false
	runes := []rune(val)

	isSuffix := func(r rune) bool {
		for _, s := range suffixes {
			if r == s {
				return true
			}
		}
		return false
	}

	for i := 0; i < len(runes); i++ {
		char := runes[i]

		if state.escaped {
			state.escaped = false
			continue
		}
		if char == '\\' {
			state.escaped = true
			continue
		}

		if state.handleQuotes(char) {
			if state.inQuote() && inWord {
				if err := checkWordSuffix(wordBuf, keywords, runes, i, isSuffix); err != nil {
					return err
				}
				inWord = false
			}
			continue
		}

		if state.inQuote() {
			continue
		}

		if char < 128 && isWordChar(byte(char)) {
			if !inWord {
				inWord = true
				wordBuf = wordBuf[:0]
			}
			wordBuf = append(wordBuf, byte(char))
		} else if inWord {

			if err := checkWordSuffix(wordBuf, keywords, runes, i, isSuffix); err != nil {
				return err
			}
			inWord = false
		}
	}

	if inWord {
		if err := checkWordSuffix(wordBuf, keywords, runes, len(runes), isSuffix); err != nil {
			return err
		}
	}
	return nil
}

func checkUnquotedKeywords(val string, keywords []string) error {
	inSingle := false
	inDouble := false
	inBacktick := false
	escaped := false

	wordBuf := make([]byte, 0, 64)
	lastChar := rune(0)
	var lastWord []byte

	for _, char := range val {
		if escaped {
			escaped = false
			continue
		}
		if char == '\\' {
			escaped = true
			continue
		}

		if char == '\'' && !inDouble && !inBacktick {
			inSingle = !inSingle

			if inSingle {
				if len(wordBuf) > 0 {
					if err := checkKeyword(wordBuf, keywords, lastChar, lastWord); err != nil {
						return err
					}

					lastWord = append(lastWord[:0], wordBuf...)
					wordBuf = wordBuf[:0]
				}
			}

			if !inSingle {
				lastChar = char
				lastWord = lastWord[:0]
			}
			continue
		}
		if char == '"' && !inSingle && !inBacktick {
			inDouble = !inDouble
			if inDouble {
				if len(wordBuf) > 0 {
					if err := checkKeyword(wordBuf, keywords, lastChar, lastWord); err != nil {
						return err
					}
					lastWord = append(lastWord[:0], wordBuf...)
					wordBuf = wordBuf[:0]
				}
			}
			if !inDouble {
				lastChar = char
				lastWord = lastWord[:0]
			}
			continue
		}
		if char == '`' && !inSingle && !inDouble {
			inBacktick = !inBacktick
			if inBacktick {
				if len(wordBuf) > 0 {
					if err := checkKeyword(wordBuf, keywords, lastChar, lastWord); err != nil {
						return err
					}
					lastWord = append(lastWord[:0], wordBuf...)
					wordBuf = wordBuf[:0]
				}
			}
			if !inBacktick {
				lastChar = char
				lastWord = lastWord[:0]
			}
			continue
		}

		if inSingle || inDouble || inBacktick {
			continue
		}

		if char < 128 && isWordChar(byte(char)) {
			wordBuf = append(wordBuf, byte(char))
		} else {

			if len(wordBuf) > 0 {
				if err := checkKeyword(wordBuf, keywords, lastChar, lastWord); err != nil {
					return err
				}
				lastWord = append(lastWord[:0], wordBuf...)
				wordBuf = wordBuf[:0]
			}

			if !unicode.IsSpace(char) {
				lastChar = char
				lastWord = lastWord[:0]
			}
		}
	}

	if len(wordBuf) > 0 {
		if err := checkKeyword(wordBuf, keywords, lastChar, lastWord); err != nil {
			return err
		}
	}
	return nil
}

func checkSQLInjection(val, base string, quoteLevel int) error {
	isSQL := base == "psql" || base == "mysql" || base == "sqlite3"
	if !isSQL {
		return nil
	}
	if quoteLevel != 0 {
		return nil
	}
	valTrimmed := strings.TrimSpace(val)
	valLower := strings.ToLower(valTrimmed)

	if base == "sqlite3" {
		if err := checkSQLiteInjection(valLower); err != nil {
			return err
		}
	}
	if base == "mysql" {
		if err := checkMySQLInjection(val); err != nil {
			return err
		}
	}
	if base == "psql" {
		if err := checkPSQLInjection(val, valTrimmed); err != nil {
			return err
		}
	}

	return checkSQLKeywords(val)
}

type quoteState struct {
	inSingle   bool
	inDouble   bool
	inBacktick bool
	escaped    bool
}

func (s *quoteState) handleQuotes(char rune) bool {
	if char == '\'' && !s.inDouble && !s.inBacktick {
		s.inSingle = !s.inSingle
		return true
	}
	if char == '"' && !s.inSingle && !s.inBacktick {
		s.inDouble = !s.inDouble
		return true
	}
	if char == '`' && !s.inSingle && !s.inDouble {
		s.inBacktick = !s.inBacktick
		return true
	}
	return false
}

func (s *quoteState) inQuote() bool {
	return s.inSingle || s.inDouble || s.inBacktick
}
