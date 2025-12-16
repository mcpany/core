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

package resource

import (
	"testing"

	"go.uber.org/mock/gomock"
)

func TestMockManagerInterface(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockManagerInterface(ctrl)

	// Exercise the mock methods to ensure coverage of generated code

	// AddResource
	mock.EXPECT().AddResource(gomock.Any()).Times(1)
	mock.AddResource(nil)

	// Clear
	mock.EXPECT().Clear().Times(1)
	mock.Clear()

	// ClearResourcesForService
	mock.EXPECT().ClearResourcesForService("service1").Times(1)
	mock.ClearResourcesForService("service1")

	// GetResource
	mock.EXPECT().GetResource("uri1").Return(nil, false).Times(1)
	mock.GetResource("uri1")

	// ListResources
	mock.EXPECT().ListResources().Return(nil).Times(1)
	mock.ListResources()

	// OnListChanged
	mock.EXPECT().OnListChanged(gomock.Any()).Times(1)
	mock.OnListChanged(nil)

	// RemoveResource
	mock.EXPECT().RemoveResource("uri2").Times(1)
	mock.RemoveResource("uri2")
}
