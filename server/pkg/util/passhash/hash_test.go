// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package passhash

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPassword(t *testing.T) {
	password := "mysecretpassword"
	hash, err := Password(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)

	assert.True(t, CheckPassword(password, hash))
	assert.False(t, CheckPassword("wrongpassword", hash))
}

func TestPassword_Error(t *testing.T) {
    // bcrypt returns error if password is too long (> 72 bytes)
	longPassword := strings.Repeat("a", 73)
	hash, err := Password(longPassword)
	assert.Error(t, err)
    assert.Contains(t, err.Error(), "password length exceeds 72 bytes")
    assert.Empty(t, hash)
}
