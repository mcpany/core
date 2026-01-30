// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package lint_test

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/lint"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestSecretFormatValidation(t *testing.T) {
	// Construct config using standard proto structs (assuming standard Go proto generation)
	// If builders are enforced, I might need to check generated code.
	// But let's try the builder with HttpService field directly.

	cfg := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Name: proto.String("insecure_service"),
				HttpService: configv1.HttpUpstreamService_builder{
					Address: proto.String("https://api.openai.com/v1?key=sk-1234567890abcdef1234567890abcdef12345678"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	linter := lint.NewLinter(cfg)
	results, err := linter.Run(context.Background())
	assert.NoError(t, err)

	found := false
	for _, r := range results {
		if r.Severity == lint.Warning && r.Message == "Value contains potential OpenAI API Key" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should detect OpenAI API Key in URL")
}
