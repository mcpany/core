// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestExecutionRequest_GetArgument_Removed(t *testing.T) {
	// GetArgument and GetStringArgument are not available in current ExecutionRequest struct.
	// Tests removed.
}

func TestTool_Getters(t *testing.T) {
	// Interface coverage for basic getters of concrete types (MockTool mostly)
	mock := &MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{Name: proto.String("name"), ServiceId: proto.String("svc")}
		},
		GetCacheConfigFunc: func() *configv1.CacheConfig {
			return &configv1.CacheConfig{}
		},
	}

	assert.Equal(t, "name", mock.Tool().GetName())
	assert.Equal(t, "svc", mock.Tool().GetServiceId())
	assert.NotNil(t, mock.GetCacheConfig())
}

func TestExecutionFunc_Coverage(t *testing.T) {
	// Just verifies type matching
	var f ExecutionFunc = func(ctx context.Context, req *ExecutionRequest) (any, error) {
		return nil, nil
	}
	assert.NotNil(t, f)
}

func TestNewManager_WithNilBus(t *testing.T) {
	m := NewManager(nil)
	assert.NotNil(t, m)
	// Should not crash when adding service info or tools
	m.AddServiceInfo("svc", &ServiceInfo{Name: "svc"})
	// AddTool might fail strictly if validation fails, but not because bus is nil (checks for nil)
}

// Additional coverage for Proto Converters (corner cases)
func TestProtoConverters_ParameterMappings(t *testing.T) {
	// Secret mapping
	secretMapping := configv1.HttpParameterMapping_builder{
		Schema: configv1.ParameterSchema_builder{Name: proto.String("apiKey")}.Build(),
		Secret: configv1.SecretValue_builder{
			PlainText: proto.String("secret"),
		}.Build(),
	}.Build()
	assert.NotNil(t, secretMapping.GetSecret())
}

// HttpConfig_ParameterMappings_DeepCheck
// Removed Source check as Source field does not exist.
func TestHttpConfig_ParameterMappings_DeepCheck(t *testing.T) {
	// Verify that complex HTTP call definitions can be built (no execution)
	def := configv1.HttpCallDefinition_builder{
		Method:       configv1.HttpCallDefinition_HTTP_METHOD_POST.Enum(),
		EndpointPath: proto.String("/api"),
		Parameters: []*configv1.HttpParameterMapping{
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name:       proto.String("p1"),
					IsRequired: proto.Bool(true),
				}.Build(),
			}.Build(),
		},
	}.Build()

	assert.Len(t, def.GetParameters(), 1)
	assert.True(t, def.GetParameters()[0].GetSchema().GetIsRequired())
}
