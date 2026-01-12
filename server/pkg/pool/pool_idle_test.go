package pool

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPool_IdleTimeout(t *testing.T) {
	// Create a pool with short idle timeout
	idleTimeout := 50 * time.Millisecond
	factoryCalls := 0
	factory := func(ctx context.Context) (*mockClient, error) {
		factoryCalls++
		return &mockClient{id: factoryCalls, isHealthy: true}, nil
	}

	p, err := New(factory, 0, 1, idleTimeout, false)
	require.NoError(t, err)
	defer p.Close()

	ctx := context.Background()

	// 1. Get a client
	c1, err := p.Get(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, c1.id)

	// 2. Return it
	p.Put(c1)

	// 3. Wait for idle timeout
	time.Sleep(100 * time.Millisecond)

	// 4. Get a client again
	// EXPECTATION: c1 should be expired and discarded. We should get a NEW client (id 2).
	c2, err := p.Get(ctx)
	require.NoError(t, err)

	// With the optimization, this should pass.
	assert.NotEqual(t, c1.id, c2.id, "Client should have been expired and replaced")
	assert.Equal(t, 2, c2.id)
	assert.True(t, c1.isClosed, "Expired client should be closed")
}
