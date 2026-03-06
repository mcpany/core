// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

// MockWatcher - Auto-generated documentation.
//
// Summary: MockWatcher is a mock implementation of the Watcher for testing.
//
// Fields:
//   - Various fields for MockWatcher.
type MockWatcher struct {
	WatchFunc func(paths []string, reloadFunc func())
	CloseFunc func()
}

// NewMockWatcher - Auto-generated documentation.
//
// Summary: NewMockWatcher creates a new mock watcher.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func NewMockWatcher() *MockWatcher {
	return &MockWatcher{}
}

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
func (m *MockWatcher) Watch(paths []string, reloadFunc func()) error {
	if m.WatchFunc != nil {
		m.WatchFunc(paths, reloadFunc)
	}
	return nil
}

// Close - Auto-generated documentation.
//
// Summary: Close mocks the Close method.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func (m *MockWatcher) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}
