// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

// MockWatcher mockWatcher represents a mock watcher.
//
// Summary: MockWatcher represents a mock watcher.
type MockWatcher struct {
	WatchFunc func(paths []string, reloadFunc func())
	CloseFunc func()
}

// NewMockWatcher creates a new mock watcher.
//
// Returns: - None.
//   - *MockWatcher: The result.
//
// Side Effects: - None.
//   - None.
//
// Summary: Initializes NewMockWatcher operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func NewMockWatcher() *MockWatcher {
	return &MockWatcher{}
}

// Watch mocks the Watch method.
//
// Parameters: - None.
//   - paths ([]string): The parameter.
//   - reloadFunc (func(): The parameter.
//
// Returns: - None.
//   - ) (error): An error if the operation fails.
//
// Errors: - None.
//   - Returns an error if ...
//
// Side Effects: - None.
//   - None.
//
// Summary: Executes Watch operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (m *MockWatcher) Watch(paths []string, reloadFunc func()) error {
	if m.WatchFunc != nil {
		m.WatchFunc(paths, reloadFunc)
	}
	return nil
}

// Close mocks the Close method.
//
// Parameters: - None.
//   - None.
//
// Side Effects: - None.
//   - None.
//
// Summary: Executes Close operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (m *MockWatcher) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}
