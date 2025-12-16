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

package prompt

import (
	"testing"

	"go.uber.org/mock/gomock"
)

func TestMockManagerInterface(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockManagerInterface(ctrl)

	// AddPrompt
	mock.EXPECT().AddPrompt(gomock.Any()).Times(1)
	mock.AddPrompt(nil)

	// Clear
	mock.EXPECT().Clear().Times(1)
	mock.Clear()

	// ClearPromptsForService
	mock.EXPECT().ClearPromptsForService("id").Times(1)
	mock.ClearPromptsForService("id")

	// GetPrompt
	mock.EXPECT().GetPrompt("name").Return(nil, false).Times(1)
	mock.GetPrompt("name")

	// ListPrompts
	mock.EXPECT().ListPrompts().Return(nil).Times(1)
	mock.ListPrompts()

	// SetMCPServer
	mock.EXPECT().SetMCPServer(nil).Times(1)
	mock.SetMCPServer(nil)

	// UpdatePrompt
	mock.EXPECT().UpdatePrompt(gomock.Any()).Times(1)
	mock.UpdatePrompt(nil)
}
