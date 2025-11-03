/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
)

const (
	// CommandStatusSuccess represents the status for a successful command execution.
	CommandStatusSuccess = "SUCCESS"
	// CommandStatusError represents the status for a failed command execution.
	CommandStatusError = "ERROR"
	// CommandStatusTimeout represents the status for a command that timed out.
	CommandStatusTimeout = "TIMEOUT"
)

const (
	// DefaultBindPort is the default port for the bind address.
	DefaultBindPort = 8070
)
