// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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
