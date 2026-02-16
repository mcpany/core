// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

// MockWatcher is a mock implementation of the Watcher for testing.
type MockWatcher struct {
	WatchFunc func(paths []string, reloadFunc func())
	CloseFunc func()
}

// NewMockWatcher creates a new mock watcher.
//
// Summary: Creates a new mock watcher.
//
// Returns:
//   - *MockWatcher: A new mock watcher instance.
func NewMockWatcher() *MockWatcher {
	return &MockWatcher{}
}

// Watch mocks the Watch method.
//
// Summary: Mocks the file watching functionality.
//
// Parameters:
//   - paths: []string. The paths to watch.
//   - reloadFunc: func(). The function to call on reload.
//
// Returns:
//   - error: An error if the watch fails.
func (m *MockWatcher) Watch(paths []string, reloadFunc func()) error {
	if m.WatchFunc != nil {
		m.WatchFunc(paths, reloadFunc)
	}
	return nil
}

// Close mocks the Close method.
//
// Summary: Mocks the close functionality.
//
// Returns:
//   None.
func (m *MockWatcher) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}
