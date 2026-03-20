// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

// MockWatcher is a mock implementation of the Watcher for testing.
//
// Summary: MockWatcher is a mock implementation of the Watcher for testing.
type MockWatcher struct {
	WatchFunc func(paths []string, reloadFunc func())
	CloseFunc func()
}

// NewMockWatcher creates a new mock watcher.
//
// Summary: NewMockWatcher creates a new mock watcher.
//
// Parameters:
//   - None.
//
// Returns:
//   - *MockWatcher: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func NewMockWatcher() *MockWatcher {
	return &MockWatcher{}
}

// Watch mocks the Watch method.
//
// Summary: Watch mocks the Watch method.
//
// Parameters:
//   - paths ([]string): The textual representation of paths.
//   - reloadFunc (func()): The provided reloadfunc data.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (m *MockWatcher) Watch(paths []string, reloadFunc func()) error {
	if m.WatchFunc != nil {
		m.WatchFunc(paths, reloadFunc)
	}
	return nil
}

// Close mocks the Close method.
//
// Summary: Close mocks the Close method.
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
//   - May modify internal state or perform external network calls.
func (m *MockWatcher) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}
