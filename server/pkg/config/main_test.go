// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Create dummy cert file for tests
	_ = os.WriteFile("dummy-cert.pem", []byte("cert"), 0600)
	code := m.Run()
	_ = os.Remove("dummy-cert.pem")
	os.Exit(code)
}
