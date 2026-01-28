// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	os.Exit(m.Run())
}
