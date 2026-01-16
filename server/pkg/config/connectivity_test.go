// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestCheckConnectivity_HTTP(t *testing.T) {
	// Start a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("valid-http"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: proto.String(server.URL),
					},
				},
			},
			{
				Name: proto.String("invalid-http"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: proto.String("http://localhost:54321"), // Assuming this port is closed
					},
				},
			},
		},
	}

	errors := CheckConnectivity(context.Background(), cfg)

	// We expect one error for the invalid service
	foundInvalid := false
	for _, err := range errors {
		if err.ServiceName == "invalid-http" {
			foundInvalid = true
		}
		if err.ServiceName == "valid-http" {
			t.Errorf("Unexpected error for valid service: %v", err)
		}
	}
	assert.True(t, foundInvalid, "Should have found error for invalid service")
}

func TestCheckConnectivity_Disabled(t *testing.T) {
	cfg := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name:    proto.String("disabled-service"),
				Disable: proto.Bool(true),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: proto.String("http://invalid.address.example.com"),
					},
				},
			},
		},
	}

	errors := CheckConnectivity(context.Background(), cfg)
	assert.Empty(t, errors, "Should not return errors for disabled services")
}

func TestCheckConnectivity_Webrtc(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("valid-webrtc"),
				ServiceConfig: &configv1.UpstreamServiceConfig_WebrtcService{
					WebrtcService: &configv1.WebrtcUpstreamService{
						Address: proto.String(server.URL),
					},
				},
			},
		},
	}

	errors := CheckConnectivity(context.Background(), cfg)
	assert.Empty(t, errors)
}

func TestCheckConnectivity_Graphql(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("valid-graphql"),
				ServiceConfig: &configv1.UpstreamServiceConfig_GraphqlService{
					GraphqlService: &configv1.GraphQLUpstreamService{
						Address: proto.String(server.URL),
					},
				},
			},
		},
	}

	errors := CheckConnectivity(context.Background(), cfg)
	assert.Empty(t, errors)
}

func TestCheckConnectivity_OpenAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Test Address
	cfg1 := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("valid-openapi-addr"),
				ServiceConfig: &configv1.UpstreamServiceConfig_OpenapiService{
					OpenapiService: &configv1.OpenapiUpstreamService{
						Address: proto.String(server.URL),
					},
				},
			},
		},
	}
	assert.Empty(t, CheckConnectivity(context.Background(), cfg1))

	// Test SpecURL
	cfg2 := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("valid-openapi-url"),
				ServiceConfig: &configv1.UpstreamServiceConfig_OpenapiService{
					OpenapiService: &configv1.OpenapiUpstreamService{
						SpecSource: &configv1.OpenapiUpstreamService_SpecUrl{
							SpecUrl: server.URL,
						},
					},
				},
			},
		},
	}
	assert.Empty(t, CheckConnectivity(context.Background(), cfg2))
}

func TestCheckConnectivity_MCP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("valid-mcp-http"),
				ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
					McpService: &configv1.McpUpstreamService{
						ConnectionType: &configv1.McpUpstreamService_HttpConnection{
							HttpConnection: &configv1.McpStreamableHttpConnection{
								HttpAddress: proto.String(server.URL),
							},
						},
					},
				},
			},
			{
				Name: proto.String("valid-mcp-stdio"),
				ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
					McpService: &configv1.McpUpstreamService{
						ConnectionType: &configv1.McpUpstreamService_StdioConnection{
							StdioConnection: &configv1.McpStdioConnection{
								Command: proto.String("echo"),
							},
						},
					},
				},
			},
		},
	}
	// Stdio is not checked for connectivity, so it should pass.
	errors := CheckConnectivity(context.Background(), cfg)
	assert.Empty(t, errors)
}

func TestCheckConnectivity_HeadFailsButGetSucceeds(t *testing.T) {
	// Server that fails HEAD but accepts GET
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("head-fail"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: proto.String(server.URL),
					},
				},
			},
		},
	}

	errors := CheckConnectivity(context.Background(), cfg)
	assert.Empty(t, errors)
}

func TestCheckConnectivity_GRPC_Mock(t *testing.T) {
	// Just test that it returns error for closed port
	cfg := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("invalid-grpc"),
				ServiceConfig: &configv1.UpstreamServiceConfig_GrpcService{
					GrpcService: &configv1.GrpcUpstreamService{
						Address: proto.String("localhost:54321"),
					},
				},
			},
		},
	}

	errors := CheckConnectivity(context.Background(), cfg)
	assert.NotEmpty(t, errors)
}
