// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

// MockWatcher is a mock implementation of the Watcher for testing.
//
// Summary: Represents a MockWatcher.
type MockWatcher struct {
	WatchFunc func(paths []string, reloadFunc func())
	CloseFunc func()
}

// NewMockWatcher creates a new mock watcher.
//
// Summary: Creates a new mock watcher.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - *MockWatcher: The result.
func NewMockWatcher() *MockWatcher {
	return &MockWatcher{}
}

// Watch watch watch.
//
// Summary: Watch watch.
//
// Parameters: - None.
//   - paths ([]string): The paths.
//   - reloadFunc (func()): The reload func.
//
// Returns: - None.
//   - error: An error if the operation fails.
func (m *MockWatcher) Watch(paths []string, reloadFunc func()) error {
	if m.WatchFunc != nil {
		m.WatchFunc(paths, reloadFunc)
	}
	return nil
}

// Close close close.
//
// Summary: Close close.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - None.
func (m *MockWatcher) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}
