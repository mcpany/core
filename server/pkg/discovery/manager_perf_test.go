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

type SlowProvider struct {
	name  string
	delay time.Duration
}

func (s *SlowProvider) Name() string {
	return s.name
}

func (s *SlowProvider) Discover(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	time.Sleep(s.delay)
	return []*configv1.UpstreamServiceConfig{
		configv1.UpstreamServiceConfig_builder{Name: &s.name}.Build(),
	}, nil
}

func TestManager_Run_Parallel(t *testing.T) {
	manager := NewManager()

	delay := 100 * time.Millisecond
	providerCount := 3

	// Register multiple slow providers
	for i := 0; i < providerCount; i++ {
		manager.RegisterProvider(&SlowProvider{
			name:  string(rune('A' + i)),
			delay: delay,
		})
	}

	start := time.Now()
	services := manager.Run(context.Background())
	duration := time.Since(start)

	assert.Len(t, services, providerCount)

	// If serial, duration would be >= delay * providerCount (300ms)
	// If parallel, duration should be ~ delay (100ms) + overhead
	// We set a lenient threshold of 200ms (2x delay) to account for CI slowness,
	// but significantly less than 3x delay.
	//
	// Note: Before the fix, this assertion is expected to FAIL.
	assert.Less(t, duration, delay*2, "Discovery took too long, likely running serially")
}
