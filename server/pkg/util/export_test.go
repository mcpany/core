// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import "sync"

// ResetSafeHTTPClientForTest resets the shared transport for testing purposes.
func ResetSafeHTTPClientForTest() {
	sharedTransport = nil
	sharedTransportOnce = sync.Once{}
}
