/*
 * Copyright 2025 Author(s) of MCP-XY
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
	// ContentTypeApplicationJSON is the content type for JSON.
	ContentTypeApplicationJSON = "application/json"
	// ContentTypeTextEventStream is the content type for text event streams.
	ContentTypeTextEventStream = "text/event-stream"
	// HeaderMcpSessionID is the header for the MCP session ID.
	HeaderMcpSessionID = "Mcp-Session-Id"
	// MethodConfigureCache is the method for configuring the cache.
	MethodConfigureCache = "configure_cache"
	// OauthCallbackPath is the path for the OAuth2 callback.
	OauthCallbackPath = "/v1/oauth2/callback"
	// DefaultOriginAllow is the default value for the Access-Control-Allow-Origin header.
	DefaultOriginAllow = "*"
	// ToolNameServiceSeparator is the separator used to construct a fully qualified tool name from a service ID and a tool name.
	ToolNameServiceSeparator = "."
	// MethodToolsCall is the method for calling a tool.
	MethodToolsCall = "tools/call"
	// MethodToolsList is the method for listing tools.
	MethodToolsList = "tools/list"
	// MethodPromptsList is the method for listing prompts.
	MethodPromptsList = "prompts/list"
	// MethodPromptsGet is the method for getting a prompt.
	MethodPromptsGet = "prompts/get"
	// MethodResourcesList is the method for listing resources.
	MethodResourcesList = "resources/list"
	// MethodResourcesRead is the method for reading a resource.
	MethodResourcesRead = "resources/read"
	// MethodResourcesSubscribe is the method for subscribing to a resource.
	MethodResourcesSubscribe = "resources/subscribe"
	// NotificationPromptsListChanged is the notification for when the prompts list has changed.
	NotificationPromptsListChanged = "notifications/prompts/list_changed"
	// NotificationResourcesListChanged is the notification for when the resources list has changed.
	NotificationResourcesListChanged = "notifications/resources/list_changed"
)

const (
	// CommandStatusSuccess is the status for a successful command execution.
	CommandStatusSuccess = "SUCCESS"
	// CommandStatusError is the status for a failed command execution.
	CommandStatusError = "ERROR"
	// CommandStatusTimeout is the status for a command that timed out.
	CommandStatusTimeout = "TIMEOUT"
)
