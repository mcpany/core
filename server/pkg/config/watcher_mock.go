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
// Parameters:
//   None.
//
// Returns:
//   - *MockWatcher: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewMockWatcher() *MockWatcher {
	return &MockWatcher{}
}

// Watch watch watch.
//
// Summary: Watch watch.
//
// Parameters:
//   - paths ([]string): The paths.
//   - reloadFunc (func()): The reload func.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
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
// Parameters:
//   None.
//
// Returns:
//   None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (m *MockWatcher) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}
