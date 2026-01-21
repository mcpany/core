// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package browser

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUpstream(t *testing.T) {
	u := NewUpstream()
	assert.NotNil(t, u)
	assert.False(t, u.initialized)
}

func TestUpstream_CheckHealth(t *testing.T) {
	u := NewUpstream()

	// Not initialized
	err := u.CheckHealth(context.Background())
	assert.Error(t, err)
	assert.Equal(t, "browser upstream not initialized", err.Error())

	// Simulate initialized but no page (partial initialization state or corruption)
	u.initialized = true
	err = u.CheckHealth(context.Background())
	assert.Error(t, err)
	assert.Equal(t, "browser page not available", err.Error())
}

func TestUpstream_Shutdown(t *testing.T) {
	u := NewUpstream()
	// Shutdown on uninitialized upstream should be safe
	err := u.Shutdown(context.Background())
	assert.NoError(t, err)
	assert.False(t, u.initialized)
}

func TestUpstream_handleToolExecution_NotInitialized(t *testing.T) {
	u := NewUpstream()
	_, err := u.handleToolExecution(context.Background(), "browser_navigate", nil)
	assert.Error(t, err)
	assert.Equal(t, "browser not initialized", err.Error())
}
