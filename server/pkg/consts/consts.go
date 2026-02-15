// Package consts defines common constants used across the application.

package consts

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

const (
	// CommandStatusSuccess represents the status for a successful command execution.
	CommandStatusSuccess = "SUCCESS"
	// CommandStatusError represents the status for a failed command execution.
	CommandStatusError = "ERROR"
	// CommandStatusTimeout represents the status for a command that timed out.
	CommandStatusTimeout = "TIMEOUT"
)
