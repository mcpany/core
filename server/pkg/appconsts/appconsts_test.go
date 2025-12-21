// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package appconsts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppConsts(t *testing.T) {
	assert.NotEmpty(t, Name)
	assert.NotEmpty(t, Version)
}
