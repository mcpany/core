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
	"fmt"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ConfigPrompt is an implementation of the Prompt interface that is backed by a
// protobuf configuration.
type ConfigPrompt struct {
	prompt    *configv1.Prompt
	serviceID string
}

// NewConfigPrompt creates a new ConfigPrompt from a protobuf configuration.
func NewConfigPrompt(prompt *configv1.Prompt, serviceID string) (Prompt, error) {
	return &ConfigPrompt{
		prompt:    prompt,
		serviceID: serviceID,
	}, nil
}

// Prompt returns the MCP representation of the prompt.
func (p *ConfigPrompt) Prompt() *mcp.Prompt {
	args := make([]*mcp.PromptArgument, len(p.prompt.GetArguments()))
	for i, arg := range p.prompt.GetArguments() {
		args[i] = &mcp.PromptArgument{
			Name:        arg.GetName(),
			Description: arg.GetDescription(),
			Required:    arg.GetRequired(),
		}
	}

	return &mcp.Prompt{
		Name:        p.prompt.GetName(),
		Title:       p.prompt.GetTitle(),
		Description: p.prompt.GetDescription(),
		Arguments:   args,
	}
}

// Service returns the ID of the service that provides this prompt.
func (p *ConfigPrompt) Service() string {
	return p.serviceID
}

// Get executes the prompt with the given arguments and returns the result.
func (p *ConfigPrompt) Get(_ context.Context, argsJSON json.RawMessage) (*mcp.GetPromptResult, error) {
	var args map[string]string
	if err := json.Unmarshal(argsJSON, &args); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	messages := make([]*mcp.PromptMessage, len(p.prompt.GetMessages()))
	for i, msg := range p.prompt.GetMessages() {
		var content mcp.Content
		switch msg.WhichContent() {
		case configv1.PromptMessage_TextContent_case:
			text := msg.GetTextContent().GetText()
			for k, v := range args {
				text = strings.ReplaceAll(text, fmt.Sprintf("{{%s}}", k), v)
			}
			content = &mcp.TextContent{Text: text}
		default:
			return nil, fmt.Errorf("unsupported content type: %T", msg.WhichContent())
		}

		role, err := stringToRole(msg.GetRole())
		if err != nil {
			return nil, err
		}

		messages[i] = &mcp.PromptMessage{
			Role:    role,
			Content: content,
		}
	}

	return &mcp.GetPromptResult{
		Messages: messages,
	}, nil
}

func stringToRole(s string) (mcp.Role, error) {
	switch strings.ToLower(s) {
	case "user":
		return "user", nil
	case "assistant":
		return "assistant", nil
	default:
		return "", fmt.Errorf("invalid role: %s", s)
	}
}
