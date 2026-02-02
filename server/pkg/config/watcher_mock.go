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
// Summary: Creates a new instance of MockWatcher for testing purposes.
//
// Parameters:
//   - None.
//
// Returns:
//   - *MockWatcher: A new mock watcher instance.
//
// Errors/Throws:
//   - None.
//
// Side Effects:
//   - None.
func NewMockWatcher() *MockWatcher {
	return &MockWatcher{}
}

// Watch mocks the Watch method.
//
// Summary: Mocks the Watch method, delegating to WatchFunc if set.
//
// Parameters:
//   - paths: []string. The list of paths to watch.
//   - reloadFunc: func(). The function to call when a reload is triggered.
//
// Returns:
//   - error: An error if WatchFunc returns one, otherwise nil.
//
// Errors/Throws:
//   - Returns error from WatchFunc if configured.
//
// Side Effects:
//   - Invokes WatchFunc if not nil.
func (m *MockWatcher) Watch(paths []string, reloadFunc func()) error {
	if m.WatchFunc != nil {
		m.WatchFunc(paths, reloadFunc)
	}
	return nil
}

// Close mocks the Close method.
//
// Summary: Mocks the Close method, delegating to CloseFunc if set.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors/Throws:
//   - None.
//
// Side Effects:
//   - Invokes CloseFunc if not nil.
func (m *MockWatcher) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}
