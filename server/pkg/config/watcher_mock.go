// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

// MockWatcher is a mock implementation of the Watcher for testing.
type MockWatcher struct {
	WatchFunc func(paths []string, reloadFunc func())
	CloseFunc func()
}

func NewMockWatcher() *MockWatcher {
	return &MockWatcher{}
}

// Watch mocks the Watch method.
//
// paths is the paths.
// reloadFunc is the reloadFunc.
//
// Returns an error if the operation fails.
func (m *MockWatcher) Watch(paths []string, reloadFunc func()) error {
	if m.WatchFunc != nil {
		m.WatchFunc(paths, reloadFunc)
	}
	return nil
}

func (m *MockWatcher) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}
