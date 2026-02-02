// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package upstream

import "strings"

func contains(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}
