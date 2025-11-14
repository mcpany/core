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

import (
	"context"
	"encoding/json"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ConfigPrompt is a prompt that is defined by a protobuf configuration.
type ConfigPrompt struct {
	prompt  *mcp.Prompt
	service string
}

// NewConfigPrompt creates a new ConfigPrompt from a protobuf definition.
func NewConfigPrompt(def *configv1.PromptDefinition, service string) *ConfigPrompt {
	name := def.GetUri()
	if strings.HasPrefix(name, "mcp://prompts/") {
		name = strings.TrimPrefix(name, "mcp://prompts/")
	}
	return &ConfigPrompt{
		prompt: &mcp.Prompt{
			Name:        name,
			Description: def.GetText(),
		},
		service: service,
	}
}

// Prompt returns the MCP representation of the prompt.
func (p *ConfigPrompt) Prompt() *mcp.Prompt {
	return p.prompt
}

// Service returns the ID of the service that provides this prompt.
func (p *ConfigPrompt) Service() string {
	return p.service
}

// Get executes the prompt with the given arguments and returns the result.
func (p *ConfigPrompt) Get(ctx context.Context, args json.RawMessage) (*mcp.GetPromptResult, error) {
	// For now, we just return the prompt text as a single user message.
	return &mcp.GetPromptResult{
		Messages: []*mcp.PromptMessage{
			{
				Role:    RoleUser,
				Content: &mcp.TextContent{Text: p.prompt.Description},
			},
		},
	}, nil
}
