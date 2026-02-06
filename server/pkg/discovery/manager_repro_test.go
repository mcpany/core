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
	// Use 3 providers to make the difference between sequential (300ms) and parallel (100ms) more distinct
	p1 := &DelayedProvider{name: "p1", delay: delay}
	p2 := &DelayedProvider{name: "p2", delay: delay}
	p3 := &DelayedProvider{name: "p3", delay: delay}

	manager.RegisterProvider(p1)
	manager.RegisterProvider(p2)
	manager.RegisterProvider(p3)

	start := time.Now()
	manager.Run(context.Background())
	duration := time.Since(start)

	// In parallel execution, duration should be roughly equal to max delay (100ms)
	// We set a generous upper bound: it should be significantly faster than sequential (300ms).
	// Let's say it must be faster than 250ms.
	// 250ms allows for 150ms of overhead which is huge but safe for slow CI.
	maxExpectedDuration := delay * 2 + (delay / 2) // 250ms

	assert.Less(t, duration, maxExpectedDuration, "Expected parallel execution to be significantly faster than sequential sum")
	assert.GreaterOrEqual(t, duration, delay, "Expected execution to take at least max delay")
}
