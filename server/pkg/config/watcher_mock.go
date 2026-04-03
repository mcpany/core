// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

// MockWatcher represents the public MockWatcher entity.
//
// Summary: Defines the structured data model representing a watcher.
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
type MockWatcher struct {
	WatchFunc func(paths []string, reloadFunc func())
	CloseFunc func()
}

// NewMockWatcher serves as a public interface for interacting with NewMockWatcher.
//
// Summary: Constructs and returns an initialized mock watcher ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewMockWatcher() *MockWatcher {
	return &MockWatcher{}
}

// Watch serves as a public interface for interacting with Watch.
//
// Summary: Watch the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *MockWatcher) Watch(paths []string, reloadFunc func()) error {
	if m.WatchFunc != nil {
		m.WatchFunc(paths, reloadFunc)
	}
	return nil
}

// Close serves as a public interface for interacting with Close.
//
// Summary: Close the  appropriately based on current system conditions.
//
// Parameters:
//   - None.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *MockWatcher) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}
