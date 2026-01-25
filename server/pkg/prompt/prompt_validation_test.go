// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package prompt_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/server/pkg/prompt"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestNewPromptFromConfig_EmptyName(t *testing.T) {
	name := ""
	definition := configv1.PromptDefinition_builder{
		Name: &name,
		Title: proto.String("Test Prompt"),
	}.Build()

	// Now this should return error
	p, err := prompt.NewPromptFromConfig(definition, "test-service")

	assert.Error(t, err)
	assert.Nil(t, p)
	assert.Contains(t, err.Error(), "invalid prompt name")
}

func TestTemplatedPrompt_Prompt_Fallback(t *testing.T) {
	name := ""
	definition := configv1.PromptDefinition_builder{
		Name: &name,
	}.Build()

	// Direct usage of TemplatedPrompt via NewTemplatedPrompt (factory)
	p := prompt.NewTemplatedPrompt(definition, "test-service")
	mcpPrompt := p.Prompt()

	// Fixed behavior: "test-service.unnamed"
	assert.Equal(t, "test-service.unnamed", mcpPrompt.Name)
}

func TestTemplatedPrompt_Get_NilText(t *testing.T) {
	role := configv1.PromptMessage_USER
	definition := configv1.PromptDefinition_builder{
		Messages: []*configv1.PromptMessage{
			configv1.PromptMessage_builder{
				Role: &role,
				// No Text set
			}.Build(),
		},
	}.Build()
	templatedPrompt := prompt.NewTemplatedPrompt(definition, "test-service")

	args, _ := json.Marshal(map[string]string{})
	_, err := templatedPrompt.Get(context.Background(), args)

	// Fixed behavior: Returns error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing text content")
}

func TestTemplatedPrompt_Get_TemplateParseError(t *testing.T) {
	role := configv1.PromptMessage_USER
	definition := configv1.PromptDefinition_builder{
		Messages: []*configv1.PromptMessage{
			configv1.PromptMessage_builder{
				Role: &role,
				Text: configv1.TextContent_builder{
					Text: proto.String("Hello, {{"), // Invalid syntax
				}.Build(),
			}.Build(),
		},
	}.Build()
	templatedPrompt := prompt.NewTemplatedPrompt(definition, "test-service")

	args, _ := json.Marshal(map[string]string{})
	_, err := templatedPrompt.Get(context.Background(), args)

	assert.Error(t, err)
}
