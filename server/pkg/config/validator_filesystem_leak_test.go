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

func TestValidate_Leak_McpService(t *testing.T) {
	// This test reproduces an issue where filesystem checks are performed even when SkipFilesystemCheckKey is set.
	// We use a path that is structurally valid (no traversal) but clearly not allowed (e.g. /tmp/non_existent_check).
	// If the check is skipped, validation should pass (or at least pass the filesystem check).
	// If the check is NOT skipped, IsAllowedPath will fail because /tmp is not in CWD/AllowedPaths.

	ctx := context.WithValue(context.Background(), SkipFilesystemCheckKey, true)

	t.Run("stdio_connection working_directory leak", func(t *testing.T) {
		cfg := configv1.McpAnyServerConfig_builder{
			UpstreamServices: []*configv1.UpstreamServiceConfig{
				configv1.UpstreamServiceConfig_builder{
					Name: proto.String("leak-stdio"),
					McpService: configv1.McpUpstreamService_builder{
						StdioConnection: configv1.McpStdioConnection_builder{
							Command:          proto.String("echo"), // Valid command
							WorkingDirectory: proto.String("/tmp/should_be_skipped"),
						}.Build(),
					}.Build(),
				}.Build(),
			},
		}.Build()

		validationErrors := Validate(ctx, cfg, Server)
		// Expect NO errors because we are skipping filesystem checks.
		// If bug exists, we will get "insecure working_directory" error.
		assert.Empty(t, validationErrors, "Should skip filesystem check for working_directory")
	})

	t.Run("bundle_connection bundle_path leak", func(t *testing.T) {
		cfg := configv1.McpAnyServerConfig_builder{
			UpstreamServices: []*configv1.UpstreamServiceConfig{
				configv1.UpstreamServiceConfig_builder{
					Name: proto.String("leak-bundle"),
					McpService: configv1.McpUpstreamService_builder{
						BundleConnection: configv1.McpBundleConnection_builder{
							BundlePath: proto.String("/tmp/should_be_skipped.bundle"),
						}.Build(),
					}.Build(),
				}.Build(),
			},
		}.Build()

		validationErrors := Validate(ctx, cfg, Server)
		// Expect NO errors because we are skipping filesystem checks.
		assert.Empty(t, validationErrors, "Should skip filesystem check for bundle_path")
	})

	t.Run("insecure path should fail even when skipped", func(t *testing.T) {
		cfg := configv1.McpAnyServerConfig_builder{
			UpstreamServices: []*configv1.UpstreamServiceConfig{
				configv1.UpstreamServiceConfig_builder{
					Name: proto.String("leak-stdio-insecure"),
					McpService: configv1.McpUpstreamService_builder{
						StdioConnection: configv1.McpStdioConnection_builder{
							Command:          proto.String("echo"),
							WorkingDirectory: proto.String("../insecure"),
						}.Build(),
					}.Build(),
				}.Build(),
			},
		}.Build()

		validationErrors := Validate(ctx, cfg, Server)
		assert.NotEmpty(t, validationErrors, "Should fail for insecure path even when skipped")
		if len(validationErrors) > 0 {
			assert.Contains(t, validationErrors[0].Error(), "insecure working_directory")
		}
	})
}
