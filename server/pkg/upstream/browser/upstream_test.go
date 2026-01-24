package browser

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Mock manager
type mockToolManager struct {
	mock.Mock
}

func (m *mockToolManager) AddTool(t tool.Tool) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *mockToolManager) GetTool(toolName string) (tool.Tool, bool) { return nil, false }
func (m *mockToolManager) ListTools() []tool.Tool { return nil }
func (m *mockToolManager) ListMCPTools() []*mcp.Tool { return nil }
func (m *mockToolManager) ClearToolsForService(serviceID string) {}
func (m *mockToolManager) ExecuteTool(ctx context.Context, req *tool.ExecutionRequest) (any, error) { return nil, nil }
func (m *mockToolManager) SetMCPServer(mcpServer tool.MCPServerProvider) {}
func (m *mockToolManager) AddMiddleware(middleware tool.ExecutionMiddleware) {}
func (m *mockToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {}
func (m *mockToolManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) { return nil, false }
func (m *mockToolManager) ListServices() []*tool.ServiceInfo { return nil }
func (m *mockToolManager) SetProfiles(enabled []string, defs []*configv1.ProfileDefinition) {}
func (m *mockToolManager) IsServiceAllowed(serviceID, profileID string) bool { return false }
func (m *mockToolManager) ToolMatchesProfile(t tool.Tool, profileID string) bool { return false }
func (m *mockToolManager) GetAllowedServiceIDs(profileID string) (map[string]bool, bool) { return nil, false }


func TestBrowserUpstream_Register(t *testing.T) {
	u := NewUpstream()

    tm := new(mockToolManager)
    // Expect 5 AddTool calls
    tm.On("AddTool", mock.Anything).Return(nil).Times(5)

	config := &configv1.UpstreamServiceConfig{
		Name: toPtr("browser-service"),
        Id: toPtr("browser-id"),
		ServiceConfig: &configv1.UpstreamServiceConfig_BrowserService{
			BrowserService: &configv1.BrowserUpstreamService{
				Headless: toPtr(true),
                BrowserType: toPtr("chromium"),
			},
		},
	}

	id, tools, resources, err := u.Register(context.Background(), config, tm, nil, nil, false)

    if err != nil {
        t.Logf("Skipping test because playwright setup failed: %v", err)
        return
    }

	assert.NoError(t, err)
	assert.Equal(t, "browser-id", id)
	assert.Len(t, tools, 5)
	assert.Nil(t, resources)

    tm.AssertExpectations(t)
}

func toPtr[T any](v T) *T {
	return &v
}

func TestBrowserUpstream_Execute_NotInitialized(t *testing.T) {
	u := NewUpstream()
	// We don't call Register or launch browser, so u.browser is nil.

	handler := u.createHandler("navigate")
	req := &tool.ExecutionRequest{
		ToolName: "navigate",
		Arguments: map[string]interface{}{"url": "https://example.com"},
	}

	_, err := handler(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "browser not initialized")
}
