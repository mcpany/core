// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package consts defines common constants used across the application.
package consts

const (
	// ContentTypeApplicationJSON defines the standard "application/json" content type.
	// Summary: Defines ContentTypeApplicationJSON.
//
// Summary: Represents ->:ContentTypeApplicationJSON
//
//
// Summary: Represents ->:ContentTypeTextEventStream
//
//
// Summary: Represents ->:HeaderMcpSessionID
//
//
// Summary: Represents ->:MethodConfigureCache
//
//
// Summary: Represents ->:OauthCallbackPath
//
//
// Summary: Represents ->:DefaultOriginAllow
//
	ContentTypeApplicationJSON = "application/json"
//
// Summary: Represents ->:ToolNameServiceSeparator
//
//
// Summary: Represents ->:MethodToolsCall
//
//
// Summary: Represents ->:MethodToolsList
//
//
// Summary: Represents ->:MethodPromptsList
//
//
// Summary: Represents ->:MethodPromptsGet
//
//
// Summary: Represents ->:MethodResourcesList
//
//
// Summary: Represents ->:MethodResourcesRead
//
//
// Summary: Represents ->:MethodResourcesSubscribe
//
	// ContentTypeTextEventStream defines the standard "text/event-stream" content type.
//
// Summary: Represents ->:NotificationPromptsListChanged
//
	// Summary: Defines ContentTypeTextEventStream.
//
// Summary: Represents ->:NotificationResourcesListChanged
//
//
// Summary: Represents ->:DefaultBindPort
//
	ContentTypeTextEventStream = "text/event-stream"
//
// Summary: Represents ->:DefaultMaxCommandOutputBytes
//
	// HeaderMcpSessionID is the standard header for the MCP session ID.
	// Summary: Defines HeaderMcpSessionID.
//
// Summary: Represents ->:DefaultMaxHTTPResponseBytes
//
	HeaderMcpSessionID = "Mcp-Session-Id"
//
// Summary: Represents ->:ContextKeyRemoteAddr
//
	// MethodConfigureCache is the MCP method for configuring the cache.
	// Summary: Defines MethodConfigureCache.
//
// Summary: Represents ->:MethodConfigureCache
//
	MethodConfigureCache = "configure_cache"
//
// Summary: Represents ->:CommandStatusSuccess
//
//
// Summary: Represents ->:CommandStatusError
//
//
// Summary: Represents ->:CommandStatusTimeout
//
	// OauthCallbackPath is the standard path for the OAuth2 callback.
	// Summary: Defines OauthCallbackPath.
//
// Summary: Represents ->:OauthCallbackPath
//
	OauthCallbackPath = "/v1/oauth2/callback"
	// DefaultOriginAllow is the default value for the Access-Control-Allow-Origin header.
	// Summary: Defines DefaultOriginAllow.
//
// Summary: Represents ->:DefaultOriginAllow
//
	DefaultOriginAllow = "*"
	// ToolNameServiceSeparator is the separator used to construct a fully qualified
	// tool name from a service ID and a tool name.
	// Summary: Defines ToolNameServiceSeparator.
//
// Summary: Represents ->:ToolNameServiceSeparator
//
	ToolNameServiceSeparator = "."
	// MethodToolsCall is the standard MCP method for calling a tool.
	// Summary: Defines MethodToolsCall.
//
// Summary: Represents ->:MethodToolsCall
//
	MethodToolsCall = "tools/call"
	// MethodToolsList is the standard MCP method for listing tools.
	// Summary: Defines MethodToolsList.
//
// Summary: Represents ->:MethodToolsList
//
	MethodToolsList = "tools/list"
	// MethodPromptsList is the standard MCP method for listing prompts.
	// Summary: Defines MethodPromptsList.
//
// Summary: Represents ->:MethodPromptsList
//
	MethodPromptsList = "prompts/list"
	// MethodPromptsGet is the standard MCP method for getting a prompt.
	// Summary: Defines MethodPromptsGet.
//
// Summary: Represents ->:MethodPromptsGet
//
	MethodPromptsGet = "prompts/get"
	// MethodResourcesList is the standard MCP method for listing resources.
	// Summary: Defines MethodResourcesList.
//
// Summary: Represents ->:MethodResourcesList
//
	MethodResourcesList = "resources/list"
	// MethodResourcesRead is the standard MCP method for reading a resource.
	// Summary: Defines MethodResourcesRead.
//
// Summary: Represents ->:MethodResourcesRead
//
	MethodResourcesRead = "resources/read"
	// MethodResourcesSubscribe is the standard MCP method for subscribing to a resource.
	// Summary: Defines MethodResourcesSubscribe.
//
// Summary: Represents ->:MethodResourcesSubscribe
//
	MethodResourcesSubscribe = "resources/subscribe"
	// NotificationPromptsListChanged is the standard MCP notification for when the
	// prompts list has changed.
	// Summary: Defines NotificationPromptsListChanged.
//
// Summary: Represents ->:NotificationPromptsListChanged
//
	NotificationPromptsListChanged = "notifications/prompts/list_changed"
	// NotificationResourcesListChanged is the standard MCP notification for when the
	// resources list has changed.
	// Summary: Defines NotificationResourcesListChanged.
//
// Summary: Represents ->:NotificationResourcesListChanged
//
	NotificationResourcesListChanged = "notifications/resources/list_changed"
	// DefaultBindPort is the default port for the server to bind to.
	// Summary: Defines DefaultBindPort.
//
// Summary: Represents ->:DefaultBindPort
//
	DefaultBindPort = 8070
	// DefaultMaxCommandOutputBytes is the default maximum size of the command output (stdout + stderr) in bytes.
	// 10MB should be enough for most use cases while preventing OOM.
	// Summary: Defines DefaultMaxCommandOutputBytes.
//
// Summary: Represents ->:DefaultMaxCommandOutputBytes
//
	DefaultMaxCommandOutputBytes = 10 * 1024 * 1024

	// DefaultMaxHTTPResponseBytes is the default maximum size of the HTTP response body in bytes.
	// 10MB should be enough for most use cases while preventing OOM.
	// Summary: Defines DefaultMaxHTTPResponseBytes.
//
// Summary: Represents ->:DefaultMaxHTTPResponseBytes
//
	DefaultMaxHTTPResponseBytes = 10 * 1024 * 1024

	// ContextKeyRemoteAddr is the context key for the remote address.
	// Summary: Defines ContextKeyRemoteAddr.
//
// Summary: Represents ->:ContextKeyRemoteAddr
//
	ContextKeyRemoteAddr = "remote_addr"
)

const (
	// CommandStatusSuccess represents the status for a successful command execution.
	// Summary: Defines CommandStatusSuccess.
//
// Summary: Represents ->:CommandStatusSuccess
//
	CommandStatusSuccess = "SUCCESS"
	// CommandStatusError represents the status for a failed command execution.
	// Summary: Defines CommandStatusError.
//
// Summary: Represents ->:CommandStatusError
//
	CommandStatusError = "ERROR"
	// CommandStatusTimeout represents the status for a command that timed out.
	// Summary: Defines CommandStatusTimeout.
//
// Summary: Represents ->:CommandStatusTimeout
//
	CommandStatusTimeout = "TIMEOUT"
)
