// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package webrtc

import (
	"context"
	"testing"

	"github.com/mcpany/core/pkg/pool"
	"github.com/stretchr/testify/assert"
)

func TestUpstream_Shutdown(t *testing.T) {
	u := NewUpstream(pool.NewManager())
	err := u.Shutdown(context.Background())
	assert.NoError(t, err)
}
