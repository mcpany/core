// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestValidate_Collection_InvalidURL(t *testing.T) {
	config := configv1.McpAnyServerConfig_builder{
		Collections: []*configv1.Collection{
			configv1.Collection_builder{
				Name:    proto.String("invalid-collection"),
				HttpUrl: proto.String("not-a-url"),
			}.Build(),
		},
	}.Build()

	errs := Validate(context.Background(), config, Server)

	// This ensures that invalid collections are caught by Validate
	assert.NotEmpty(t, errs, "Expected validation error for invalid URL in collection")
}

func TestValidate_Collection_EmptyURL(t *testing.T) {
	config := configv1.McpAnyServerConfig_builder{
		Collections: []*configv1.Collection{
			configv1.Collection_builder{
				Name:    proto.String("empty-url-collection"),
				HttpUrl: proto.String(""),
			}.Build(),
		},
	}.Build()

	errs := Validate(context.Background(), config, Server)

	// This ensures that invalid collections are caught by Validate
	assert.NotEmpty(t, errs, "Expected validation error for empty URL in collection")
	// Verify exact error if possible or just existence
}

func TestValidate_Collection_InlineWithSkills(t *testing.T) {
	config := configv1.McpAnyServerConfig_builder{
		Collections: []*configv1.Collection{
			configv1.Collection_builder{
				Name:   proto.String("skills-collection"),
				Skills: []string{"test-skill"},
			}.Build(),
		},
	}.Build()

	errs := Validate(context.Background(), config, Server)
	assert.Empty(t, errs, "Expected no validation error for inline collection with skills")
}
