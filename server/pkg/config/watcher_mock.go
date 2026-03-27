// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

// MockWatcher is a mock implementation of the Watcher for testing.
//
// Summary. Represents a MockWatcher.
type MockWatcher struct {
	WatchFunc func(paths []string, reloadFunc func())
	CloseFunc func()
}

// NewMockWatcher provides newmockwatcher functionality.
//
// Summary: NewMockWatcher.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func NewMockWatcher() *MockWatcher {
	return &MockWatcher{}
}

// Watch provides watch functionality.
//
// Summary: Watch.
//
// Parameters.
//   - paths: The parameter.
//   - reloadFunc: The parameter.
//
// Returns.
//   - result: The result.
func (m *MockWatcher) Watch(paths []string, reloadFunc func()) error {
	if m.WatchFunc != nil {
		m.WatchFunc(paths, reloadFunc)
	}
	return nil
}

// Close provides close functionality.
//
// Summary: Close.
//
// Parameters.
//   - None.
//
// Returns.
//   - None.
func (m *MockWatcher) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}
