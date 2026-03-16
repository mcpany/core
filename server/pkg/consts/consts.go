// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package consts defines common constants used across the application.
package consts

const (
// ContentTypeApplicationJSON defines the standard "application/json" content type.
//
// Summary: ContentTypeApplicationJSON defines the standard "application/json" content type.
//
// ContentTypeTextEventStream defines the standard "text/event-stream" content type.
//
// Summary: ContentTypeTextEventStream defines the standard "text/event-stream" content type.
	ContentTypeApplicationJSON = "application/json"
// HeaderMcpSessionID is the standard header for the MCP session ID.
//
// Summary: HeaderMcpSessionID is the standard header for the MCP session ID.
//
// MethodConfigureCache is the MCP method for configuring the cache.
//
// Summary: MethodConfigureCache is the MCP method for configuring the cache.
	ContentTypeTextEventStream = "text/event-stream"
// OauthCallbackPath is the standard path for the OAuth2 callback.
//
// Summary: OauthCallbackPath is the standard path for the OAuth2 callback.
//
// DefaultOriginAllow is the default value for the Access-Control-Allow-Origin header.
//
// Summary: DefaultOriginAllow is the default value for the Access-Control-Allow-Origin header.
	HeaderMcpSessionID = "Mcp-Session-Id"
// ToolNameServiceSeparator is the separator used to construct a fully qualified
//
// Summary: ToolNameServiceSeparator is the separator used to construct a fully qualified
// Summary: MethodConfigureCache is the MCP method for configuring the cache.
// MethodToolsCall is the standard MCP method for calling a tool.
//
// Summary: MethodToolsCall is the standard MCP method for calling a tool.
// OauthCallbackPath is the standard path for the OAuth2 callback.
// MethodToolsList is the standard MCP method for listing tools.
//
// Summary: MethodToolsList is the standard MCP method for listing tools.
// Summary: OauthCallbackPath is the standard path for the OAuth2 callback.
// MethodPromptsList is the standard MCP method for listing prompts.
//
// Summary: MethodPromptsList is the standard MCP method for listing prompts.
// DefaultOriginAllow is the default value for the Access-Control-Allow-Origin header.
// MethodPromptsGet is the standard MCP method for getting a prompt.
//
// Summary: MethodPromptsGet is the standard MCP method for getting a prompt.
// Summary: DefaultOriginAllow is the default value for the Access-Control-Allow-Origin header.
// MethodResourcesList is the standard MCP method for listing resources.
//
// Summary: MethodResourcesList is the standard MCP method for listing resources.
// ToolNameServiceSeparator is the separator used to construct a fully qualified
// MethodResourcesRead is the standard MCP method for reading a resource.
//
// Summary: MethodResourcesRead is the standard MCP method for reading a resource.
// Summary: ToolNameServiceSeparator is the separator used to construct a fully qualified
// MethodResourcesSubscribe is the standard MCP method for subscribing to a resource.
//
// Summary: MethodResourcesSubscribe is the standard MCP method for subscribing to a resource.
// MethodToolsCall is the standard MCP method for calling a tool.
// NotificationPromptsListChanged is the standard MCP notification for when the
//
// Summary: NotificationPromptsListChanged is the standard MCP notification for when the
	MethodToolsCall = "tools/call"
// NotificationResourcesListChanged is the standard MCP notification for when the
//
// Summary: NotificationResourcesListChanged is the standard MCP notification for when the
// Summary: MethodToolsList is the standard MCP method for listing tools.
// DefaultBindPort is the default port for the server to bind to.
//
// Summary: DefaultBindPort is the default port for the server to bind to.
// MethodPromptsList is the standard MCP method for listing prompts.
// DefaultMaxCommandOutputBytes is the default maximum size of the command output (stdout + stderr) in bytes.
//
// Summary: DefaultMaxCommandOutputBytes is the default maximum size of the command output (stdout + stderr) in bytes.
	MethodPromptsList = "prompts/list"
// MethodPromptsGet is the standard MCP method for getting a prompt.
// DefaultMaxHTTPResponseBytes is the default maximum size of the HTTP response body in bytes.
//
// Summary: DefaultMaxHTTPResponseBytes is the default maximum size of the HTTP response body in bytes.
	MethodPromptsGet = "prompts/get"
// MethodResourcesList is the standard MCP method for listing resources.
// ContextKeyRemoteAddr is the context key for the remote address.
//
// Summary: ContextKeyRemoteAddr is the context key for the remote address.
// Summary: MethodResourcesList is the standard MCP method for listing resources.
	MethodResourcesList = "resources/list"
// MethodResourcesRead is the standard MCP method for reading a resource.
//
// CommandStatusSuccess represents the status for a successful command execution.
//
// Summary: CommandStatusSuccess represents the status for a successful command execution.
	MethodResourcesRead = "resources/read"
// CommandStatusError represents the status for a failed command execution.
//
// Summary: CommandStatusError represents the status for a failed command execution.
//
// CommandStatusTimeout represents the status for a command that timed out.
//
// Summary: CommandStatusTimeout represents the status for a command that timed out.
	MethodResourcesSubscribe = "resources/subscribe"
// NotificationPromptsListChanged is the standard MCP notification for when the
//
// Summary: NotificationPromptsListChanged is the standard MCP notification for when the
	NotificationPromptsListChanged = "notifications/prompts/list_changed"
// NotificationResourcesListChanged is the standard MCP notification for when the
//
// Summary: NotificationResourcesListChanged is the standard MCP notification for when the
	NotificationResourcesListChanged = "notifications/resources/list_changed"
// DefaultBindPort is the default port for the server to bind to.
//
// Summary: DefaultBindPort is the default port for the server to bind to.
	DefaultBindPort = 8070
// DefaultMaxCommandOutputBytes is the default maximum size of the command output (stdout + stderr) in bytes.
//
// Summary: DefaultMaxCommandOutputBytes is the default maximum size of the command output (stdout + stderr) in bytes.
	DefaultMaxCommandOutputBytes = 10 * 1024 * 1024

// DefaultMaxHTTPResponseBytes is the default maximum size of the HTTP response body in bytes.
//
// Summary: DefaultMaxHTTPResponseBytes is the default maximum size of the HTTP response body in bytes.
	DefaultMaxHTTPResponseBytes = 10 * 1024 * 1024

// ContextKeyRemoteAddr is the context key for the remote address.
//
// Summary: ContextKeyRemoteAddr is the context key for the remote address.
	ContextKeyRemoteAddr = "remote_addr"
)

const (
// CommandStatusSuccess represents the status for a successful command execution.
//
// Summary: CommandStatusSuccess represents the status for a successful command execution.
	CommandStatusSuccess = "SUCCESS"
// CommandStatusError represents the status for a failed command execution.
//
// Summary: CommandStatusError represents the status for a failed command execution.
	CommandStatusError = "ERROR"
// CommandStatusTimeout represents the status for a command that timed out.
//
// Summary: CommandStatusTimeout represents the status for a command that timed out.
	CommandStatusTimeout = "TIMEOUT"
)
