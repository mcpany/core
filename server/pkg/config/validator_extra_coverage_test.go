// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestValidateOAuth2Auth_IssuerUrl_Error(t *testing.T) {
	ctx := context.Background()
	// Test case where TokenUrl is empty, so it falls back to IssuerUrl validation
	auth := configv1.OAuth2Auth_builder{
		TokenUrl:  proto.String(""),
		IssuerUrl: proto.String("not-a-url"),
	}.Build()
	err := validateOAuth2Auth(ctx, auth)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid oauth2 issuer_url")
}

func TestValidateSchema_InvalidJsonSchema(t *testing.T) {
	// Test case where schema is structurally valid as structpb but invalid JSON Schema
	// e.g. "type" field has invalid value
	s := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"type": {Kind: &structpb.Value_StringValue{StringValue: "invalid_type_name"}},
		},
	}
	err := validateSchema(s)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JSON schema")
}

func TestMcpService_NoConnectionType(t *testing.T) {
	s := configv1.McpUpstreamService_builder{}.Build()
	err := validateMcpService(context.Background(), s)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no connection_type")
}

func TestValidateAuth_SecretValidationErrors(t *testing.T) {
	ctx := context.Background()

	// Basic Auth: Password secret validation fails (e.g. invalid file path)
	basicAuth := configv1.BasicAuth_builder{
		Username: proto.String("user"),
		Password: configv1.SecretValue_builder{
			FilePath: proto.String("/invalid/path\x00"),
		}.Build(),
	}.Build()
	auth := configv1.Authentication_builder{
		BasicAuth: basicAuth,
	}.Build()
	err := validateUpstreamAuthentication(ctx, auth, AuthValidationContextOutgoing)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "basic auth password validation failed")

	// Bearer Token: Token secret validation fails
	bearerAuth := configv1.BearerTokenAuth_builder{
		Token: configv1.SecretValue_builder{
			FilePath: proto.String("/invalid/path\x00"),
		}.Build(),
	}.Build()
	auth = configv1.Authentication_builder{
		BearerToken: bearerAuth,
	}.Build()
	err = validateUpstreamAuthentication(ctx, auth, AuthValidationContextOutgoing)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bearer token validation failed")

	// OAuth2: ClientID/Secret validation fails
	oauth := configv1.OAuth2Auth_builder{
		TokenUrl: proto.String("https://example.com/token"),
		ClientId: configv1.SecretValue_builder{
			FilePath: proto.String("/invalid/path\x00"),
		}.Build(),
	}.Build()
	auth = configv1.Authentication_builder{
		Oauth2: oauth,
	}.Build()
	err = validateUpstreamAuthentication(ctx, auth, AuthValidationContextOutgoing)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "oauth2 client_id validation failed")
}
