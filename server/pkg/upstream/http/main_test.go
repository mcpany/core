package http

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Enable loopback resources for all tests in this package by default
	// as they mostly rely on httptest (localhost).
	os.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	os.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES", "true")

	os.Exit(m.Run())
}
