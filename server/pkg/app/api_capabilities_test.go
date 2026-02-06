package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	mcp_router_v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// Define Mocks for Capabilities Test

type CapTestMockResource struct {
	uri     string
	service string
}

func (m *CapTestMockResource) Resource() *mcp.Resource {
	return &mcp.Resource{URI: m.uri}
}
func (m *CapTestMockResource) Service() string { return m.service }
func (m *CapTestMockResource) Read(ctx context.Context) (*mcp.ReadResourceResult, error) {
	return nil, nil
}
func (m *CapTestMockResource) Subscribe(ctx context.Context) error { return nil }

type CapTestMockPrompt struct {
	name    string
	service string
}

func (m *CapTestMockPrompt) Prompt() *mcp.Prompt {
	return &mcp.Prompt{Name: m.name}
}
func (m *CapTestMockPrompt) Service() string { return m.service }
func (m *CapTestMockPrompt) Get(ctx context.Context, args json.RawMessage) (*mcp.GetPromptResult, error) {
	return nil, nil
}

func TestHandleServices_CapabilitiesCount(t *testing.T) {
	busProvider, _ := bus.NewProvider(nil)
	tm := tool.NewManager(busProvider)
	rm := resource.NewManager()
	pm := prompt.NewManager()

	// Register Tools
	tm.AddTool(&TestMockTool{toolDef: mcp_router_v1.Tool_builder{Name: proto.String("tool1"), ServiceId: proto.String("test-service")}.Build()})
	tm.AddTool(&TestMockTool{toolDef: mcp_router_v1.Tool_builder{Name: proto.String("tool2"), ServiceId: proto.String("test-service")}.Build()})

	// Register Resources
	rm.AddResource(&CapTestMockResource{uri: "res://1", service: "test-service"})
	rm.AddResource(&CapTestMockResource{uri: "res://2", service: "other-service"})

	// Register Prompts
	pm.AddPrompt(&CapTestMockPrompt{name: "prompt1", service: "test-service"})
	pm.AddPrompt(&CapTestMockPrompt{name: "prompt2", service: "test-service"})
	pm.AddPrompt(&CapTestMockPrompt{name: "prompt3", service: "test-service"})

	app := NewApplication()
	app.ToolManager = tm
	app.ResourceManager = rm
	app.PromptManager = pm

	app.ServiceRegistry = &TestMockServiceRegistry{
		services: func() []*configv1.UpstreamServiceConfig {
			s1 := configv1.UpstreamServiceConfig_builder{
				Name: proto.String("test-service"),
				Id:   proto.String("test-service"),
			}.Build()
			s2 := configv1.UpstreamServiceConfig_builder{
				Name: proto.String("other-service"),
				Id:   proto.String("other-service"),
			}.Build()
			return []*configv1.UpstreamServiceConfig{s1, s2}
		}(),
	}

	req := httptest.NewRequest(http.MethodGet, "/services", nil)
	rr := httptest.NewRecorder()

	handler := app.handleServices(&MockServiceStore{})
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var response []map[string]any
	json.Unmarshal(rr.Body.Bytes(), &response)
	require.Len(t, response, 2)

	for _, s := range response {
		if s["name"] == "test-service" {
			assert.Equal(t, float64(2), s["tool_count"], "Tool count mismatch")
			assert.Equal(t, float64(1), s["resource_count"], "Resource count mismatch")
			assert.Equal(t, float64(3), s["prompt_count"], "Prompt count mismatch")
		}
		if s["name"] == "other-service" {
			assert.Equal(t, float64(0), s["tool_count"])
			assert.Equal(t, float64(1), s["resource_count"])
			assert.Equal(t, float64(0), s["prompt_count"])
		}
	}
}
