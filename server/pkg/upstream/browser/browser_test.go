// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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

func TestRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	u := NewBrowserUpstream()
	tm := tool.NewMockManagerInterface(ctrl)

	// Expect 5 tools to be added
	tm.EXPECT().AddTool(gomock.Any()).Return(nil).Times(5)

	serviceConfig := &configv1.UpstreamServiceConfig{
		Id: proto.String("browser-service"),
	}

	id, tools, _, err := u.Register(context.Background(), serviceConfig, tm, nil, nil, false)

	assert.NoError(t, err)
	assert.Equal(t, "browser-service", id)
	assert.Len(t, tools, 5) // open, screenshot, click, type, content
}
