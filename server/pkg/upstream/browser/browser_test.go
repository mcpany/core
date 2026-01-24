package browser

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestBrowserUpstream_Register_NilConfig(t *testing.T) {
	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test_browser_nil"),
		// No ServiceConfig set for BrowserService
	}

	u := NewUpstream()
	b, _ := bus.NewProvider(nil)
	tm := tool.NewManager(b)

	_, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "browser service config is nil")
}
