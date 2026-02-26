// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

// MockWatcher is a mock implementation of the Watcher for testing.
//
// Summary: is a mock implementation of the Watcher for testing.
type MockWatcher struct {
	WatchFunc func(paths []string, reloadFunc func())
	CloseFunc func()
}

// NewMockWatcher creates a new mock watcher.
//
//
// Summary: creates a new mock watcher.
//
// Returns:
// - *MockWatcher: The result.
//
// Side Effects:
// - None.
func NewMockWatcher() *MockWatcher {
	return &MockWatcher{}
}

// Watch mocks the Watch method.
//
//
// Summary: mocks the Watch method.
//
// Parameters:
// - paths ([]string): The parameter.
// - reloadFunc (func(): The parameter.
//
// Returns:
// - ) (error): An error if the operation fails.
//
// Errors:
// - Returns an error if ...
//
// Side Effects:
// - None.
func (m *MockWatcher) Watch(paths []string, reloadFunc func()) error {
	if m.WatchFunc != nil {
		m.WatchFunc(paths, reloadFunc)
	}
	return nil
}

// Close mocks the Close method.
//
//
// Summary: mocks the Close method.
//
// Parameters:
// - None.
//
// Side Effects:
// - None.
func (m *MockWatcher) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}
