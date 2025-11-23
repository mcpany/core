// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestTemplatedPrompt_Get(t *testing.T) {
	definition := configv1.PromptDefinition_builder{
		Name: proto.String("test-prompt"),
		Messages: []*configv1.PromptMessage{
			func() *configv1.PromptMessage {
				role := configv1.PromptMessage_USER
				return configv1.PromptMessage_builder{
					Role: &role,
					Text: configv1.TextContent_builder{
						Text: proto.String("Hello, {{name}}!"),
					}.Build(),
				}.Build()
			}(),
		},
	}.Build()
	prompt := NewTemplatedPrompt(definition, "test-service")

	args := map[string]string{
		"name": "world",
	}
	argsBytes, err := json.Marshal(args)
	require.NoError(t, err)

	result, err := prompt.Get(context.Background(), argsBytes)
	require.NoError(t, err)

	require.Len(t, result.Messages, 1)
	textContent, ok := result.Messages[0].Content.(*mcp.TextContent)
	require.True(t, ok)
	assert.Equal(t, "Hello, world!", textContent.Text)
}

func TestTemplatedPrompt_Prompt(t *testing.T) {
	definition := configv1.PromptDefinition_builder{
		Name:  proto.String("test-prompt"),
		Title: proto.String("Test Prompt"),
	}.Build()
	prompt := NewTemplatedPrompt(definition, "test-service")

	mcpPrompt := prompt.Prompt()

	assert.Equal(t, "test-service.test-prompt", mcpPrompt.Name)
	assert.Equal(t, "Test Prompt", mcpPrompt.Title)
}

func TestTemplatedPrompt_Service(t *testing.T) {
	definition := configv1.PromptDefinition_builder{}.Build()
	prompt := NewTemplatedPrompt(definition, "test-service")

	serviceID := prompt.Service()

	assert.Equal(t, "test-service", serviceID)
}

func TestTemplatedPrompt_Get_UnmarshalError(t *testing.T) {
	definition := configv1.PromptDefinition_builder{}.Build()
	prompt := NewTemplatedPrompt(definition, "test-service")

	invalidArgs := json.RawMessage(`{"name": world"}`) // Invalid JSON

	_, err := prompt.Get(context.Background(), invalidArgs)

	assert.Error(t, err)
}

func TestTemplatedPrompt_Get_RenderError(t *testing.T) {
	definition := configv1.PromptDefinition_builder{
		Messages: []*configv1.PromptMessage{
			func() *configv1.PromptMessage {
				role := configv1.PromptMessage_USER
				return configv1.PromptMessage_builder{
					Role: &role,
					Text: configv1.TextContent_builder{
						Text: proto.String("Hello, {{name}}!"),
					}.Build(),
				}.Build()
			}(),
		},
	}.Build()
	prompt := NewTemplatedPrompt(definition, "test-service")

	args := map[string]string{} // Missing "name" argument
	argsBytes, err := json.Marshal(args)
	require.NoError(t, err)

	_, err = prompt.Get(context.Background(), argsBytes)

	assert.Error(t, err)
}
