/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tool

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"
)

func TestMockManagerInterface(t *testing.T) {
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

	// Clear
	mock.EXPECT().Clear().Times(1)
	mock.Clear()

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

	// ListTools
	mock.EXPECT().ListTools().Return(nil).Times(1)
	mock.ListTools()

	// SetMCPServer
	mock.EXPECT().SetMCPServer(nil).Times(1)
	mock.SetMCPServer(nil)
}
