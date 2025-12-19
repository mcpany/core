package util_test

import (
	"path/filepath"
	"testing"

	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestResolveSecret_AbsolutePathsBlocked(t *testing.T) {
	t.Setenv("MCPANY_ENFORCE_RELATIVE_PATHS", "true")

	// Use an absolute path
	absPath, err := filepath.Abs("secret.txt")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	secret := &configv1.SecretValue{}
	secret.SetFilePath(absPath)

	// This should now FAIL because of IsRelativePath check
	_, err = util.ResolveSecret(secret)
	assert.Error(t, err, "ResolveSecret should block absolute paths")
	if err != nil {
		assert.Contains(t, err.Error(), "path must be relative")
	}
}
