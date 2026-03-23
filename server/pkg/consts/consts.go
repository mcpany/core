// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package consts defines common constants used across the application.
package consts

const (
	// ContentTypeApplicationJSON defines the standard "application/json" content type.
	//
	// Summary: Defines the standard HTTP Content-Type for JSON payloads.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - string: "application/json"
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	ContentTypeApplicationJSON = "application/json"

	// ContentTypeTextEventStream defines the standard "text/event-stream" content type.
	//
	// Summary: Defines the standard HTTP Content-Type for Server-Sent Events (SSE).
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - string: "text/event-stream"
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	ContentTypeTextEventStream = "text/event-stream"

	// HeaderMcpSessionID is the standard header for the MCP session ID.
	//
	// Summary: Defines the custom HTTP header used to track active MCP sessions.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - string: "Mcp-Session-Id"
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	HeaderMcpSessionID = "Mcp-Session-Id"

	// MethodConfigureCache is the MCP method for configuring the cache.
	//
	// Summary: Defines the JSON-RPC method used to configure caching behavior.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - string: "configure_cache"
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	MethodConfigureCache = "configure_cache"

	// OauthCallbackPath is the standard path for the OAuth2 callback.
	//
	// Summary: Defines the URL path where OAuth2 providers should redirect after authentication.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - string: "/v1/oauth2/callback"
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	OauthCallbackPath = "/v1/oauth2/callback"

	// DefaultOriginAllow is the default value for the Access-Control-Allow-Origin header.
	//
	// Summary: Defines the default wildcard CORS allowed origin.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - string: "*"
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	DefaultOriginAllow = "*"

	// ToolNameServiceSeparator is the separator used to construct a fully qualified
	// tool name from a service ID and a tool name.
	//
	// Summary: Defines the substring used to namespace tool names.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - string: "."
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	ToolNameServiceSeparator = "."

	// MethodToolsCall is the standard MCP method for calling a tool.
	//
	// Summary: Defines the JSON-RPC method to execute an MCP tool.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - string: "tools/call"
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	MethodToolsCall = "tools/call"

	// MethodToolsList is the standard MCP method for listing tools.
	//
	// Summary: Defines the JSON-RPC method to enumerate available tools.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - string: "tools/list"
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	MethodToolsList = "tools/list"

	// MethodPromptsList is the standard MCP method for listing prompts.
	//
	// Summary: Defines the JSON-RPC method to enumerate available prompts.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - string: "prompts/list"
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	MethodPromptsList = "prompts/list"

	// MethodPromptsGet is the standard MCP method for getting a prompt.
	//
	// Summary: Defines the JSON-RPC method to fetch a specific prompt's contents.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - string: "prompts/get"
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	MethodPromptsGet = "prompts/get"

	// MethodResourcesList is the standard MCP method for listing resources.
	//
	// Summary: Defines the JSON-RPC method to enumerate available resources.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - string: "resources/list"
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	MethodResourcesList = "resources/list"

	// MethodResourcesRead is the standard MCP method for reading a resource.
	//
	// Summary: Defines the JSON-RPC method to read a specific resource's contents.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - string: "resources/read"
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	MethodResourcesRead = "resources/read"

	// MethodResourcesSubscribe is the standard MCP method for subscribing to a resource.
	//
	// Summary: Defines the JSON-RPC method to subscribe to resource updates.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - string: "resources/subscribe"
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	MethodResourcesSubscribe = "resources/subscribe"

	// NotificationPromptsListChanged is the standard MCP notification for when the
	// prompts list has changed.
	//
	// Summary: Defines the JSON-RPC notification sent when the prompts list changes.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - string: "notifications/prompts/list_changed"
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	NotificationPromptsListChanged = "notifications/prompts/list_changed"

	// NotificationResourcesListChanged is the standard MCP notification for when the
	// resources list has changed.
	//
	// Summary: Defines the JSON-RPC notification sent when the resources list changes.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - string: "notifications/resources/list_changed"
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	NotificationResourcesListChanged = "notifications/resources/list_changed"

	// DefaultBindPort is the default port for the server to bind to.
	//
	// Summary: Defines the default port number the server listens on if not configured.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - int: 8070
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	DefaultBindPort = 8070

	// DefaultMaxCommandOutputBytes is the default maximum size of the command output (stdout + stderr) in bytes.
	//
	// Summary: Specifies the default limit for captured standard output from shell commands.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - int: 10MB in bytes
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	DefaultMaxCommandOutputBytes = 10 * 1024 * 1024

	// DefaultMaxHTTPResponseBytes is the default maximum size of the HTTP response body in bytes.
	//
	// Summary: Specifies the default memory limit for fetching downstream HTTP bodies.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - int: 10MB in bytes
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	DefaultMaxHTTPResponseBytes = 10 * 1024 * 1024

	// ContextKeyRemoteAddr is the context key for the remote address.
	//
	// Summary: Defines the key used to store the client's remote network address in request contexts.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - string: "remote_addr"
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	ContextKeyRemoteAddr = "remote_addr"
)

const (
	// CommandStatusSuccess represents the status for a successful command execution.
	//
	// Summary: Indicates an execution completed properly with exit code 0.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - string: "SUCCESS"
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	CommandStatusSuccess = "SUCCESS"

	// CommandStatusError represents the status for a failed command execution.
	//
	// Summary: Indicates an execution failed, crashed, or yielded a non-zero exit code.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - string: "ERROR"
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	CommandStatusError = "ERROR"

	// CommandStatusTimeout represents the status for a command that timed out.
	//
	// Summary: Indicates an execution exceeded its allowed time-to-live.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - string: "TIMEOUT"
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	CommandStatusTimeout = "TIMEOUT"
)
