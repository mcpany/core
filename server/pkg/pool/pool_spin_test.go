// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package pool

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUnhealthyClient struct {
	checkCount *int32
}

// Close ...
// Summary: Close
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// IsHealthy ...
// Summary: IsHealthy
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	atomic.AddInt32(m.checkCount, 1)
	return false
}

// TestPoolGet_BusyLoop ...
// Summary: TestPoolGet_BusyLoop
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// This test verifies if the pool enters a busy loop when clients are consistently unhealthy.
	var checkCount int32

	factory := func(ctx context.Context) (*mockUnhealthyClient, error) {
		return &mockUnhealthyClient{checkCount: &checkCount}, nil
	}

	// Create a pool with 1 connection, disableHealthCheck=FALSE
	// initial=1, maxIdle=1, maxActive=1
	p, err := New[*mockUnhealthyClient](factory, 1, 1, 1, 0, false)
	require.NoError(t, err)
	defer p.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Try to get a client. It should fail (timeout) because the client is unhealthy.
	_, err = p.Get(ctx)
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)

	// Check how many times IsHealthy was called.
	count := atomic.LoadInt32(&checkCount)
	t.Logf("IsHealthy called %d times in 100ms", count)

	if count < 1000 {
		t.Log("Busy loop not detected (count low)")
	} else {
		t.Log("Busy loop DETECTED")
		// Assert failure to demonstrate the bug
		assert.Fail(t, "Busy loop detected in Pool.Get")
	}
}
