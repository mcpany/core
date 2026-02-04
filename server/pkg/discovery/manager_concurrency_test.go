// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"context"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

// SlowMockProvider simulates a slow discovery process.
type SlowMockProvider struct {
	name     string
	duration time.Duration
}

func (m *SlowMockProvider) Name() string {
	return m.name
}

func (m *SlowMockProvider) Discover(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	time.Sleep(m.duration)
	return []*configv1.UpstreamServiceConfig{}, nil
}

func TestManager_Run_Concurrency(t *testing.T) {
	manager := NewManager()

	// Register 3 providers, each taking 100ms
	duration := 100 * time.Millisecond
	manager.RegisterProvider(&SlowMockProvider{name: "slow1", duration: duration})
	manager.RegisterProvider(&SlowMockProvider{name: "slow2", duration: duration})
	manager.RegisterProvider(&SlowMockProvider{name: "slow3", duration: duration})

	start := time.Now()
	manager.Run(context.Background())
	elapsed := time.Since(start)

	t.Logf("Discovery took %v", elapsed)

	// Post-optimization check:
	// Providers run in parallel. Total time should be roughly max(duration) + overhead.
	// 100ms + overhead. Should be well under 200ms.
	// Previous sequential run was ~300ms.
	assert.Less(t, int64(elapsed), int64(200*time.Millisecond), "Expected parallel execution to take less than 200ms")
}
