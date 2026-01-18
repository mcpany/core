// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import "sync"

// ResetSafeSecretClientForTest resets the singleton for testing purposes.
// This file is only compiled during tests.
func ResetSafeSecretClientForTest() {
	safeSecretClientOnce = sync.Once{}
	safeSecretClient = nil
}
