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

type SleepingProvider struct {
	name     string
	duration time.Duration
}

func (m *SleepingProvider) Name() string {
	return m.name
}

func (m *SleepingProvider) Discover(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	time.Sleep(m.duration)
	return []*configv1.UpstreamServiceConfig{
		configv1.UpstreamServiceConfig_builder{Name: &m.name}.Build(),
	}, nil
}

func TestManager_Run_Parallel(t *testing.T) {
	manager := NewManager()

	// Register 2 providers that sleep for 500ms each
	p1 := &SleepingProvider{name: "p1", duration: 500 * time.Millisecond}
	p2 := &SleepingProvider{name: "p2", duration: 500 * time.Millisecond}

	manager.RegisterProvider(p1)
	manager.RegisterProvider(p2)

	start := time.Now()
	services := manager.Run(context.Background())
	duration := time.Since(start)

	assert.Len(t, services, 2)

	// If sequential, it takes >= 1000ms. If parallel, it takes ~500ms.
	// We assert it takes less than 800ms to allow for some overhead but prove parallelism.
	assert.Less(t, duration, 800*time.Millisecond, "Discovery took too long, expected parallel execution")
}
