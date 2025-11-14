// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestNewFromProto(t *testing.T) {
	// Test case 1: Valid PromptDefinition
	validDef := &configv1.PromptDefinition{}
	validDef.SetName("test-prompt")
	validDef.SetDescription("A test prompt")
	validDef.SetTemplate("Hello, {{.name}}!")
	validDef.SetInputSchema(&structpb.Struct{
		Fields: map[string]*structpb.Value{
			"type": {
				Kind: &structpb.Value_StringValue{StringValue: "object"},
			},
			"properties": {
				Kind: &structpb.Value_StructValue{
					StructValue: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"name": {
								Kind: &structpb.Value_StructValue{
									StructValue: &structpb.Struct{
										Fields: map[string]*structpb.Value{
											"type": {
												Kind: &structpb.Value_StringValue{StringValue: "string"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})

	prompt, err := NewFromProto(validDef, "test-service")
	require.NoError(t, err)
	assert.NotNil(t, prompt)
	assert.Equal(t, "test-prompt", prompt.Name)
	assert.Equal(t, "A test prompt", prompt.Description)
	assert.Equal(t, "Hello, {{.name}}!", prompt.Template)
	assert.NotNil(t, prompt.InputSchema)

	// Test case 2: Nil PromptDefinition
	prompt, err = NewFromProto(nil, "test-service")
	require.NoError(t, err)
	assert.Nil(t, prompt)

	// Test case 3: Invalid InputSchema
	invalidDef := &configv1.PromptDefinition{}
	invalidDef.SetName("test-prompt")
	invalidDef.SetDescription("A test prompt")
	invalidDef.SetTemplate("Hello, {{.name}}!")
	invalidDef.SetInputSchema(&structpb.Struct{
		Fields: map[string]*structpb.Value{
			"type": {
				Kind: &structpb.Value_StringValue{StringValue: "invalid-type"},
			},
		},
	})

	prompt, err = NewFromProto(invalidDef, "test-service")
	require.Error(t, err)
	assert.Nil(t, prompt)
}
