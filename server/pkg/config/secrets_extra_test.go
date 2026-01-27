// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestStripSecretsFromAuth_OAuth2(t *testing.T) {
	auth := &configv1.Authentication{}
	oauth := &configv1.OAuth2Auth{}

	clientId := &configv1.SecretValue{}
	clientId.SetPlainText("client-id")
	oauth.SetClientId(clientId)

	clientSecret := &configv1.SecretValue{}
	clientSecret.SetPlainText("client-secret")
	oauth.SetClientSecret(clientSecret)

	auth.SetOauth2(oauth)

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
		svc := &configv1.UpstreamServiceConfig{}
		grpc := &configv1.GrpcUpstreamService{}
		grpc.SetAddress("127.0.0.1:50051")
		svc.SetGrpcService(grpc)
		return svc
	}()
	StripSecretsFromService(grpcSvc)
	assert.Equal(t, "127.0.0.1:50051", grpcSvc.GetGrpcService().GetAddress())

	// OpenAPI Service (currently no-op)
	openapiSvc := func() *configv1.UpstreamServiceConfig {
		svc := &configv1.UpstreamServiceConfig{}
		openapi := &configv1.OpenapiUpstreamService{}
		openapi.SetAddress("http://api.example.com")
		svc.SetOpenapiService(openapi)
		return svc
	}()
	StripSecretsFromService(openapiSvc)
	assert.Equal(t, "http://api.example.com", openapiSvc.GetOpenapiService().GetAddress())
}

func TestStripSecretsFromMcpService_Calls(t *testing.T) {
	mcpSvc := func() *configv1.UpstreamServiceConfig {
		svc := &configv1.UpstreamServiceConfig{}
		mcp := &configv1.McpUpstreamService{}
		mcp.SetCalls(map[string]*configv1.MCPCallDefinition{
			"call1": {},
		})
		svc.SetMcpService(mcp)
		return svc
	}()
	StripSecretsFromService(mcpSvc)
	// Just ensuring it doesn't panic and code path is executed
	assert.NotNil(t, mcpSvc.GetMcpService().GetCalls()["call1"])
}
