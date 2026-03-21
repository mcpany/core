// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestReproFileSecretValidation(t *testing.T) {
	// Create a temp file with invalid content
	tmpFile, err := os.CreateTemp("", "secret")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("invalid-key")
	require.NoError(t, err)
	tmpFile.Close()

	oldIsAllowed := validation.IsAllowedPath
	defer func() { validation.IsAllowedPath = oldIsAllowed }()
	validation.IsAllowedPath = func(path string) error { return nil }

	oldFileExists := validation.FileExists
	defer func() { validation.FileExists = oldFileExists }()
	validation.FileExists = func(path string) error { return nil }

	oldLookPath := execLookPath
	defer func() { execLookPath = oldLookPath }()
	execLookPath = func(file string) (string, error) {
		return "/bin/ls", nil
	}

	config := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Name: proto.String("test-file-secret"),
				McpService: configv1.McpUpstreamService_builder{
					StdioConnection: configv1.McpStdioConnection_builder{
						Command: proto.String("ls"),
						Env: map[string]*configv1.SecretValue{
							"TEST_KEY": configv1.SecretValue_builder{
								FilePath:        proto.String(tmpFile.Name()),
								ValidationRegex: proto.String("^sk-[a-zA-Z0-9]{10}$"),
							}.Build(),
						},
					}.Build(),
				}.Build(),
			}.Build(),
		},
	}.Build()

	errs := Validate(context.Background(), config, Server)

	assert.NotEmpty(t, errs, "Validation errors expected")
	if len(errs) > 0 {
		// Update: Regex validation for file paths is now disabled for security reasons (Blind File Read)
		assert.Contains(t, errs[0].Err.Error(), "validation regex is not supported for secret file paths")
	}
}
