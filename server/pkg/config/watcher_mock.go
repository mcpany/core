// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0
// Summary: MockWatcher is a mock implementation of the Watcher for testing.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// NewMockWatcher creates a new mock watcher.
//
// Returns:
//   - *MockWatcher: The result.
//
// Side Effects:
//   - None.
//
//
// Errors:
//   - An error if it fails.
// Watch mocks the Watch method.
//
// Parameters:
//   - paths ([]string): The parameter.
//   - reloadFunc (func(): The parameter.
//
// Returns:
//   - ) (error): An error if the operation fails.
//
// Errors:
//   - Returns an error if ...
//
// Side Effects:
//   - None.
// Close mocks the Close method.
//
// Parameters:
//   - None.
//
// Side Effects:
//   - None.
//
//
// Errors:
//   - An error if it fails.
package config

type MockWatcher struct {
	WatchFunc	func(paths []string, reloadFunc func())
	CloseFunc	func()
}

func NewMockWatcher() *MockWatcher {
	return &MockWatcher{}
}

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
