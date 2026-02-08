package health

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Allow loopback resources for tests
	os.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	code := m.Run()
	os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")
	os.Exit(code)
}
