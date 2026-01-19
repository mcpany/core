// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestValidate_BugFix_ListenAddressWhitespace(t *testing.T) {
	// Regression test for bug where spaces in listen address caused validation failure
	// This simulates a user providing a config file with accidental spaces in the listen address.
	cfg := &configv1.McpAnyServerConfig{
		GlobalSettings: &configv1.GlobalSettings{
			McpListenAddress: lo.ToPtr(": 8080"), // Note the space
		},
	}

	errs := Validate(context.Background(), cfg, Server)
	assert.Empty(t, errs, "Validation should pass for listen address with space")
}
