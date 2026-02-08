package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	configv1 "github.com/mcpany/core/proto/config/v1"
	mcp_router_v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type TaggedMockTool struct {
	toolDef *mcp_router_v1.Tool
}

func (m *TaggedMockTool) Tool() *mcp_router_v1.Tool { return m.toolDef }
func (m *TaggedMockTool) MCPTool() *mcp.Tool {
	return &mcp.Tool{
		Name: m.toolDef.GetName(),
	}
}
func (m *TaggedMockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	return nil, nil
}
func (m *TaggedMockTool) GetCacheConfig() *configv1.CacheConfig { return nil }

func TestHandleTools_IncludesTags(t *testing.T) {
	busProvider, _ := bus.NewProvider(nil)
	tm := tool.NewManager(busProvider)

	toolDef := mcp_router_v1.Tool_builder{
		Name:      proto.String("test-tool"),
		ServiceId: proto.String("test-service"),
		Tags:      []string{"tag1", "tag2"},
	}.Build()

	tm.AddTool(&TaggedMockTool{toolDef: toolDef})

	app := NewApplication()
	app.ToolManager = tm

	req := httptest.NewRequest(http.MethodGet, "/tools", nil)
	w := httptest.NewRecorder()

	handler := app.handleTools()
	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response []map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	require.Len(t, response, 1)

	toolResp := response[0]
	assert.Equal(t, "test-tool", toolResp["name"])
	assert.Equal(t, "test-service", toolResp["serviceId"])

	tags, ok := toolResp["tags"].([]interface{})
	require.True(t, ok, "tags should be present and an array")
	assert.Contains(t, tags, "tag1")
	assert.Contains(t, tags, "tag2")
}
