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
	auth := &configv1.Authentication{
		AuthMethod: &configv1.Authentication_Oauth2{
			Oauth2: &configv1.OAuth2Auth{
				ClientId: &configv1.SecretValue{
					Value: &configv1.SecretValue_PlainText{PlainText: "client-id"},
				},
				ClientSecret: &configv1.SecretValue{
					Value: &configv1.SecretValue_PlainText{PlainText: "client-secret"},
				},
			},
		},
	}

	StripSecretsFromAuth(auth)

	oauth := auth.GetOauth2()
	assert.NotNil(t, oauth)

	// ClientID should be scrubbed now
	assert.Nil(t, oauth.ClientId.Value, "Plain text ClientId should be cleared")

	// ClientSecret should be scrubbed
	assert.Nil(t, oauth.ClientSecret.Value, "Plain text ClientSecret should be cleared")
}

func TestStripSecretsFromService_MoreTypes(t *testing.T) {
	// gRPC Service (currently no-op but need coverage)
	grpcSvc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_GrpcService{
			GrpcService: &configv1.GrpcUpstreamService{
				Address: proto.String("localhost:50051"),
			},
		},
	}
	StripSecretsFromService(grpcSvc)
	assert.Equal(t, "localhost:50051", grpcSvc.GetGrpcService().GetAddress())

	// OpenAPI Service (currently no-op)
	openapiSvc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_OpenapiService{
			OpenapiService: &configv1.OpenapiUpstreamService{
				Address: proto.String("http://api.example.com"),
			},
		},
	}
	StripSecretsFromService(openapiSvc)
	assert.Equal(t, "http://api.example.com", openapiSvc.GetOpenapiService().GetAddress())
}

func TestStripSecretsFromMcpService_Calls(t *testing.T) {
	mcpSvc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
			McpService: &configv1.McpUpstreamService{
				Calls: map[string]*configv1.MCPCallDefinition{
					"call1": {},
				},
			},
		},
	}
	StripSecretsFromService(mcpSvc)
	// Just ensuring it doesn't panic and code path is executed
	assert.NotNil(t, mcpSvc.GetMcpService().Calls["call1"])
}
