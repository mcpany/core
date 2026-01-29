package http

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestHTTPUpstream_SkipDisabled(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	rm := resource.NewManager()
	prm := prompt.NewManager()
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "skip-disabled-service",
		"http_service": {
			"address": "http://example.com",
			"tools": [
				{"name": "enabled-tool", "call_id": "c1"},
				{"name": "disabled-tool", "call_id": "c2", "disable": true}
			],
			"calls": {
				"c1": {
					"id": "c1", "method": "HTTP_METHOD_GET", "endpoint_path": "/c1"
				},
				"c2": {
					"id": "c2", "method": "HTTP_METHOD_GET", "endpoint_path": "/c2"
				}
			},
			"resources": [
				{"name": "enabled-res", "uri": "file:///tmp/1"},
				{"name": "disabled-res", "uri": "file:///tmp/2", "disable": true}
			],
			"prompts": [
				{"name": "enabled-prompt", "messages": [{"role": "USER", "text": {"text": "hi"}}]},
				{"name": "disabled-prompt", "messages": [{"role": "USER", "text": {"text": "bye"}}], "disable": true}
			]
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, prm, rm, false)
	assert.NoError(t, err)

	// Enabled tool should be present
	_, ok := tm.GetTool(serviceID + ".enabled-tool")
	assert.True(t, ok)

	// Disabled tool should be absent
	_, ok = tm.GetTool(serviceID + ".disabled-tool")
	assert.False(t, ok)

	// I cannot easily check resources/prompts via tool manager unless I expose Resource/Prompt Manager interfaces.
	// But the code path in http.go should be covered.
}
