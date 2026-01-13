// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

import (
	"context"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestTimeout_Execute(t *testing.T) {
	tests := []struct {
		name        string
		timeout     time.Duration
		workDelay   time.Duration
		expectError bool
	}{
		{
			name:        "Success within timeout",
			timeout:     100 * time.Millisecond,
			workDelay:   10 * time.Millisecond,
			expectError: false,
		},
		{
			name:        "Failure due to timeout",
			timeout:     10 * time.Millisecond,
			workDelay:   100 * time.Millisecond,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeout := NewTimeout(durationpb.New(tt.timeout))
			err := timeout.Execute(context.Background(), func(ctx context.Context) error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(tt.workDelay):
					return nil
				}
			})

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, context.DeadlineExceeded, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManager_Execute_WithTimeout(t *testing.T) {
	config := &configv1.ResilienceConfig{
		Timeout: durationpb.New(50 * time.Millisecond),
	}
	manager := NewManager(config)

	t.Run("Timeout triggers", func(t *testing.T) {
		err := manager.Execute(context.Background(), func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(100 * time.Millisecond):
				return nil
			}
		})
		assert.Error(t, err)
		assert.Equal(t, context.DeadlineExceeded, err)
	})

	t.Run("Timeout does not trigger", func(t *testing.T) {
		err := manager.Execute(context.Background(), func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(10 * time.Millisecond):
				return nil
			}
		})
		assert.NoError(t, err)
	})
}
