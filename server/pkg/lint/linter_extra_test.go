// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package lint

import (
	"context"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLinter_Run_MaxConnections(t *testing.T) {
	cfg := configv1.McpAnyServerConfig_builder{
		GlobalSettings: configv1.GlobalSettings_builder{
			MaxConnections: proto.Int32(2000),
		}.Build(),
	}.Build()

	linter := NewLinter(cfg)
	results, err := linter.Run(context.Background())
	assert.NoError(t, err)

	found := false
	for _, r := range results {
		if r.Severity == Info &&
			strings.Contains(r.Message,
				"High max connections") {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected info about high connections")
}

func TestLinter_Run_EmptyServerId(t *testing.T) {
	cfg := configv1.McpAnyServerConfig_builder{
		ServerId: ptr(""),
	}.Build()

	linter := NewLinter(cfg)
	results, err := linter.Run(context.Background())
	assert.NoError(t, err)

	found := false
	for _, r := range results {
		if r.Severity == Warning &&
			strings.Contains(r.Message,
				"Server ID is empty") {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected warning about empty Server ID")
}

func TestLinter_Run_DuplicateServiceId(t *testing.T) {
	cfg := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Id: ptr("duplicate"),
				HttpService: configv1.
					HttpUpstreamService_builder{
					Address: ptr("https://api1.com"),
				}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Id: ptr("duplicate"),
				HttpService: configv1.
					HttpUpstreamService_builder{
					Address: ptr("https://api2.com"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	linter := NewLinter(cfg)
	results, err := linter.Run(context.Background())
	assert.NoError(t, err)

	found := false
	for _, r := range results {
		if r.Severity == Error &&
			strings.Contains(r.Message,
				"Duplicate service ID: duplicate") {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected error about duplicate ID")
}
