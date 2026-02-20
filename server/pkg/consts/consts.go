// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package consts defines common constants used across the application.
package consts

const (
	// ContentTypeApplicationJSON defines the standard "application/json" content type.
	//
	// Summary: Represents the MIME type for JSON content.
	ContentTypeApplicationJSON = "application/json"

	// ContentTypeTextEventStream defines the standard "text/event-stream" content type.
	//
	// Summary: Represents the MIME type for Server-Sent Events (SSE).
	ContentTypeTextEventStream = "text/event-stream"

	// HeaderMcpSessionID is the standard header for the MCP session ID.
	//
	// Summary: Identifies the MCP session ID in HTTP headers.
	HeaderMcpSessionID = "Mcp-Session-Id"

	// MethodConfigureCache is the MCP method for configuring the cache.
	//
	// Summary: Identifies the MCP method for cache configuration.
	MethodConfigureCache = "configure_cache"

	// OauthCallbackPath is the standard path for the OAuth2 callback.
	//
	// Summary: Defines the URL path for OAuth2 callback handling.
	OauthCallbackPath = "/v1/oauth2/callback"

	// DefaultOriginAllow is the default value for the Access-Control-Allow-Origin header.
	//
	// Summary: Defines the default CORS origin allowance.
	DefaultOriginAllow = "*"

	// ToolNameServiceSeparator is the separator used to construct a fully qualified
	// tool name from a service ID and a tool name.
	//
	// Summary: Separates service ID and tool name in fully qualified tool identifiers.
	ToolNameServiceSeparator = "."

	// MethodToolsCall is the standard MCP method for calling a tool.
	//
	// Summary: Identifies the MCP method for tool execution.
	MethodToolsCall = "tools/call"

	// MethodToolsList is the standard MCP method for listing tools.
	//
	// Summary: Identifies the MCP method for tool listing.
	MethodToolsList = "tools/list"

	// MethodPromptsList is the standard MCP method for listing prompts.
	//
	// Summary: Identifies the MCP method for prompt listing.
	MethodPromptsList = "prompts/list"

	// MethodPromptsGet is the standard MCP method for getting a prompt.
	//
	// Summary: Identifies the MCP method for prompt retrieval.
	MethodPromptsGet = "prompts/get"

	// MethodResourcesList is the standard MCP method for listing resources.
	//
	// Summary: Identifies the MCP method for resource listing.
	MethodResourcesList = "resources/list"

	// MethodResourcesRead is the standard MCP method for reading a resource.
	//
	// Summary: Identifies the MCP method for resource reading.
	MethodResourcesRead = "resources/read"

	// MethodResourcesSubscribe is the standard MCP method for subscribing to a resource.
	//
	// Summary: Identifies the MCP method for resource subscription.
	MethodResourcesSubscribe = "resources/subscribe"

	// NotificationPromptsListChanged is the standard MCP notification for when the
	// prompts list has changed.
	//
	// Summary: Identifies the notification for prompt list changes.
	NotificationPromptsListChanged = "notifications/prompts/list_changed"

	// NotificationResourcesListChanged is the standard MCP notification for when the
	// resources list has changed.
	//
	// Summary: Identifies the notification for resource list changes.
	NotificationResourcesListChanged = "notifications/resources/list_changed"

	// DefaultBindPort is the default port for the server to bind to.
	//
	// Summary: Defines the default port number for the server.
	DefaultBindPort = 8070

	// DefaultMaxCommandOutputBytes is the default maximum size of the command output (stdout + stderr) in bytes.
	// 10MB should be enough for most use cases while preventing OOM.
	//
	// Summary: Limits the size of command execution output to prevent memory exhaustion.
	DefaultMaxCommandOutputBytes = 10 * 1024 * 1024

	// DefaultMaxHTTPResponseBytes is the default maximum size of the HTTP response body in bytes.
	// 10MB should be enough for most use cases while preventing OOM.
	//
	// Summary: Limits the size of HTTP response bodies to prevent memory exhaustion.
	DefaultMaxHTTPResponseBytes = 10 * 1024 * 1024

	// ContextKeyRemoteAddr is the context key for the remote address.
	//
	// Summary: Identifies the context key for storing the remote client address.
	ContextKeyRemoteAddr = "remote_addr"
)

const (
	// CommandStatusSuccess represents the status for a successful command execution.
	//
	// Summary: Indicates successful command execution.
	CommandStatusSuccess = "SUCCESS"

	// CommandStatusError represents the status for a failed command execution.
	//
	// Summary: Indicates failed command execution.
	CommandStatusError = "ERROR"

	// CommandStatusTimeout represents the status for a command that timed out.
	//
	// Summary: Indicates command execution timeout.
	CommandStatusTimeout = "TIMEOUT"
)
