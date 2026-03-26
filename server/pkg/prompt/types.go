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
//
// Summary: Represents a ErrPromptNotFound.
var ErrPromptNotFound = errors.New("prompt not found")

// Prompt prompt represents a prompt.
//
// Summary: Prompt represents a prompt.
type Prompt interface {
	// Prompt returns the MCP prompt definition.
	//
	// Returns: - None.
	//   - *mcp.Prompt: The MCP prompt definition.
	Prompt() *mcp.Prompt

	// Service returns the ID of the service that provides this prompt.
	//
	// Returns: - None.
	//   - string: The service ID.
	Service() string

	// Definition returns the raw configuration definition of the prompt.
	//
	// Returns: - None.
	//   - *configv1.PromptDefinition: The prompt definition.
	Definition() *configv1.PromptDefinition

	// Get executes the prompt with the provided arguments.
	//
	// Parameters: - None.
	//   - ctx: The context for the request.
	//   - args: The arguments for the prompt as a raw JSON message.
	//
	// Returns: - None.
	//   - *mcp.GetPromptResult: The result of the prompt execution.
	//   - error: An error if the operation fails.
	Get(ctx context.Context, args json.RawMessage) (*mcp.GetPromptResult, error)
}

// MCPServerProvider mCPServerProvider represents a mcp server provider.
//
// Summary: MCPServerProvider represents a mcp server provider.
type MCPServerProvider interface {
	// Server returns the underlying MCP server instance.
	//
	// Returns: - None.
	//   - *mcp.Server: The MCP server instance.
	Server() *mcp.Server
}

// TemplatedPrompt templatedPrompt represents a templated prompt.
//
// Summary: TemplatedPrompt represents a templated prompt.
type TemplatedPrompt struct {
	definition        *configv1.PromptDefinition
	serviceID         string
	compiledTemplates []*transformer.TextTemplate
}

// NewTemplatedPrompt creates a new TemplatedPrompt instance.
//
// Summary: Initializes a new TemplatedPrompt.
//
// Parameters: - None.
//   - definition: The prompt definition from configuration.
//   - serviceID: The ID of the service providing the prompt.
//
// Returns: - None.
//   - *TemplatedPrompt: The initialized TemplatedPrompt.
//   - error: An error if the prompt templates cannot be compiled.
//
// Errors: - None.
//   - Returns error if template parsing fails.
//
// Side Effects: - None.
//   - Compiles all message templates.
func NewTemplatedPrompt(definition *configv1.PromptDefinition, serviceID string) (*TemplatedPrompt, error) {
	// ⚡ BOLT: Pre-compile templates to avoid parsing on every request.
	// Randomized Selection from Top 5 High-Impact Targets
	compiledTemplates := make([]*transformer.TextTemplate, len(definition.GetMessages()))
	for i, msg := range definition.GetMessages() {
		if text := msg.GetText(); text != nil {
			tpl, err := transformer.NewTemplate(text.GetText(), "{{", "}}")
			if err != nil {
				return nil, err
			}
			compiledTemplates[i] = tpl
		}
	}

	return &TemplatedPrompt{
		definition:        definition,
		serviceID:         serviceID,
		compiledTemplates: compiledTemplates,
	}, nil
}

// Prompt prompt prompt.
//
// Summary: Prompt prompt.
//
// Parameters:
//   - None.
//
// Returns:
//   - *mcp.Prompt: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

// Service service service.
//
// Summary: Service service.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (p *TemplatedPrompt) Service() string {
	return p.serviceID
}

// Definition definition definition.
//
// Summary: Definition definition.
//
// Parameters:
//   - None.
//
// Returns:
//   - *configv1.PromptDefinition: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (p *TemplatedPrompt) Definition() *configv1.PromptDefinition {
	return p.definition
}

// Get executes the prompt with the provided arguments.
//
// Summary: Executes the prompt.
//
// It renders the prompt template using the provided arguments.
//
// Parameters: - None.
//   - _: The context (unused in this implementation).
//   - args: The arguments for the prompt as a raw JSON message.
//
// Returns: - None.
//   - *mcp.GetPromptResult: The result of the prompt execution.
//   - error: An error if the operation fails (e.g., template rendering error).
//
// Errors: - None.
//   - Returns error if args cannot be unmarshaled.
//   - Returns error if template rendering fails.
func (p *TemplatedPrompt) Get(_ context.Context, args json.RawMessage) (*mcp.GetPromptResult, error) {
	var inputs map[string]any
	if err := json.Unmarshal(args, &inputs); err != nil {
		return nil, err
	}

	messages := make([]*mcp.PromptMessage, len(p.definition.GetMessages()))
	for i, msg := range p.definition.GetMessages() {
		if text := msg.GetText(); text != nil {
			// Use pre-compiled template
			tpl := p.compiledTemplates[i]
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
// Summary: Creates a Prompt from configuration.
//
// Parameters: - None.
//   - definition: The prompt definition from configuration.
//   - serviceID: The ID of the service providing the prompt.
//
// Returns: - None.
//   - Prompt: The created Prompt instance.
//   - error: An error if the prompt cannot be created.
func NewPromptFromConfig(definition *configv1.PromptDefinition, serviceID string) (Prompt, error) {
	return NewTemplatedPrompt(definition, serviceID)
}
