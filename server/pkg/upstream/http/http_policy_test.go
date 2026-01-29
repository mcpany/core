package http

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestHTTPUpstream_InvalidCallPolicy(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	// Invalid regex in call_policies
	configJSON := `{
		"name": "policy-test-service",
		"http_service": {
			"address": "http://example.com",
			"tools": [{"name": "test-op", "call_id": "test-op-call"}],
			"calls": {
				"test-op-call": {
					"id": "test-op-call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/test"
				}
			}
		},
		"call_policies": [
			{
				"rules": [
					{
						"name_regex": "(",
						"action": "ALLOW"
					}
				]
			}
		]
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	// Register should verify call policies and if they fail, it might log error but what does it return?
	// The implementation of Register calls createAndRegisterHTTPTools.
	// createAndRegisterHTTPTools returns nil if policy compilation fails.
	// Register returns the list of tools.

	serviceID, tools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	assert.NoError(t, err) // It doesn't return error, just logs it and returns nil tools.

	// tools should be nil or empty
	assert.Empty(t, tools)

	// And no tools should be registered
	// We can't check all tools easily without ListTools which might not exist or work as expected in this context,
	// but we can check the specific tool we expected.
	// Also tools return value is the list of discovered tools.

	sanitizedToolName, _ := util.SanitizeToolName("test-op")
	toolID := serviceID + "." + sanitizedToolName
	_, ok := tm.GetTool(toolID)
	assert.False(t, ok, "Tool should not be registered if policy compilation fails")
}
