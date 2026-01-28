package graphql

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	os.Exit(m.Run())
}
