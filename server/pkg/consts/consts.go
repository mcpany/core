// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package consts defines common constants used across the application.
package consts

// ContentTypeTextEventStream defines a constant or variable for content type text event stream.
// HeaderMcpSessionID defines a constant or variable for header mcp session i d.
// MethodConfigureCache defines a constant or variable for method configure cache.
// ContextKeyRemoteAddr defines a constant or variable for context key remote addr.
// DefaultMaxHTTPResponseBytes defines a constant or variable for default max h t t p response bytes.
// DefaultMaxCommandOutputBytes defines a constant or variable for default max command output bytes.
// DefaultBindPort defines a constant or variable for default bind port.
// NotificationResourcesListChanged defines a constant or variable for notification resources list changed.
// NotificationPromptsListChanged defines a constant or variable for notification prompts list changed.
// MethodResourcesSubscribe defines a constant or variable for method resources subscribe.
// OauthCallbackPath defines a constant or variable for oauth callback path.
// MethodResourcesList defines a constant or variable for method resources list.
// ContentTypeApplicationJSON defines a constant or variable for content type application j s o n.
// MethodPromptsList defines a constant or variable for method prompts list.
// MethodToolsList defines a constant or variable for method tools list.
// MethodToolsCall defines a constant or variable for method tools call.
// ToolNameServiceSeparator defines a constant or variable for tool name service separator.
// DefaultOriginAllow defines a constant or variable for default origin allow.
// MethodResourcesRead defines a constant or variable for method resources read.
// MethodPromptsGet defines a constant or variable for method prompts get.
const (
	// ContentTypeApplicationJSON defines the standard "application/json" content type.
	ContentTypeApplicationJSON = "application/json"
	// ContentTypeTextEventStream defines the standard "text/event-stream" content type.
	ContentTypeTextEventStream = "text/event-stream"
	// HeaderMcpSessionID is the standard header for the MCP session ID.
	HeaderMcpSessionID = "Mcp-Session-Id"
	// MethodConfigureCache is the MCP method for configuring the cache.
	MethodConfigureCache = "configure_cache"
	// OauthCallbackPath is the standard path for the OAuth2 callback.
	OauthCallbackPath = "/v1/oauth2/callback"
	// DefaultOriginAllow is the default value for the Access-Control-Allow-Origin header.
	DefaultOriginAllow = "*"
	// ToolNameServiceSeparator is the separator used to construct a fully qualified
	// tool name from a service ID and a tool name.
	ToolNameServiceSeparator = "."
	// MethodToolsCall is the standard MCP method for calling a tool.
	MethodToolsCall = "tools/call"
	// MethodToolsList is the standard MCP method for listing tools.
	MethodToolsList = "tools/list"
	// MethodPromptsList is the standard MCP method for listing prompts.
	MethodPromptsList = "prompts/list"
	// MethodPromptsGet is the standard MCP method for getting a prompt.
	MethodPromptsGet = "prompts/get"
	// MethodResourcesList is the standard MCP method for listing resources.
	MethodResourcesList = "resources/list"
	// MethodResourcesRead is the standard MCP method for reading a resource.
	MethodResourcesRead = "resources/read"
	// MethodResourcesSubscribe is the standard MCP method for subscribing to a resource.
	MethodResourcesSubscribe = "resources/subscribe"
	// NotificationPromptsListChanged is the standard MCP notification for when the
	// prompts list has changed.
	NotificationPromptsListChanged = "notifications/prompts/list_changed"
	// NotificationResourcesListChanged is the standard MCP notification for when the
	// resources list has changed.
	NotificationResourcesListChanged = "notifications/resources/list_changed"
	// DefaultBindPort is the default port for the server to bind to.
	DefaultBindPort = 8070
	// DefaultMaxCommandOutputBytes is the default maximum size of the command output (stdout + stderr) in bytes.
	// 10MB should be enough for most use cases while preventing OOM.
	DefaultMaxCommandOutputBytes = 10 * 1024 * 1024

	// DefaultMaxHTTPResponseBytes is the default maximum size of the HTTP response body in bytes.
	// 10MB should be enough for most use cases while preventing OOM.
	DefaultMaxHTTPResponseBytes = 10 * 1024 * 1024

	// ContextKeyRemoteAddr is the context key for the remote address.
	ContextKeyRemoteAddr = "remote_addr"
)

// CommandStatusError defines a constant or variable for command status error.
// CommandStatusTimeout defines a constant or variable for command status timeout.
// CommandStatusSuccess defines a constant or variable for command status success.
const (
	// CommandStatusSuccess represents the status for a successful command execution.
	CommandStatusSuccess = "SUCCESS"
	// CommandStatusError represents the status for a failed command execution.
	CommandStatusError = "ERROR"
	// CommandStatusTimeout represents the status for a command that timed out.
	CommandStatusTimeout = "TIMEOUT"
)
