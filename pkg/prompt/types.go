// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/mcpany/core/pkg/transformer"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	ErrPromptNotFound = errors.New("prompt not found")
)

// Prompt is the fundamental interface for any executable prompt in the system.
type Prompt interface {
	Prompt() *mcp.Prompt
	Service() string
	Get(ctx context.Context, args json.RawMessage) (*mcp.GetPromptResult, error)
}

// MCPServerProvider defines an interface for components that can provide an
// instance of an *mcp.Server. This is used to decouple the PromptManager from the
// concrete server implementation.
type MCPServerProvider interface {
	Server() *mcp.Server
}

// TemplatedPrompt implements the Prompt interface for a prompt that is defined
// by a template.
type TemplatedPrompt struct {
	definition *configv1.PromptDefinition
	serviceID  string
}

// NewTemplatedPrompt creates a new TemplatedPrompt.
func NewTemplatedPrompt(definition *configv1.PromptDefinition, serviceID string) *TemplatedPrompt {
	return &TemplatedPrompt{
		definition: definition,
		serviceID:  serviceID,
	}
}

// Prompt returns the MCP prompt definition.
func (p *TemplatedPrompt) Prompt() *mcp.Prompt {
	args := make([]*mcp.PromptArgument, len(p.definition.GetArguments()))
	for i, arg := range p.definition.GetArguments() {
		args[i] = &mcp.PromptArgument{
			Name:        arg.GetName(),
			Description: arg.GetDescription(),
			Required:    arg.GetRequired(),
		}
	}
	sanitizedName, _ := util.SanitizeToolName(p.definition.GetName())

	return &mcp.Prompt{
		Name:        p.serviceID + "." + sanitizedName,
		Title:       p.definition.GetTitle(),
		Description: p.definition.GetDescription(),
		Arguments:   args,
	}
}

// Service returns the ID of the service that provides this prompt.
func (p *TemplatedPrompt) Service() string {
	return p.serviceID
}

// Get executes the prompt with the provided arguments.
func (p *TemplatedPrompt) Get(ctx context.Context, args json.RawMessage) (*mcp.GetPromptResult, error) {
	var inputs map[string]any
	if err := json.Unmarshal(args, &inputs); err != nil {
		return nil, err
	}

	messages := make([]*mcp.PromptMessage, len(p.definition.GetMessages()))
	for i, msg := range p.definition.GetMessages() {
		if text := msg.GetText(); text != nil {
			tpl, err := transformer.NewTemplate(text.GetText(), "{{", "}}")
			if err != nil {
				return nil, err
			}
			renderedText, err := tpl.Render(inputs)
			if err != nil {
				return nil, err
			}
			messages[i] = &mcp.PromptMessage{
				Role:    mcp.Role(msg.GetRole()),
				Content: &mcp.TextContent{Text: renderedText},
			}
		}
	}

	return &mcp.GetPromptResult{
		Description: p.definition.GetDescription(),
		Messages:    messages,
	}, nil
}
