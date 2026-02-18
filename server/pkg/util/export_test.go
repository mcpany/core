// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import "sync"

// ResetSafeHTTPClient resets the internal singleton. Strictly for testing.
// This is exported only for testing purposes (white-box testing from external packages).
func ResetSafeHTTPClient() {
	transportOnce = sync.Once{}
	sharedSafeTransport = nil
}
