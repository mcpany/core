package upstream

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

type MockUpstream struct{}

func (m *MockUpstream) Shutdown(ctx context.Context) error { return nil }

func (m *MockUpstream) Register(ctx context.Context, config *configv1.UpstreamServiceConfig, tm tool.ManagerInterface, pm prompt.ManagerInterface, rm resource.ManagerInterface, isReload bool) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	return "", nil, nil, nil
}
func (m *MockUpstream) Close() error { return nil }

func TestUpstreamInterface(t *testing.T) {
	// Verify MockUpstream implements Upstream
	var _ Upstream = &MockUpstream{}
	assert.True(t, true)
}
