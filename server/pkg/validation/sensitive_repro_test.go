// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSensitivePath_Repro(t *testing.T) {
	// Vulnerability: Accessing files inside .git directory
	err := IsSensitivePath(".git/config")
	assert.Error(t, err, "Should fail because .git is sensitive")
	assert.Contains(t, err.Error(), "access to sensitive directory \".git\" is denied")

	// Vulnerability: Accessing config files with different names
	err = IsSensitivePath("config.prod.yaml")
	assert.Error(t, err, "Should fail because config.prod.yaml is sensitive")
	assert.Contains(t, err.Error(), "access to sensitive configuration file \"config.prod.yaml\" is denied")

    // Additional checks
    err = IsSensitivePath("foo/.ssh/authorized_keys")
	assert.Error(t, err, "Should fail because .ssh is sensitive")

    err = IsSensitivePath(".env.local")
	assert.Error(t, err, "Should fail because .env.local is sensitive")
}
