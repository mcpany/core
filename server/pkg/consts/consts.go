// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package consts defines common constants used across the application.
package consts

const (
// Summary: ContentTypeApplicationJSON defines the standard "application/json" content type. Defines ContentTypeApplicationJSON.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
	ContentTypeApplicationJSON = "application/json"
// Summary: ContentTypeTextEventStream defines the standard "text/event-stream" content type. Defines ContentTypeTextEventStream.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
	ContentTypeTextEventStream = "text/event-stream"
// Summary: HeaderMcpSessionID is the standard header for the MCP session ID. Defines HeaderMcpSessionID.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
	HeaderMcpSessionID = "Mcp-Session-Id"
// Summary: MethodConfigureCache is the MCP method for configuring the cache. Defines MethodConfigureCache.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
	MethodConfigureCache = "configure_cache"
// Summary: OauthCallbackPath is the standard path for the OAuth2 callback. Defines OauthCallbackPath.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
	OauthCallbackPath = "/v1/oauth2/callback"
// Summary: DefaultOriginAllow is the default value for the Access-Control-Allow-Origin header. Defines DefaultOriginAllow.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
	DefaultOriginAllow = "*"
// Summary: ToolNameServiceSeparator is the separator used to construct a fully qualified tool name from a service ID and a tool name. Defines ToolNameServiceSeparator.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
	ToolNameServiceSeparator = "."
// Summary: MethodToolsCall is the standard MCP method for calling a tool. Defines MethodToolsCall.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
	MethodToolsCall = "tools/call"
// Summary: MethodToolsList is the standard MCP method for listing tools. Defines MethodToolsList.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
	MethodToolsList = "tools/list"
// Summary: MethodPromptsList is the standard MCP method for listing prompts. Defines MethodPromptsList.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
	MethodPromptsList = "prompts/list"
// Summary: MethodPromptsGet is the standard MCP method for getting a prompt. Defines MethodPromptsGet.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
	MethodPromptsGet = "prompts/get"
// Summary: MethodResourcesList is the standard MCP method for listing resources. Defines MethodResourcesList.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
	MethodResourcesList = "resources/list"
// Summary: MethodResourcesRead is the standard MCP method for reading a resource. Defines MethodResourcesRead.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
	MethodResourcesRead = "resources/read"
// Summary: MethodResourcesSubscribe is the standard MCP method for subscribing to a resource. Defines MethodResourcesSubscribe.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
	MethodResourcesSubscribe = "resources/subscribe"
// Summary: NotificationPromptsListChanged is the standard MCP notification for when the prompts list has changed. Defines NotificationPromptsListChanged.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
	NotificationPromptsListChanged = "notifications/prompts/list_changed"
// Summary: NotificationResourcesListChanged is the standard MCP notification for when the resources list has changed. Defines NotificationResourcesListChanged.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
	NotificationResourcesListChanged = "notifications/resources/list_changed"
// Summary: DefaultBindPort is the default port for the server to bind to. Defines DefaultBindPort.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
	DefaultBindPort = 8070
// Summary: DefaultMaxCommandOutputBytes is the default maximum size of the command output (stdout + stderr) in bytes. 10MB should be enough for most use cases while preventing OOM. Defines DefaultMaxCommandOutputBytes.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
	DefaultMaxCommandOutputBytes = 10 * 1024 * 1024

// Summary: DefaultMaxHTTPResponseBytes is the default maximum size of the HTTP response body in bytes. 10MB should be enough for most use cases while preventing OOM. Defines DefaultMaxHTTPResponseBytes.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
	DefaultMaxHTTPResponseBytes = 10 * 1024 * 1024

// Summary: ContextKeyRemoteAddr is the context key for the remote address. Defines ContextKeyRemoteAddr.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
	ContextKeyRemoteAddr = "remote_addr"
)

const (
// Summary: CommandStatusSuccess represents the status for a successful command execution. Defines CommandStatusSuccess.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
	CommandStatusSuccess = "SUCCESS"
// Summary: CommandStatusError represents the status for a failed command execution. Defines CommandStatusError.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
	CommandStatusError = "ERROR"
// Summary: CommandStatusTimeout represents the status for a command that timed out. Defines CommandStatusTimeout.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
	CommandStatusTimeout = "TIMEOUT"
)
