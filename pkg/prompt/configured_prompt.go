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
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/mcpany/core/pkg/util/template"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ConfiguredPrompt is an implementation of the Prompt interface that is backed
// by a prompt definition from the static configuration.
type ConfiguredPrompt struct {
	definition *configv1.PromptDefinition
	serviceID  string
}

// NewConfiguredPrompt creates a new ConfiguredPrompt from a prompt definition.
func NewConfiguredPrompt(definition *configv1.PromptDefinition, serviceID string) (*ConfiguredPrompt, error) {
	if definition == nil {
		return nil, fmt.Errorf("prompt definition is nil")
	}
	if definition.GetName() == "" {
		return nil, fmt.Errorf("prompt name is required")
	}
	return &ConfiguredPrompt{definition: definition, serviceID: serviceID}, nil
}

// Prompt returns the MCP representation of the prompt.
func (p *ConfiguredPrompt) Prompt() *mcp.Prompt {
	args := p.definition.GetArguments()
	mcpArgs := make([]*mcp.PromptArgument, len(args))
	for i, arg := range args {
		mcpArgs[i] = &mcp.PromptArgument{
			Name:        arg.GetName(),
			Description: arg.GetDescription(),
			Required:    arg.GetRequired(),
		}
	}

	return &mcp.Prompt{
		Name:        p.definition.GetName(),
		Title:       p.definition.GetTitle(),
		Description: p.definition.GetDescription(),
		Arguments:   mcpArgs,
	}
}

// Service returns the ID of the service that provides this prompt.
func (p *ConfiguredPrompt) Service() string {
	return p.serviceID
}

// Get returns the prompt message, with any templates rendered using the
// provided arguments.
func (p *ConfiguredPrompt) Get(ctx context.Context, args json.RawMessage) (*mcp.GetPromptResult, error) {
	var arguments map[string]any
	if err := json.Unmarshal(args, &arguments); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	result := &mcp.GetPromptResult{
		Messages: make([]*mcp.PromptMessage, 0, len(p.definition.GetMessages())),
	}

	for _, msg := range p.definition.GetMessages() {
		var content mcp.Content
		var role mcp.Role

		switch r := msg.GetRole(); r {
		case configv1.PromptMessage_USER:
			role = mcp.Role("user")
		case configv1.PromptMessage_ASSISTANT:
			role = mcp.Role("assistant")
		default:
			return nil, fmt.Errorf("unsupported role: %v", r)
		}

		if text := msg.GetText(); text != nil {
			renderedText, err := template.Render(text.GetText(), arguments)
			if err != nil {
				return nil, fmt.Errorf("failed to render text template: %w", err)
			}
			content = &mcp.TextContent{Text: renderedText}
		} else if image := msg.GetImage(); image != nil {
			data, err := base64.StdEncoding.DecodeString(image.GetData())
			if err != nil {
				return nil, fmt.Errorf("failed to decode image data: %w", err)
			}
			content = &mcp.ImageContent{Data: data, MIMEType: image.GetMimeType()}
		} else if audio := msg.GetAudio(); audio != nil {
			data, err := base64.StdEncoding.DecodeString(audio.GetData())
			if err != nil {
				return nil, fmt.Errorf("failed to decode audio data: %w", err)
			}
			content = &mcp.AudioContent{Data: data, MIMEType: audio.GetMimeType()}
		} else if resource := msg.GetResource(); resource != nil {
			res := resource.GetResource()
			if res == nil {
				return nil, fmt.Errorf("resource content is nil")
			}
			if static := res.GetStatic(); static != nil {
				content = &mcp.TextContent{Text: static.GetTextContent()}
			}
		} else {
			return nil, fmt.Errorf("unsupported message content type")
		}

		result.Messages = append(result.Messages, &mcp.PromptMessage{
			Role:    role,
			Content: content,
		})
	}
	return result, nil
}
