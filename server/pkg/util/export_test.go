// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import "net/http"

// ResetSafeHTTPClientForTest resets the shared transport for testing purposes.
func ResetSafeHTTPClientForTest() {
	transportMu.Lock()
	defer transportMu.Unlock()
	transportCache = make(map[string]*http.Transport)
}
