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

type DelayedProvider struct {
	name  string
	delay time.Duration
}

func (d *DelayedProvider) Name() string {
	return d.name
}

func (d *DelayedProvider) Discover(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	time.Sleep(d.delay)
	return []*configv1.UpstreamServiceConfig{}, nil
}

func TestManager_Run_Parallelism(t *testing.T) {
	manager := NewManager()

	delay := 100 * time.Millisecond
	p1 := &DelayedProvider{name: "p1", delay: delay}
	p2 := &DelayedProvider{name: "p2", delay: delay}

	manager.RegisterProvider(p1)
	manager.RegisterProvider(p2)

	start := time.Now()
	manager.Run(context.Background())
	duration := time.Since(start)

	// In parallel execution, duration should be roughly equal to max delay (100ms)
	// It should definitely be less than sum of delays (200ms)
	assert.Less(t, duration, delay*2-10*time.Millisecond, "Expected parallel execution to be faster than sum of delays")
	assert.GreaterOrEqual(t, duration, delay, "Expected execution to take at least max delay")
}
