package prompt

import (
	"fmt"

	"github.com/mcpany/core/pkg/mcp"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// Prompt represents a configured prompt that can be executed.
type Prompt struct {
	Name        string
	Description string
	Template    string
	InputSchema *mcp.Schema
	ServiceID   string
}

// NewFromProto creates a new Prompt from a PromptDefinition protobuf message.
func NewFromProto(def *configv1.PromptDefinition, serviceID string) (*Prompt, error) {
	if def == nil {
		return nil, nil
	}

	inputSchema, err := tool.NewSchemaFromProto(def.GetInputSchema())
	if err != nil {
		return nil, fmt.Errorf("failed to create schema from proto: %w", err)
	}

	return &Prompt{
		Name:        def.GetName(),
		Description: def.GetDescription(),
		Template:    def.GetTemplate(),
		InputSchema: inputSchema,
		ServiceID:   serviceID,
	}, nil
}
