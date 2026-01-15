// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"fmt"
	"net"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestValidateServiceConnection(t *testing.T) {
	// Start a listener on a random port for success case
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer lis.Close()

	successAddr := "http://" + lis.Addr().String()
	successHost, successPort, _ := net.SplitHostPort(lis.Addr().String())

	// Find a free port and don't listen on it for failure case
	l2, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	failAddr := "http://" + l2.Addr().String()
	l2.Close() // Close it immediately so it's unreachable

	// Create OpenAPI config using JSON to handle oneof complexity
	openapiSuccessJSON := fmt.Sprintf(`{
		"name": "test-openapi",
		"openapi_service": {
			"spec_url": "%s"
		}
	}`, successAddr)
	openapiSuccessConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(openapiSuccessJSON), openapiSuccessConfig))

	tests := []struct {
		name        string
		service     *configv1.UpstreamServiceConfig
		expectErr   bool
		errContains string
	}{
		{
			name: "HTTP Success",
			service: &configv1.UpstreamServiceConfig{
				Name: proto.String("test-http"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: proto.String(successAddr),
					},
				},
			},
			expectErr: false,
		},
		{
			name: "HTTP Failure (Connection Refused)",
			service: &configv1.UpstreamServiceConfig{
				Name: proto.String("test-http-fail"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: proto.String(failAddr),
					},
				},
			},
			expectErr:   true,
			errContains: "connectivity check failed",
		},
		{
			name: "gRPC Success",
			service: &configv1.UpstreamServiceConfig{
				Name: proto.String("test-grpc"),
				ServiceConfig: &configv1.UpstreamServiceConfig_GrpcService{
					GrpcService: &configv1.GrpcUpstreamService{
						Address: proto.String(net.JoinHostPort(successHost, successPort)),
					},
				},
			},
			expectErr: false,
		},
		{
			name: "gRPC Failure",
			service: &configv1.UpstreamServiceConfig{
				Name: proto.String("test-grpc-fail"),
				ServiceConfig: &configv1.UpstreamServiceConfig_GrpcService{
					GrpcService: &configv1.GrpcUpstreamService{
						Address: proto.String("localhost:1"), // Likely closed
					},
				},
			},
			expectErr:   true,
			errContains: "connectivity check failed",
		},
		{
			name: "Disabled Service (Should Skip)",
			// Service populated below using protojson to avoid *bool issues
			service: nil,
			expectErr: false,
		},
		{
			name: "Non-Network Service (Should Skip)",
			service: &configv1.UpstreamServiceConfig{
				Name: proto.String("test-cmd"),
				ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
					CommandLineService: &configv1.CommandLineUpstreamService{
						Command: proto.String("ls"),
					},
				},
			},
			expectErr: false,
		},
		{
			name:      "OpenAPI Spec URL Success",
			service:   openapiSuccessConfig,
			expectErr: false,
		},
		{
			name: "Invalid Address Format",
			service: &configv1.UpstreamServiceConfig{
				Name: proto.String("test-invalid"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: proto.String("not-a-url"),
					},
				},
			},
			expectErr:   true,
			errContains: "failed to parse address",
		},
	}

    // Fix Disable field using protojson
    disabledJSON := `{
        "name": "test-disabled",
        "disable": true,
        "http_service": { "address": "http://invalid.address.that.fails:12345" }
    }`
    disabledConfig := &configv1.UpstreamServiceConfig{}
    require.NoError(t, protojson.Unmarshal([]byte(disabledJSON), disabledConfig))
    tests[4].service = disabledConfig


	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateServiceConnection(context.Background(), tt.service)
			if tt.expectErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParseHostPort(t *testing.T) {
	tests := []struct {
		input      string
		expectHost string
		expectPort string
		expectErr  bool
	}{
		{"http://example.com", "example.com", "80", false},
		{"https://example.com", "example.com", "443", false},
		{"http://example.com:8080", "example.com", "8080", false},
		{"localhost:50051", "localhost", "50051", false},
		{"127.0.0.1:9090", "127.0.0.1", "9090", false},
		{"example.com", "", "", true},       // Missing port and scheme
		{"ftp://example.com", "", "", true}, // Unknown scheme default port
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			h, p, err := parseHostPort(tt.input)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectHost, h)
				assert.Equal(t, tt.expectPort, p)
			}
		})
	}
}
