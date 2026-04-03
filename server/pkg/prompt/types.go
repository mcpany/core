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

// ErrPromptNotFound represents the public ErrPromptNotFound entity.
//
// Summary: Defines the structured data model representing a prompt not found.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
var ErrPromptNotFound = errors.New("prompt not found")

// Prompt represents the public Prompt entity.
//
// Summary: Defines the structured data model representing a .
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

	// Definition returns the raw configuration definition of the prompt.
	//
	// Returns:
	//   - *configv1.PromptDefinition: The prompt definition.
	Definition() *configv1.PromptDefinition

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

// MCPServerProvider represents the public MCPServerProvider entity.
//
// Summary: Defines the structured data model representing a server provider.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type MCPServerProvider interface {
	// Server returns the underlying MCP server instance.
	//
	// Returns:
	//   - *mcp.Server: The MCP server instance.
	Server() *mcp.Server
}

// TemplatedPrompt represents the public TemplatedPrompt entity.
//
// Summary: Defines the structured data model representing a prompt.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type TemplatedPrompt struct {
	definition        *configv1.PromptDefinition
	serviceID         string
	compiledTemplates []*transformer.TextTemplate
}

// NewTemplatedPrompt serves as a public interface for interacting with NewTemplatedPrompt.
//
// Summary: Constructs and returns an initialized templated prompt ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// Prompt serves as a public interface for interacting with Prompt.
//
// Summary: Prompt the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// Service serves as a public interface for interacting with Service.
//
// Summary: Service the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (p *TemplatedPrompt) Service() string {
	return p.serviceID
}

// Definition serves as a public interface for interacting with Definition.
//
// Summary: Definition the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (p *TemplatedPrompt) Definition() *configv1.PromptDefinition {
	return p.definition
}

// Get serves as a public interface for interacting with Get.
//
// Summary: Fetches and returns the underlying  from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// NewPromptFromConfig serves as a public interface for interacting with NewPromptFromConfig.
//
// Summary: Constructs and returns an initialized prompt from config ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewPromptFromConfig(definition *configv1.PromptDefinition, serviceID string) (Prompt, error) {
	return NewTemplatedPrompt(definition, serviceID)
}
