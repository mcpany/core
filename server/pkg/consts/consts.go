// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package consts defines common constants used across the application.
package consts

const (
	// ContentTypeApplicationJSON defines the standard "application/json" content type.
	// Summary: Standard JSON content type string.
	ContentTypeApplicationJSON = "application/json"
	// ContentTypeTextEventStream defines the standard "text/event-stream" content type.
	// Summary: Standard SSE content type string.
	ContentTypeTextEventStream = "text/event-stream"
	// HeaderMcpSessionID is the standard header for the MCP session ID.
	// Summary: HTTP header key for MCP session identification.
	HeaderMcpSessionID = "Mcp-Session-Id"
	// MethodConfigureCache is the MCP method for configuring the cache.
	// Summary: MCP method name for cache configuration.
	MethodConfigureCache = "configure_cache"
	// OauthCallbackPath is the standard path for the OAuth2 callback.
	// Summary: URL path for OAuth2 authentication callbacks.
	OauthCallbackPath = "/v1/oauth2/callback"
	// DefaultOriginAllow is the default value for the Access-Control-Allow-Origin header.
	// Summary: Default wildcard CORS origin.
	DefaultOriginAllow = "*"
	// ToolNameServiceSeparator is the separator used to construct a fully qualified
	// tool name from a service ID and a tool name.
	// Summary: Separator used in namespaced tool identifiers.
	ToolNameServiceSeparator = "."
	// MethodToolsCall is the standard MCP method for calling a tool.
	// Summary: MCP method name for tool execution.
	MethodToolsCall = "tools/call"
	// MethodToolsList is the standard MCP method for listing tools.
	// Summary: MCP method name for tool discovery.
	MethodToolsList = "tools/list"
	// MethodPromptsList is the standard MCP method for listing prompts.
	// Summary: MCP method name for prompt discovery.
	MethodPromptsList = "prompts/list"
	// MethodPromptsGet is the standard MCP method for getting a prompt.
	// Summary: MCP method name for prompt retrieval.
	MethodPromptsGet = "prompts/get"
	// MethodResourcesList is the standard MCP method for listing resources.
	// Summary: MCP method name for resource discovery.
	MethodResourcesList = "resources/list"
	// MethodResourcesRead is the standard MCP method for reading a resource.
	// Summary: MCP method name for resource reading.
	MethodResourcesRead = "resources/read"
	// MethodResourcesSubscribe is the standard MCP method for subscribing to a resource.
	// Summary: MCP method name for resource subscription.
	MethodResourcesSubscribe = "resources/subscribe"
	// NotificationPromptsListChanged is the standard MCP notification for when the
	// prompts list has changed.
	// Summary: MCP notification name for prompt list updates.
	NotificationPromptsListChanged = "notifications/prompts/list_changed"
	// NotificationResourcesListChanged is the standard MCP notification for when the
	// resources list has changed.
	// Summary: MCP notification name for resource list updates.
	NotificationResourcesListChanged = "notifications/resources/list_changed"
	// DefaultBindPort is the default port for the server to bind to.
	// Summary: Default TCP port for the MCP server.
	DefaultBindPort = 8070
	// DefaultMaxCommandOutputBytes is the default maximum size of the command output (stdout + stderr) in bytes.
	// 10MB should be enough for most use cases while preventing OOM.
	// Summary: Default byte limit for CLI tool output.
	DefaultMaxCommandOutputBytes = 10 * 1024 * 1024

	// DefaultMaxHTTPResponseBytes is the default maximum size of the HTTP response body in bytes.
	// 10MB should be enough for most use cases while preventing OOM.
	// Summary: Default byte limit for upstream HTTP responses.
	DefaultMaxHTTPResponseBytes = 10 * 1024 * 1024

	// ContextKeyRemoteAddr is the context key for the remote address.
	// Summary: Context key for the client's remote address.
	ContextKeyRemoteAddr = "remote_addr"
)

const (
	// CommandStatusSuccess represents the status for a successful command execution.
	// Summary: Execution status for successful commands.
	CommandStatusSuccess = "SUCCESS"
	// CommandStatusError represents the status for a failed command execution.
	// Summary: Execution status for failed commands.
	CommandStatusError = "ERROR"
	// CommandStatusTimeout represents the status for a command that timed out.
	// Summary: Execution status for commands that exceeded their timeout.
	CommandStatusTimeout = "TIMEOUT"
)
