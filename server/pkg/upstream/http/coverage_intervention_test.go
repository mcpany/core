package http

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestCoverageIntervention_Shutdown(t *testing.T) {
	tests := []struct {
		name        string
		checker     *mockChecker
		shouldCrash bool
	}{
		{
			name:        "nil checker",
			checker:     nil,
			shouldCrash: false,
		},
		{
			name:        "valid checker",
			checker:     &mockChecker{},
			shouldCrash: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUpstream()
			if tt.checker != nil {
				u.checker = tt.checker
			}
			err := u.Shutdown(context.Background())
			require.NoError(t, err)
			if tt.checker != nil {
				require.True(t, tt.checker.stopped)
			}
		})
	}
}

type mockChecker struct {
	stopped bool
}

func (m *mockChecker) Stop() {
	m.stopped = true
}

func TestCoverageIntervention_RegisterEdgeCases(t *testing.T) {
	invalidSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"invalid_type": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"type": structpb.NewStringValue("invalid"),
				},
			}),
		},
	}

	tests := []struct {
		name          string
		serviceConfig *configv1.UpstreamServiceConfig
		expectedTools []string
	}{
		{
			name: "invalid base url",
			serviceConfig: &configv1.UpstreamServiceConfig{
				ServiceId: "test-service-1",
				Upstream: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpService{
						Address: "://invalid-url",
						Calls: map[string]*configv1.HttpCall{
							"call1": {EndpointPath: "/path"},
						},
					},
				},
			},
			expectedTools: []string{},
		},
		{
			name: "export policy allowlist",
			serviceConfig: &configv1.UpstreamServiceConfig{
				ServiceId: "test-service-2",
				ToolExportPolicy: &configv1.ToolExportPolicy{
					Policy: &configv1.ToolExportPolicy_Allowlist{
						Allowlist: &configv1.ToolExportPolicy_List{
							ToolNames: []string{"allowed_tool"},
						},
					},
				},
				Upstream: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpService{
						Address: "http://localhost",
						Calls: map[string]*configv1.HttpCall{
							"allowed_tool": {EndpointPath: "/allowed"},
							"blocked_tool": {EndpointPath: "/blocked"},
						},
					},
				},
			},
			expectedTools: []string{"allowed_tool"},
		},
		{
			name: "schema and path edge cases",
			serviceConfig: &configv1.UpstreamServiceConfig{
				ServiceId: "test-service-3",
				Upstream: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpService{
						Address: "http://localhost",
						Calls: map[string]*configv1.HttpCall{
							"invalid_schema": {
								EndpointPath: "/path",
								Parameters:   invalidSchema,
							},
							"empty_path": {
								EndpointPath: "//",
							},
						},
					},
				},
			},
			expectedTools: []string{"empty_path"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUpstream()
			toolManager := tool.NewManager()
			resourceManager := resource.NewManager()

			_, tools, _, err := u.Register(context.Background(), tt.serviceConfig, toolManager, nil, resourceManager, false)
			require.NoError(t, err)

			var registeredNames []string
			for _, tool := range tools {
				registeredNames = append(registeredNames, tool.Name)
			}

			require.ElementsMatch(t, tt.expectedTools, registeredNames)
		})
	}
}
