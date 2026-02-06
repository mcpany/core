package integration

import (
	"os"
	"strings"
)

// IsCI returns true if running in a CI environment.
func IsCI() bool {
	ci := os.Getenv("CI")
	return strings.ToLower(ci) == "true" || ci == "1"
}
