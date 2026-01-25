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

func TestNewUpstream(t *testing.T) {
	u := NewUpstream()
	assert.NotNil(t, u)
}

func TestRegister_NoBrowserConfig(t *testing.T) {
	u := NewUpstream()

	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-browser"),
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockManager := tool.NewMockManagerInterface(ctrl)

	_, _, _, err := u.Register(context.Background(), config, mockManager, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "browser service config is nil")
}

func TestRegister_BrowserConfig(t *testing.T) {
	u := NewUpstream().(*Upstream)
	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-browser"),
		ServiceConfig: &configv1.UpstreamServiceConfig_BrowserService{
			BrowserService: &configv1.BrowserUpstreamService{
				BrowserType:    proto.String("chromium"),
				Headless:       proto.Bool(true),
				UserAgent:      proto.String("test-agent"),
				ViewportWidth:  proto.Int32(800),
				ViewportHeight: proto.Int32(600),
				ScreenshotDir:  proto.String("/tmp/screenshots"),
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockManager := tool.NewMockManagerInterface(ctrl)

	mockManager.EXPECT().AddServiceInfo(gomock.Any(), gomock.Any()).AnyTimes()
	mockManager.EXPECT().AddTool(gomock.Any()).AnyTimes()

	_, _, _, err := u.Register(context.Background(), config, mockManager, nil, nil, false)

	// Verify config is loaded correctly into the struct
	assert.Equal(t, "chromium", u.browserType)
	assert.Equal(t, true, u.headless)
	assert.Equal(t, "test-agent", u.userAgent)
	assert.Equal(t, int32(800), u.viewportWidth)
	assert.Equal(t, int32(600), u.viewportHeight)
	assert.Equal(t, "/tmp/screenshots", u.screenshotDir)

	if err != nil {
		t.Logf("Register failed (expected in environments without playwright): %v", err)
	} else {
		t.Log("Register succeeded")
		_ = u.Shutdown(context.Background())
	}
}
