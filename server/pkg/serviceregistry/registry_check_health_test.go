package serviceregistry

import (
	"context"
	"errors"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
)

type mockHealthUpstream struct {
	err error
}

func (m *mockHealthUpstream) CheckHealth(ctx context.Context) error {
	return m.err
}

func (m *mockHealthUpstream) Shutdown(ctx context.Context) error { return nil }

func (m *mockHealthUpstream) Register(
	ctx context.Context,
	serviceConfig *configv1.UpstreamServiceConfig,
	toolManager tool.ManagerInterface,
	promptManager prompt.ManagerInterface,
	resourceManager resource.ManagerInterface,
	isReload bool,
) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	return "", nil, nil, nil
}

func TestCheckHealth(t *testing.T) {
	r := &ServiceRegistry{
		upstreams: make(map[string]upstream.Upstream),
	}

	name := "test-service"
	id, _ := util.SanitizeServiceName(name)

	// Case 1: Service not found
	err := r.CheckHealth(context.Background(), name)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Case 2: Healthy
	r.upstreams[id] = &mockHealthUpstream{err: nil}
	err = r.CheckHealth(context.Background(), name)
	assert.NoError(t, err)

	// Case 3: Unhealthy
	r.upstreams[id] = &mockHealthUpstream{err: errors.New("connection refused")}
	err = r.CheckHealth(context.Background(), name)
	assert.Error(t, err)
	assert.Equal(t, "connection refused", err.Error())
}
