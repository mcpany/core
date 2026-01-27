// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestDisabledServiceIsIncluded(t *testing.T) {
	// Create a manager with default profile
	manager := config.NewUpstreamServiceManager([]string{"default"})

	// Create a configuration with one disabled service
	// Create a configuration with one disabled service
	http1 := configv1.HttpUpstreamService_builder{
		Address: proto.String("http://example.com"),
	}.Build()

	s1 := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String("disabled-service"),
		Disable:     proto.Bool(true),
		HttpService: http1,
	}.Build()

	http2 := configv1.HttpUpstreamService_builder{
		Address: proto.String("http://example.com"),
	}.Build()

	s2 := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String("enabled-service"),
		Disable:     proto.Bool(false),
		HttpService: http2,
	}.Build()

	cfg := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{s1, s2},
	}.Build()

	// Load and merge services
	services, err := manager.LoadAndMergeServices(context.Background(), cfg)
	require.NoError(t, err)

	// Check if the disabled service is present in the output
	// We expect 2 services, one disabled and one enabled.

	// Print names for debugging
	for _, s := range services {
		t.Logf("Service: %s, Disabled: %v", s.GetName(), s.GetDisable())
	}

	// Assertions will likely fail if bug is present
	foundDisabled := false
	for _, s := range services {
		if s.GetName() == "disabled-service" {
			foundDisabled = true
			assert.True(t, s.GetDisable(), "Service should be marked disabled")
		}
	}
	assert.True(t, foundDisabled, "Disabled service should be present in the list")
	assert.Len(t, services, 2, "Should have 2 services")
}
