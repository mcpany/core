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

package prompt

import "github.com/modelcontextprotocol/go-sdk/mcp"

const (
	// RoleUser represents the "user" role in a prompt message.
	RoleUser = "user"
	// RoleAssistant represents the "assistant" role in a prompt message.
	RoleAssistant = "assistant"
)

const (
	// ContentTypeText represents a text content type in a prompt message.
	ContentTypeText = "text"
	// ContentTypeImage represents an image content type in a prompt message.
	ContentTypeImage = "image"
	// ContentTypeAudio represents an audio content type in a prompt message.
	ContentTypeAudio = "audio"
	// ContentTypeResource represents a resource content type in a prompt message.
	ContentTypeResource = "resource"
)

// Argument defines an argument for a prompt, including its name,
// description, and whether it is required.
type Argument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// Message represents a single message within a prompt, including its role and
// content.
type Message struct {
	Role    string      `json:"role"`
	Content mcp.Content `json:"content"`
}
