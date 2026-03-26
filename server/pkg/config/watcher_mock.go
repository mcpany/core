// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

// MockWatcher is a mock implementation of the Watcher for testing.
//
// Summary: Mock implementation of the configuration watcher for unit tests.
type MockWatcher struct {
	WatchFunc func(paths []string, reloadFunc func())
	CloseFunc func()
}

// NewMockWatcher creates a new mock watcher.
//
// Summary: Initializes a new MockWatcher instance for testing.
//
// Returns:
//   - *MockWatcher: A pointer to the newly created mock watcher.
//
// Parameters:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewMockWatcher() *MockWatcher {
	return &MockWatcher{}
}

// Watch mocks the Watch method.
//
// Summary: Invokes the mock's WatchFunc if defined.
//
// Parameters:
//   - paths: []string. The file paths to monitor for changes.
//   - reloadFunc: func(). The callback function to execute when a change is detected.
//
// Returns:
//   - error: Nil.
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

// Close mocks the Close method.
//
// Summary: Invokes the mock's CloseFunc if defined.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
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
