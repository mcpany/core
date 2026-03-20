// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0
// Summary: MockWatcher is a mock implementation of the Watcher for testing.
//
// Side Effects:
//   - None.
//
// Summary: NewMockWatcher creates a new mock watcher.
//
// Returns:
//   - *MockWatcher: The result.
//
// Side Effects:
//   - None.
//
// Parameters:
//   - None.
//
// Errors:
//   - None.
//
// Summary: Watch mocks the Watch method.
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
//
// Summary: Close mocks the Close method.
//
// Parameters:
//   - None.
//
// Side Effects:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
package config

type MockWatcher struct {
	WatchFunc func(paths []string, reloadFunc func())
	CloseFunc func()
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
