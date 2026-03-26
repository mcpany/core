// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package prompt_test

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/prompt"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

var _ = mcpsdk.TextContent{}

// TestTemplatedPrompt_Get ...
// Summary: TestTemplatedPrompt_Get
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	role := configv1.PromptMessage_USER
	definition := configv1.PromptDefinition_builder{
		Name:        proto.String("test-prompt"),
		Title:       proto.String("Test Prompt"),
		Description: proto.String("A test prompt"),
		InputSchema: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"required": {Kind: &structpb.Value_ListValue{ListValue: &structpb.ListValue{Values: []*structpb.Value{{Kind: &structpb.Value_StringValue{StringValue: "name"}}}}}},
				"properties": {Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"name": {Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"description": {Kind: &structpb.Value_StringValue{StringValue: "The name to use in the prompt"}},
								"type":        {Kind: &structpb.Value_StringValue{StringValue: "string"}},
							},
						}}},
					},
				}}},
				"type": {Kind: &structpb.Value_StringValue{StringValue: "object"}},
			},
		},
		Messages: []*configv1.PromptMessage{
			configv1.PromptMessage_builder{
				Role: &role,
				Text: configv1.TextContent_builder{
					Text: proto.String("Hello, {{name}}"),
				}.Build(),
			}.Build(),
		},
	}.Build()
	templatedPrompt, err := prompt.NewTemplatedPrompt(definition, "test-service")
	assert.NoError(t, err)

	t.Run("with string arguments", func(t *testing.T) {
		args, _ := json.Marshal(map[string]string{"name": "world"})
		result, err := templatedPrompt.Get(context.Background(), args)
		assert.NoError(t, err)
		assert.Equal(t, "Hello, world", result.Messages[0].Content.(*mcpsdk.TextContent).Text)
	})

	t.Run("with generic arguments", func(t *testing.T) {
		args, _ := json.Marshal(map[string]any{"name": "world"})
		result, err := templatedPrompt.Get(context.Background(), args)
		assert.NoError(t, err)
		assert.Equal(t, "Hello, world", result.Messages[0].Content.(*mcpsdk.TextContent).Text)
	})
}

// TestTemplatedPrompt_Get_UnmarshalError ...
// Summary: TestTemplatedPrompt_Get_UnmarshalError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	definition := configv1.PromptDefinition_builder{}.Build()
	templatedPrompt, err := prompt.NewTemplatedPrompt(definition, "test-service")
	assert.NoError(t, err)

	_, err = templatedPrompt.Get(context.Background(), []byte("invalid json"))
	assert.Error(t, err)
}

// TestTemplatedPrompt_Get_RenderError ...
// Summary: TestTemplatedPrompt_Get_RenderError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	role := configv1.PromptMessage_USER
	definition := configv1.PromptDefinition_builder{
		Messages: []*configv1.PromptMessage{
			configv1.PromptMessage_builder{
				Role: &role,
				Text: configv1.TextContent_builder{
					Text: proto.String("Hello, {{name}}"),
				}.Build(),
			}.Build(),
		},
	}.Build()
	templatedPrompt, err := prompt.NewTemplatedPrompt(definition, "test-service")
	assert.NoError(t, err)

	args, _ := json.Marshal(map[string]string{})
	_, err = templatedPrompt.Get(context.Background(), args)
	assert.Error(t, err)
}

// TestTemplatedPrompt_Prompt ...
// Summary: TestTemplatedPrompt_Prompt
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	definition := configv1.PromptDefinition_builder{
		Name:        proto.String("test-prompt"),
		Title:       proto.String("Test Prompt"),
		Description: proto.String("A test prompt"),
		InputSchema: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"required": {Kind: &structpb.Value_ListValue{ListValue: &structpb.ListValue{Values: []*structpb.Value{{Kind: &structpb.Value_StringValue{StringValue: "name"}}}}}},
				"properties": {Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"name": {Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"description": {Kind: &structpb.Value_StringValue{StringValue: "The name to use in the prompt"}},
								"type":        {Kind: &structpb.Value_StringValue{StringValue: "string"}},
							},
						}}},
					},
				}}},
				"type": {Kind: &structpb.Value_StringValue{StringValue: "object"}},
			},
		},
	}.Build()
	templatedPrompt, err := prompt.NewTemplatedPrompt(definition, "test-service")
	assert.NoError(t, err)

	mcpPrompt := templatedPrompt.Prompt()

	assert.Equal(t, "test-service.test-prompt", mcpPrompt.Name)
	assert.Equal(t, "Test Prompt", mcpPrompt.Title)
	assert.Equal(t, "A test prompt", mcpPrompt.Description)
	assert.Len(t, mcpPrompt.Arguments, 1)
	assert.Equal(t, "name", mcpPrompt.Arguments[0].Name)
}

// TestTemplatedPrompt_Service ...
// Summary: TestTemplatedPrompt_Service
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	definition := configv1.PromptDefinition_builder{}.Build()
	templatedPrompt, err := prompt.NewTemplatedPrompt(definition, "test-service")
	assert.NoError(t, err)
	assert.Equal(t, "test-service", templatedPrompt.Service())
}

// TestNewPromptFromConfig ...
// Summary: TestNewPromptFromConfig
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	definition := configv1.PromptDefinition_builder{
		Name:        proto.String("test-prompt"),
		Title:       proto.String("Test Prompt"),
		Description: proto.String("A test prompt"),
	}.Build()

	p, err := prompt.NewPromptFromConfig(definition, "test-service")
	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, "test-service", p.Service())
	assert.Equal(t, "test-service.test-prompt", p.Prompt().Name)
}

// TestTemplatedPrompt_Get_NoText ...
// Summary: TestTemplatedPrompt_Get_NoText
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	role := configv1.PromptMessage_USER
	definition := configv1.PromptDefinition_builder{
		Messages: []*configv1.PromptMessage{
			configv1.PromptMessage_builder{
				Role: &role,
				// No Text set
			}.Build(),
		},
	}.Build()
	templatedPrompt, err := prompt.NewTemplatedPrompt(definition, "test-service")
	assert.NoError(t, err)

	args, _ := json.Marshal(map[string]string{})
	result, err := templatedPrompt.Get(context.Background(), args)
	assert.NoError(t, err)
	assert.Len(t, result.Messages, 1)
	assert.Nil(t, result.Messages[0]) // It should be nil if we didn't populate it
}
