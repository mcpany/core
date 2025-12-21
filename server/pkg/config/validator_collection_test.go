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

func TestValidate_UpstreamServiceCollection_InvalidURL(t *testing.T) {
	config := &configv1.McpAnyServerConfig{
		UpstreamServiceCollections: []*configv1.UpstreamServiceCollection{
			{
				Name:    proto.String("invalid-collection"),
				HttpUrl: proto.String("not-a-url"),
			},
		},
	}

	errs := Validate(context.Background(), config, Server)

	// This ensures that invalid collections are caught by Validate
	assert.NotEmpty(t, errs, "Expected validation error for invalid URL in collection")
}

func TestValidate_UpstreamServiceCollection_EmptyURL(t *testing.T) {
	config := &configv1.McpAnyServerConfig{
		UpstreamServiceCollections: []*configv1.UpstreamServiceCollection{
			{
				Name:    proto.String("empty-url-collection"),
				HttpUrl: proto.String(""),
			},
		},
	}

	errs := Validate(context.Background(), config, Server)

	// This ensures that invalid collections are caught by Validate
	assert.NotEmpty(t, errs, "Expected validation error for empty URL in collection")
}
