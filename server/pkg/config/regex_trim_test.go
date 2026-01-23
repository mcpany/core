// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestRegexTrimValidationBug(t *testing.T) {
	// Setup: plain_text value with spaces
	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("test-trim-bug"),
				ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
					McpService: &configv1.McpUpstreamService{
						ConnectionType: &configv1.McpUpstreamService_StdioConnection{
							StdioConnection: &configv1.McpStdioConnection{
								Command: proto.String("ls"),
								Env: map[string]*configv1.SecretValue{
									"TEST_TRIM": {
										Value: &configv1.SecretValue_PlainText{
											PlainText: " value ",
										},
										ValidationRegex: proto.String("^value$"),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// This assumes that execLookPath is mocked or "ls" exists.
	// We need to mock execLookPath to avoid dependency on system "ls" or PATH.
	// But `validator.go` uses `execLookPath` var which we can override in test package.

	oldLookPath := execLookPath
	defer func() { execLookPath = oldLookPath }()
	execLookPath = func(file string) (string, error) {
		return "/bin/ls", nil
	}

	errs := Validate(context.Background(), config, Server)

	// Expectation: It should pass because validation logic now trims the value.
	assert.Empty(t, errs, "Validation errors not expected")
}
