// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package diagnostics

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

func TestDiagnoseService_EmptyConfig(t *testing.T) {
	config := &configv1.UpstreamServiceConfig{}
	report := DiagnoseService(context.Background(), config)

	if report.ServiceName != "" {
		t.Errorf("Expected empty service name, got %s", report.ServiceName)
	}
	if len(report.Steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(report.Steps))
	}
	if report.Steps[0].Status != "failure" {
		t.Errorf("Expected failure status, got %s", report.Steps[0].Status)
	}
}

func TestDiagnoseService_HTTP(t *testing.T) {
	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-http"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String("https://example.com"),
			},
		},
	}
	report := DiagnoseService(context.Background(), config)

	if report.ServiceName != "test-http" {
		t.Errorf("Expected service name test-http, got %s", report.ServiceName)
	}
	// 1 config check + 1 dns + 1 handshake = 3 steps
	// Note: In unit tests without network mocking, real network calls might fail.
	// But we are asserting number of steps generated, not their status.
	// Actually, diagnoseHTTP does execute the steps and append them.
	// If it fails, it still appends the step with failure status.
	// So checking length is safe.
	if len(report.Steps) != 3 {
		t.Errorf("Expected 3 steps, got %d", len(report.Steps))
	}
}

func TestDiagnoseService_Command(t *testing.T) {
	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-cmd"),
		ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
			CommandLineService: &configv1.CommandLineUpstreamService{
				Command: proto.String("ls"),
			},
		},
	}
	report := DiagnoseService(context.Background(), config)

	if report.ServiceName != "test-cmd" {
		t.Errorf("Expected service name test-cmd, got %s", report.ServiceName)
	}
	// 1 config check + 1 command check = 2 steps
	if len(report.Steps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(report.Steps))
	}
	// Check if command check passed (ls should exist)
	if report.Steps[1].Status != "success" {
		t.Errorf("Expected command check success, got %s", report.Steps[1].Status)
	}
}

func TestDiagnoseService_MCP_Stdio(t *testing.T) {
	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-mcp-stdio"),
		ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
			McpService: &configv1.McpUpstreamService{
				ConnectionType: &configv1.McpUpstreamService_StdioConnection{
					StdioConnection: &configv1.McpStdioConnection{
						Command: proto.String("echo"),
					},
				},
			},
		},
	}
	report := DiagnoseService(context.Background(), config)

	if report.ServiceName != "test-mcp-stdio" {
		t.Errorf("Expected service name test-mcp-stdio, got %s", report.ServiceName)
	}
	// 1 config check + 1 command check = 2 steps
	if len(report.Steps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(report.Steps))
	}
}

func TestDiagnoseService_MCP_HTTP(t *testing.T) {
	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-mcp-http"),
		ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
			McpService: &configv1.McpUpstreamService{
				ConnectionType: &configv1.McpUpstreamService_HttpConnection{
					HttpConnection: &configv1.McpStreamableHttpConnection{
						HttpAddress: proto.String("http://example.com"),
					},
				},
			},
		},
	}
	report := DiagnoseService(context.Background(), config)

	if report.ServiceName != "test-mcp-http" {
		t.Errorf("Expected service name test-mcp-http, got %s", report.ServiceName)
	}
	// 1 config check + 1 dns + 1 tcp = 3 steps
	if len(report.Steps) != 3 {
		t.Errorf("Expected 3 steps, got %d", len(report.Steps))
	}
}
