// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package consts defines common constants used across the application.
// Summary: ContentTypeApplicationJSON defines the standard "application/json" content type.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: ContentTypeTextEventStream defines the standard "text/event-stream" content type.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: HeaderMcpSessionID is the standard header for the MCP session ID.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: MethodConfigureCache is the MCP method for configuring the cache.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: OauthCallbackPath is the standard path for the OAuth2 callback.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: DefaultOriginAllow is the default value for the Access-Control-Allow-Origin header.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: ToolNameServiceSeparator is the separator used to construct a fully qualified
// tool name from a service ID and a tool name.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: MethodToolsCall is the standard MCP method for calling a tool.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: MethodToolsList is the standard MCP method for listing tools.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: MethodPromptsList is the standard MCP method for listing prompts.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: MethodPromptsGet is the standard MCP method for getting a prompt.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: MethodResourcesList is the standard MCP method for listing resources.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: MethodResourcesRead is the standard MCP method for reading a resource.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: MethodResourcesSubscribe is the standard MCP method for subscribing to a resource.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: NotificationPromptsListChanged is the standard MCP notification for when the
// prompts list has changed.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: NotificationResourcesListChanged is the standard MCP notification for when the
// resources list has changed.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: DefaultBindPort is the default port for the server to bind to.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: DefaultMaxCommandOutputBytes is the default maximum size of the command output (stdout + stderr) in bytes.
// 10MB should be enough for most use cases while preventing OOM.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: DefaultMaxHTTPResponseBytes is the default maximum size of the HTTP response body in bytes.
// 10MB should be enough for most use cases while preventing OOM.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: ContextKeyRemoteAddr is the context key for the remote address.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: CommandStatusSuccess represents the status for a successful command execution.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: CommandStatusError represents the status for a failed command execution.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: CommandStatusTimeout represents the status for a command that timed out.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
package consts

const (
	ContentTypeApplicationJSON	= "application/json"

	ContentTypeTextEventStream	= "text/event-stream"

	HeaderMcpSessionID	= "Mcp-Session-Id"

	MethodConfigureCache	= "configure_cache"

	OauthCallbackPath	= "/v1/oauth2/callback"

	DefaultOriginAllow	= "*"

	ToolNameServiceSeparator	= "."

	MethodToolsCall	= "tools/call"

	MethodToolsList	= "tools/list"

	MethodPromptsList	= "prompts/list"

	MethodPromptsGet	= "prompts/get"

	MethodResourcesList	= "resources/list"

	MethodResourcesRead	= "resources/read"

	MethodResourcesSubscribe	= "resources/subscribe"

	NotificationPromptsListChanged	= "notifications/prompts/list_changed"

	NotificationResourcesListChanged	= "notifications/resources/list_changed"

	DefaultBindPort	= 8070

	DefaultMaxCommandOutputBytes	= 10 * 1024 * 1024

	DefaultMaxHTTPResponseBytes	= 10 * 1024 * 1024

	ContextKeyRemoteAddr	= "remote_addr"
)

const (
	CommandStatusSuccess	= "SUCCESS"

	CommandStatusError	= "ERROR"

	CommandStatusTimeout	= "TIMEOUT"
)
