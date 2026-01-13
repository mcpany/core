// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

func TestVerifyServices(t *testing.T) {
	// Enable loopback for testing
	os.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	defer os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")

	// Start a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tests := []struct {
		name           string
		config         *configv1.McpAnyServerConfig
		expectedErrors map[string]bool // map[ServiceName]shouldHaveError
	}{
		{
			name: "HttpService Reachable",
			config: &configv1.McpAnyServerConfig{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{
						Name: proto.String("reachable"),
						ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
							HttpService: &configv1.HttpUpstreamService{
								Address: proto.String(server.URL),
							},
						},
					},
				},
			},
			expectedErrors: map[string]bool{"reachable": false},
		},
		{
			name: "HttpService Unreachable",
			config: &configv1.McpAnyServerConfig{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{
						Name: proto.String("unreachable"),
						ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
							HttpService: &configv1.HttpUpstreamService{
								Address: proto.String("http://localhost:54321"), // Assume this port is closed
							},
						},
					},
				},
			},
			expectedErrors: map[string]bool{"unreachable": true},
		},
		{
			name: "Disabled Service Skipped",
			config: &configv1.McpAnyServerConfig{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{
						Name:    proto.String("disabled"),
						Disable: proto.Bool(true),
						ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
							HttpService: &configv1.HttpUpstreamService{
								Address: proto.String("http://localhost:54321"),
							},
						},
					},
				},
			},
			expectedErrors: map[string]bool{"disabled": false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			VerifyServices(context.Background(), tt.config)
			for _, svc := range tt.config.UpstreamServices {
				shouldErr := tt.expectedErrors[svc.GetName()]
				if shouldErr && svc.GetConfigError() == "" {
					t.Errorf("Service %s expected error but got none", svc.GetName())
				}
				if !shouldErr && svc.GetConfigError() != "" {
					t.Errorf("Service %s expected no error but got: %s", svc.GetName(), svc.GetConfigError())
				}

				// Verify Disable is NOT set by VerifyServices
				if shouldErr && svc.GetDisable() {
                    if svc.GetName() == "unreachable" {
                        t.Errorf("Service %s should not be disabled automatically", svc.GetName())
                    }
                }
			}
		})
	}
}

func TestVerifyServices_Grpc(t *testing.T) {
	// Enable loopback for testing connectivity verification (uses SafeDialer)
	os.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	defer os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")

	// Start a dummy TCP listener to simulate gRPC port open
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer l.Close()
	addr := l.Addr().String()

	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("grpc-reachable"),
				ServiceConfig: &configv1.UpstreamServiceConfig_GrpcService{
					GrpcService: &configv1.GrpcUpstreamService{
						Address: proto.String(addr),
					},
				},
			},
			{
				Name: proto.String("grpc-unreachable"),
				ServiceConfig: &configv1.UpstreamServiceConfig_GrpcService{
					GrpcService: &configv1.GrpcUpstreamService{
						Address: proto.String("127.0.0.1:54322"),
					},
				},
			},
		},
	}

	VerifyServices(context.Background(), config)

	for _, svc := range config.UpstreamServices {
		if svc.GetName() == "grpc-reachable" && svc.GetConfigError() != "" {
			t.Errorf("grpc-reachable got error: %s", svc.GetConfigError())
		}
		if svc.GetName() == "grpc-unreachable" && svc.GetConfigError() == "" {
			t.Errorf("grpc-unreachable expected error")
		}
	}
}
