// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestStripSecretsFromAuth_OAuth2(t *testing.T) {
	auth := configv1.Authentication_builder{
		Oauth2: configv1.OAuth2Auth_builder{
			ClientId: configv1.SecretValue_builder{
				PlainText: proto.String("client-id"),
			}.Build(),
			ClientSecret: configv1.SecretValue_builder{
				PlainText: proto.String("client-secret"),
			}.Build(),
		}.Build(),
	}.Build()

	StripSecretsFromAuth(auth)

	scrubbedOauth := auth.GetOauth2()
	assert.NotNil(t, scrubbedOauth)

	// ClientID should be scrubbed now
	assert.Empty(t, scrubbedOauth.GetClientId().GetPlainText(), "Plain text ClientId should be cleared")

	// ClientSecret should be scrubbed
	assert.Empty(t, scrubbedOauth.GetClientSecret().GetPlainText(), "Plain text ClientSecret should be cleared")
}

func TestStripSecretsFromService_MoreTypes(t *testing.T) {
	// gRPC Service (currently no-op but need coverage)
	grpcSvc := func() *configv1.UpstreamServiceConfig {
		return configv1.UpstreamServiceConfig_builder{
			GrpcService: configv1.GrpcUpstreamService_builder{
				Address: proto.String("127.0.0.1:50051"),
			}.Build(),
		}.Build()
	}()
	StripSecretsFromService(grpcSvc)
	assert.Equal(t, "127.0.0.1:50051", grpcSvc.GetGrpcService().GetAddress())

	// OpenAPI Service (currently no-op)
	openapiSvc := func() *configv1.UpstreamServiceConfig {
		return configv1.UpstreamServiceConfig_builder{
			OpenapiService: configv1.OpenapiUpstreamService_builder{
				Address: proto.String("http://api.example.com"),
			}.Build(),
		}.Build()
	}()
	StripSecretsFromService(openapiSvc)
	assert.Equal(t, "http://api.example.com", openapiSvc.GetOpenapiService().GetAddress())
}

func TestStripSecretsFromMcpService_Calls(t *testing.T) {
	mcpSvc := func() *configv1.UpstreamServiceConfig {
		return configv1.UpstreamServiceConfig_builder{
			McpService: configv1.McpUpstreamService_builder{
				Calls: map[string]*configv1.MCPCallDefinition{
					"call1": configv1.MCPCallDefinition_builder{}.Build(),
				},
			}.Build(),
		}.Build()
	}()
	StripSecretsFromService(mcpSvc)
	// Just ensuring it doesn't panic and code path is executed
	assert.NotNil(t, mcpSvc.GetMcpService().GetCalls()["call1"])
}
