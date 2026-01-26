// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestBuildCommandFromStdioConfig_Validation(t *testing.T) {
	ctx := context.Background()

	t.Run("Validation passed", func(t *testing.T) {
		stdio := configv1.McpStdioConnection_builder{
			Command: proto.String("echo"),
			Env: map[string]*configv1.SecretValue{
				"REQUIRED_VAR": configv1.SecretValue_builder{
					PlainText: proto.String("present"),
				}.Build(),
			},
			Validation: configv1.EnvValidation_builder{
				RequiredEnv: []string{"REQUIRED_VAR"},
			}.Build(),
		}.Build()

		cmd, err := buildCommandFromStdioConfig(ctx, stdio, false)
		assert.NoError(t, err)
		assert.NotNil(t, cmd)
		assert.Contains(t, cmd.Env, "REQUIRED_VAR=present")
	})

	t.Run("Validation failed - missing var", func(t *testing.T) {
		stdio := configv1.McpStdioConnection_builder{
			Command: proto.String("echo"),
			Env:     map[string]*configv1.SecretValue{},
			Validation: configv1.EnvValidation_builder{
				RequiredEnv: []string{"MISSING_VAR"},
			}.Build(),
		}.Build()

		cmd, err := buildCommandFromStdioConfig(ctx, stdio, false)
		assert.Error(t, err)
		assert.Nil(t, cmd)
		assert.Contains(t, err.Error(), "missing required environment variables: MISSING_VAR")
	})

	t.Run("Validation passed - inherited var", func(t *testing.T) {
		t.Setenv("LANG", "C.UTF-8")
		stdio := configv1.McpStdioConnection_builder{
			Command: proto.String("echo"),
			Validation: configv1.EnvValidation_builder{
				RequiredEnv: []string{"LANG"},
			}.Build(),
		}.Build()

		cmd, err := buildCommandFromStdioConfig(ctx, stdio, false)
		assert.NoError(t, err)
		assert.NotNil(t, cmd)
		found := false
		for _, e := range cmd.Env {
			if e == "LANG=C.UTF-8" {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected LANG to be present in cmd.Env")
	})

	t.Run("Validation failed - multiple missing", func(t *testing.T) {
		stdio := configv1.McpStdioConnection_builder{
			Command: proto.String("echo"),
			Validation: configv1.EnvValidation_builder{
				RequiredEnv: []string{"MISSING_1", "MISSING_2"},
			}.Build(),
		}.Build()

		_, err := buildCommandFromStdioConfig(ctx, stdio, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required environment variables: MISSING_1, MISSING_2")
	})

	t.Run("Validation with Docker command", func(t *testing.T) {
		stdio := configv1.McpStdioConnection_builder{
			Command: proto.String("docker"),
			Args:    []string{"run", "ubuntu"},
			Env: map[string]*configv1.SecretValue{
				"DOCKER_ENV": configv1.SecretValue_builder{
					PlainText: proto.String("val"),
				}.Build(),
			},
			Validation: configv1.EnvValidation_builder{
				RequiredEnv: []string{"DOCKER_ENV"},
			}.Build(),
		}.Build()

		cmd, err := buildCommandFromStdioConfig(ctx, stdio, false)
		assert.NoError(t, err)
		assert.NotNil(t, cmd)
		assert.Contains(t, cmd.Env, "DOCKER_ENV=val")
	})

	t.Run("Validation failed with Docker command", func(t *testing.T) {
		stdio := configv1.McpStdioConnection_builder{
			Command: proto.String("docker"),
			Args:    []string{"run", "ubuntu"},
			Validation: configv1.EnvValidation_builder{
				RequiredEnv: []string{"MISSING_DOCKER_VAR"},
			}.Build(),
		}.Build()

		_, err := buildCommandFromStdioConfig(ctx, stdio, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required environment variables: MISSING_DOCKER_VAR")
	})
}

// Helper to construct Proto objects for tests if needed
func strPtr(s string) *string {
	return proto.String(s)
}
