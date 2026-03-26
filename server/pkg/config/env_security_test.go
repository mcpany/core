// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestExpand_RestrictedVar ...
// Summary: TestExpand_RestrictedVar
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Setup restricted variable
	os.Setenv("MCPANY_SECRET_KEY", "super-secret")
	defer os.Unsetenv("MCPANY_SECRET_KEY")

	// Input with restricted var
	input := []byte("secret: ${MCPANY_SECRET_KEY}")

	// Expect error
	_, err := expand(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable MCPANY_SECRET_KEY is restricted")
}

// TestExpand_AllowedVar ...
// Summary: TestExpand_AllowedVar
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Setup allowed variable via Allowlist
	os.Setenv("MCPANY_ALLOWED_ENV", "MCPANY_PUBLIC_*")
	defer os.Unsetenv("MCPANY_ALLOWED_ENV")

	os.Setenv("MCPANY_PUBLIC_KEY", "public-key")
	defer os.Unsetenv("MCPANY_PUBLIC_KEY")

	input := []byte("key: ${MCPANY_PUBLIC_KEY}")

	// Expect success
	expanded, err := expand(input)
	assert.NoError(t, err)
	assert.Equal(t, "key: public-key", string(expanded))
}

// TestExpand_StrictMode ...
// Summary: TestExpand_StrictMode
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Enable Strict Mode
	os.Setenv("MCPANY_STRICT_ENV_MODE", "true")
	defer os.Unsetenv("MCPANY_STRICT_ENV_MODE")

	os.Setenv("SOME_VAR", "some-value")
	defer os.Unsetenv("SOME_VAR")

	input := []byte("val: ${SOME_VAR}")

	// Expect error because NOT in allowlist and Strict Mode is ON
	_, err := expand(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable SOME_VAR is restricted")
}
