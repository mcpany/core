package tool

import (
	"fmt"
	"os"
	"testing"

	"github.com/mcpany/core/server/pkg/validation"
)

func TestMain(m *testing.M) {
	fmt.Println("DEBUG: Running TestMain in package tool")
	// Mock IsSafeURL to allow all URLs during tests in this package.
	// This is necessary because many tests use httptest.NewServer which runs on 127.0.0.1.
	// We also set the env var to ensure any test restoring IsSafeURL still passes.
	os.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")

	originalIsSafeURL := validation.IsSafeURL
	validation.IsSafeURL = func(urlStr string) error { return nil }
	defer func() { validation.IsSafeURL = originalIsSafeURL }()

	os.Exit(m.Run())
}
