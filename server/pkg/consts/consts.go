// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package consts defines common constants used across the application.
package consts

const (
	// ContentTypeApplicationJSON defines the standard "application/json" content type.
	// Summary: Defines ContentTypeApplicationJSON.
	ContentTypeApplicationJSON = "application/json"
	// ContentTypeTextEventStream defines the standard "text/event-stream" content type.
	// Summary: Defines ContentTypeTextEventStream.
	ContentTypeTextEventStream = "text/event-stream"
	// HeaderMcpSessionID is the standard header for the MCP session ID.
	// Summary: Defines HeaderMcpSessionID.
	HeaderMcpSessionID = "Mcp-Session-Id"
	// MethodConfigureCache is the MCP method for configuring the cache.
	// Summary: Defines MethodConfigureCache.
	MethodConfigureCache = "configure_cache"
	// OauthCallbackPath is the standard path for the OAuth2 callback.
	// Summary: Defines OauthCallbackPath.
	OauthCallbackPath = "/v1/oauth2/callback"
	// DefaultOriginAllow is the default value for the Access-Control-Allow-Origin header.
	// Summary: Defines DefaultOriginAllow.
	DefaultOriginAllow = "*"
	// ToolNameServiceSeparator is the separator used to construct a fully qualified
	// tool name from a service ID and a tool name.
	// Summary: Defines ToolNameServiceSeparator.
	ToolNameServiceSeparator = "."
	// MethodToolsCall is the standard MCP method for calling a tool.
	// Summary: Defines MethodToolsCall.
	MethodToolsCall = "tools/call"
	// MethodToolsList is the standard MCP method for listing tools.
	// Summary: Defines MethodToolsList.
	MethodToolsList = "tools/list"
	// MethodPromptsList is the standard MCP method for listing prompts.
	// Summary: Defines MethodPromptsList.
	MethodPromptsList = "prompts/list"
	// MethodPromptsGet is the standard MCP method for getting a prompt.
	// Summary: Defines MethodPromptsGet.
	MethodPromptsGet = "prompts/get"
	// MethodResourcesList is the standard MCP method for listing resources.
	// Summary: Defines MethodResourcesList.
	MethodResourcesList = "resources/list"
	// MethodResourcesRead is the standard MCP method for reading a resource.
	// Summary: Defines MethodResourcesRead.
	MethodResourcesRead = "resources/read"
	// MethodResourcesSubscribe is the standard MCP method for subscribing to a resource.
	// Summary: Defines MethodResourcesSubscribe.
	MethodResourcesSubscribe = "resources/subscribe"
	// NotificationPromptsListChanged is the standard MCP notification for when the
	// prompts list has changed.
	// Summary: Defines NotificationPromptsListChanged.
	NotificationPromptsListChanged = "notifications/prompts/list_changed"
	// NotificationResourcesListChanged is the standard MCP notification for when the
	// resources list has changed.
	// Summary: Defines NotificationResourcesListChanged.
	NotificationResourcesListChanged = "notifications/resources/list_changed"
	// DefaultBindPort is the default port for the server to bind to.
	// Summary: Defines DefaultBindPort.
	DefaultBindPort = 8070
	// DefaultMaxCommandOutputBytes is the default maximum size of the command output (stdout + stderr) in bytes.
	// 10MB should be enough for most use cases while preventing OOM.
	// Summary: Defines DefaultMaxCommandOutputBytes.
	DefaultMaxCommandOutputBytes = 10 * 1024 * 1024

	// DefaultMaxHTTPResponseBytes is the default maximum size of the HTTP response body in bytes.
	// 10MB should be enough for most use cases while preventing OOM.
	// Summary: Defines DefaultMaxHTTPResponseBytes.
	DefaultMaxHTTPResponseBytes = 10 * 1024 * 1024

	// ContextKeyRemoteAddr is the context key for the remote address.
	// Summary: Defines ContextKeyRemoteAddr.
	ContextKeyRemoteAddr = "remote_addr"
)

const (
	// CommandStatusSuccess represents the status for a successful command execution.
	// Summary: Defines CommandStatusSuccess.
	CommandStatusSuccess = "SUCCESS"
	// CommandStatusError represents the status for a failed command execution.
	// Summary: Defines CommandStatusError.
	CommandStatusError = "ERROR"
	// CommandStatusTimeout represents the status for a command that timed out.
	// Summary: Defines CommandStatusTimeout.
	CommandStatusTimeout = "TIMEOUT"
)
