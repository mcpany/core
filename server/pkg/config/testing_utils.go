// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import "sync"

// ResetGlobalSettings resets the global settings singleton for testing purposes.
// This function is intended to be used only in tests.
func ResetGlobalSettings() {
	globalSettings = nil
	once = sync.Once{}
}
