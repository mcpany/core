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

func (m *mockUnhealthyClient) Close() error { return nil }
func (m *mockUnhealthyClient) IsHealthy(ctx context.Context) bool {
	atomic.AddInt32(m.checkCount, 1)
	return false
}

func TestPoolGet_BusyLoop(t *testing.T) {
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
