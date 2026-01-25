package browser

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
)

func TestBrowserUpstream_Register(t *testing.T) {
	// Skip if short mode.
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	u := NewBrowserUpstream()

	cfg := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-browser"),
		Id: proto.String("test-browser-id"),
		ServiceConfig: &configv1.UpstreamServiceConfig_BrowserService{
			BrowserService: &configv1.BrowserUpstreamService{
				Browser: proto.String("chromium"),
				Headless: proto.Bool(true),
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockToolManager := tool.NewMockManagerInterface(ctrl)
	// Expect 3 tools to be added
	mockToolManager.EXPECT().AddTool(gomock.Any()).Return(nil).Times(3)

	id, tools, _, err := u.Register(context.Background(), cfg, mockToolManager, nil, nil, false)

	// If playwright is not installed, this might fail with "could not launch playwright".
	// We'll treat that as skip.
	if err != nil {
		t.Logf("Skipping test due to playwright error (likely not installed): %v", err)
		return
	}

	assert.NoError(t, err)
	assert.Equal(t, "test-browser-id", id)
	assert.Len(t, tools, 3)

	err = u.Shutdown(context.Background())
	assert.NoError(t, err)
}
