// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package browser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUpstream(t *testing.T) {
	u := NewUpstream()
	assert.NotNil(t, u)
	assert.False(t, u.initialized)
}
