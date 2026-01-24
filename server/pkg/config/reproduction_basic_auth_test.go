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

func TestBasicAuthPasswordHashReproduction(t *testing.T) {
	// Setup: User with BasicAuth using password_hash, NO password secret
	config := &configv1.McpAnyServerConfig{
		Users: []*configv1.User{
			{
				Id: proto.String("test-user"),
				Authentication: &configv1.Authentication{
					AuthMethod: &configv1.Authentication_BasicAuth{
						BasicAuth: &configv1.BasicAuth{
							Username:     proto.String("user"),
							PasswordHash: proto.String("$2a$12$R9h/cIPz0gi.URNNX3kh2OPST9/PgBkqquii.V3Ug8bR.7nzvcZvm"), // Example hash
						},
					},
				},
			},
		},
	}

	// Validate
	errs := Validate(context.Background(), config, Server)

	// After fix, this should pass because password_hash is provided for incoming auth (User)
	assert.Empty(t, errs, "Validation errors not expected")
}
