// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"context"
	"encoding/json"
	"errors"
	"sort"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/transformer"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ErrPromptNotFound is returned when a requested prompt is not found.
var ErrPromptNotFound = errors.New("prompt not found")

// Prompt is the fundamental interface for any executable prompt in the system.
type Prompt interface {
	// Prompt returns the MCP prompt definition.
	//
	// Returns:
	//   - *mcp.Prompt: The MCP prompt definition.
	Prompt() *mcp.Prompt

	// Service returns the ID of the service that provides this prompt.
	//
	// Returns:
	//   - string: The service ID.
	Service() string

	// Get executes the prompt with the provided arguments.
	//
	// Parameters:
	//   - ctx: The context for the request.
	//   - args: The arguments for the prompt as a raw JSON message.
	//
	// Returns:
	//   - *mcp.GetPromptResult: The result of the prompt execution.
	//   - error: An error if the operation fails.
	Get(ctx context.Context, args json.RawMessage) (*mcp.GetPromptResult, error)
}

// MCPServerProvider defines an interface for components that can provide an instance of an *mcp.Server.
//
// This is used to decouple the Manager from the concrete server implementation.
type MCPServerProvider interface {
	// Server returns the underlying MCP server instance.
	//
	// Returns:
	//   - *mcp.Server: The MCP server instance.
	Server() *mcp.Server
}

// TemplatedPrompt implements the Prompt interface for a prompt that is defined by a template.
type TemplatedPrompt struct {
	definition *configv1.PromptDefinition
	serviceID  string
}

// NewTemplatedPrompt creates a new TemplatedPrompt instance.
//
// Parameters:
//   - definition: The prompt definition from configuration.
//   - serviceID: The ID of the service providing the prompt.
//
// Returns:
//   - *TemplatedPrompt: The initialized TemplatedPrompt.
func NewTemplatedPrompt(definition *configv1.PromptDefinition, serviceID string) *TemplatedPrompt {
	return &TemplatedPrompt{
		definition: definition,
		serviceID:  serviceID,
	}
}

// Prompt returns the MCP prompt definition.
//
// Returns:
//   - *mcp.Prompt: The MCP prompt definition.
func (p *TemplatedPrompt) Prompt() *mcp.Prompt {
	args := make([]*mcp.PromptArgument, 0)
	if p.definition.GetInputSchema() != nil {
		fields := p.definition.GetInputSchema().GetFields()
		if props, ok := fields["properties"]; ok {
			if propsStruct := props.GetStructValue(); propsStruct != nil {
				// Collect keys to sort them for deterministic order
				keys := make([]string, 0, len(propsStruct.GetFields()))
				for k := range propsStruct.GetFields() {
					keys = append(keys, k)
				}
				sort.Strings(keys)

				for _, name := range keys {
					val := propsStruct.GetFields()[name]
					desc := ""
					if valStruct := val.GetStructValue(); valStruct != nil {
						if d, ok := valStruct.GetFields()["description"]; ok {
							desc = d.GetStringValue()
						}
					}

					required := false
					if req, ok := fields["required"]; ok {
						if reqList := req.GetListValue(); reqList != nil {
							for _, v := range reqList.GetValues() {
								if v.GetStringValue() == name {
									required = true
									break
								}
							}
						}
					}

					args = append(args, &mcp.PromptArgument{
						Name:        name,
						Description: desc,
						Required:    required,
					})
				}
			}
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
//
// Returns:
//   - string: The service ID.
func (p *TemplatedPrompt) Service() string {
	return p.serviceID
}

// Get executes the prompt with the provided arguments.
//
// It renders the prompt template using the provided arguments.
//
// Parameters:
//   - _ : The context (unused in this implementation).
//   - args: The arguments for the prompt as a raw JSON message.
//
// Returns:
//   - *mcp.GetPromptResult: The result of the prompt execution.
//   - error: An error if the operation fails (e.g., template rendering error).
func (p *TemplatedPrompt) Get(_ context.Context, args json.RawMessage) (*mcp.GetPromptResult, error) {
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

// NewPromptFromConfig creates a new Prompt from a configuration definition.
//
// Parameters:
//   - definition: The prompt definition from configuration.
//   - serviceID: The ID of the service providing the prompt.
//
// Returns:
//   - Prompt: The created Prompt instance.
//   - error: An error if the prompt cannot be created.
func NewPromptFromConfig(definition *configv1.PromptDefinition, serviceID string) (Prompt, error) {
	return NewTemplatedPrompt(definition, serviceID), nil
}
