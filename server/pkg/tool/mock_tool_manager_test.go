// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"
)

func TestMockManagerInterface(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockManagerInterface(ctrl)

	// Exercise generated methods

	// AddMiddleware
	mock.EXPECT().AddMiddleware(gomock.Any()).Times(1)
	mock.AddMiddleware(nil)

	// AddServiceInfo
	mock.EXPECT().AddServiceInfo("id", nil).Times(1)
	mock.AddServiceInfo("id", nil)

	// AddTool
	mock.EXPECT().AddTool(gomock.Any()).Return(nil).Times(1)
	_ = mock.AddTool(nil)

	// ClearToolsForService
	mock.EXPECT().ClearToolsForService("id").Times(1)
	mock.ClearToolsForService("id")

	// ExecuteTool
	mock.EXPECT().ExecuteTool(gomock.Any(), gomock.Any()).Return("result", nil).Times(1)
	_, _ = mock.ExecuteTool(context.Background(), nil)

	// GetServiceInfo
	mock.EXPECT().GetServiceInfo("id").Return(nil, false).Times(1)
	mock.GetServiceInfo("id")

	// GetTool
	mock.EXPECT().GetTool("name").Return(nil, false).Times(1)
	mock.GetTool("name")

	// ListServices
	mock.EXPECT().ListServices().Return(nil).Times(1)
	mock.ListServices()

	// ListTools
	mock.EXPECT().ListTools().Return(nil).Times(1)
	mock.ListTools()

	// SetMCPServer
	mock.EXPECT().SetMCPServer(nil).Times(1)
	mock.SetMCPServer(nil)
}
