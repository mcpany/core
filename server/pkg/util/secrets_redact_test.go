// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

func TestRedactSlice_Coverage(t *testing.T) {
	// redactSlice is unexported, but we are in package util.
	// It's technically unused by the package logic (replaced by redactSliceMaybe),
	// but kept for legacy/compatibility?
	// To cover it, we call it directly.

	// Case 1: Redaction needed (early return path)
	input := []interface{}{
		"normal",
		map[string]interface{}{"secret": "value", "public": "ok"},
		[]interface{}{"nested", map[string]interface{}{"password": "123"}},
	}

	redacted := redactSlice(input)

	if len(redacted) != 3 {
		t.Errorf("expected 3 elements")
	}

	// Check element 1 (map)
	m := redacted[1].(map[string]interface{})
	if m["secret"] != "[REDACTED]" {
		t.Errorf("map secret not redacted")
	}
	if m["public"] != "ok" {
		t.Errorf("map public modified")
	}

	// Check element 2 (slice)
	s := redacted[2].([]interface{})
	m2 := s[1].(map[string]interface{})
	if m2["password"] != "[REDACTED]" {
		t.Errorf("nested map password not redacted")
	}

	// Case 2: No redaction needed (deep copy path)
	inputClean := []interface{}{
		"safe",
		map[string]interface{}{"public": "val"},
		[]interface{}{"safe_nested"},
	}
	redactedClean := redactSlice(inputClean)
	if len(redactedClean) != 3 {
		t.Errorf("clean expected 3 elements")
	}
	// Verify content
	if redactedClean[0] != "safe" {
		t.Errorf("element 0 mismatch")
	}
}

func TestStripSecrets_Coverage(t *testing.T) {
	// Test the empty stripping functions to ensure coverage

	// 1. GrpcService
	grpcSvc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_GrpcService{
			GrpcService: &configv1.GrpcUpstreamService{},
		},
	}
	StripSecretsFromService(grpcSvc) // Should call stripSecretsFromGrpcService

	// 2. OpenapiService
	openapiSvc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_OpenapiService{
			OpenapiService: &configv1.OpenapiUpstreamService{},
		},
	}
	StripSecretsFromService(openapiSvc) // Should call stripSecretsFromOpenapiService

	// 3. McpService calls stripSecretsFromMcpCall
	mcpSvc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
			McpService: &configv1.McpUpstreamService{
				Calls: map[string]*configv1.MCPCallDefinition{
					"test": {},
				},
			},
		},
	}
	StripSecretsFromService(mcpSvc) // Should call stripSecretsFromMcpCall
}
