// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package pool

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPool_PanicInFactory(t *testing.T) {
	// Factory panics.
	factory := func(_ context.Context) (*mockClient, error) {
		panic("factory panic")
	}

	// We use minSize=0 to start empty.
	p, err := New(factory, 0, 1, 0, false)
	require.NoError(t, err)
	defer func() { _ = p.Close() }()

	// First Get should panic.
	assert.Panics(t, func() {
		_, _ = p.Get(context.Background())
	})

	// The permit acquired before factory call should be released.
	// If not released, subsequent Get calls (with a non-panicking factory if we could switch it,
	// or simply checking if we can acquire permit) will fail/block if maxSize=1.

	// Since we can't change the factory of the existing pool instance easily to verify
	// if permit is available, we can try to access the semaphore directly via reflection or unsafe,
	// OR we can check if the pool is "stuck" by making a new request that should work if permit is available.
	// But the factory is baked in and panics.

	// So this test setup is tricky.
	// We need a factory that panics ONCE, then works.
}

func TestPool_PanicInFactory_Recover(t *testing.T) {
	panicked := false
	factory := func(_ context.Context) (*mockClient, error) {
		if !panicked {
			panicked = true
			panic("factory panic")
		}
		return &mockClient{isHealthy: true}, nil
	}

	p, err := New(factory, 0, 1, 0, false)
	require.NoError(t, err)
	defer func() { _ = p.Close() }()

	// First Get panics
	assert.Panics(t, func() {
		_, _ = p.Get(context.Background())
	})

	// Second Get should succeed if permit was released.
	// If permit was leaked, the pool (maxSize=1) thinks it's full.
	// Since channel is empty, it will try to create new, but TryAcquire(1) will fail.
	// Then it waits on channel (which is empty).
	// So it blocks forever.

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	client, err := p.Get(ctx)
	if err != nil {
		t.Logf("Get failed: %v", err)
	} else {
		t.Logf("Get succeeded")
		p.Put(client)
	}

	// If permit leaked, we expect timeout (DeadlineExceeded)
	assert.NoError(t, err, "Pool should recover after factory panic")
}

func TestPool_PanicInIsHealthy(t *testing.T) {
	// Factory creates client that panics on IsHealthy
	factory := func(_ context.Context) (*mockClient, error) {
		return &mockClient{isHealthy: true}, nil // Initially healthy-ish
	}

	p, err := New(factory, 0, 1, 0, false)
	require.NoError(t, err)
	defer func() { _ = p.Close() }()

	// Get client
	c, err := p.Get(context.Background())
	require.NoError(t, err)

	// Put it back.
	p.Put(c)

	// Now we hack the client to panic on IsHealthy (which is called by Get when retrieving from pool)
	// We can't easily change the behavior of existing struct instance if it's not designed for it.
	// But our mockClient has `isHealthy` bool. It doesn't have a hook.
	// We need a new mock client that panics.
}

type panicMockClient struct {
	mockClient
	panicOnHealth bool
}

func (c *panicMockClient) IsHealthy(ctx context.Context) bool {
	if c.panicOnHealth {
		panic("health check panic")
	}
	return c.mockClient.IsHealthy(ctx)
}

func TestPool_PanicInIsHealthy_Recover(t *testing.T) {
	factory := func(_ context.Context) (*panicMockClient, error) {
		return &panicMockClient{mockClient: mockClient{isHealthy: true}}, nil
	}

	// maxSize=1
	p, err := New(factory, 0, 1, 0, false)
	require.NoError(t, err)
	defer func() { _ = p.Close() }()

	c, err := p.Get(context.Background())
	require.NoError(t, err)

	// Put it back
	p.Put(c)

	// Now set it to panic
	c.panicOnHealth = true

	// Next Get will retrieve it, call IsHealthy -> Panic.
	assert.Panics(t, func() {
		_, _ = p.Get(context.Background())
	})

	// If permit leaked, pool thinks it's full (permit held for the client that caused panic).
	// Channel is empty (client was taken out).
	// TryAcquire fails.
	// Waits on channel forever.

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Try to get another client (should create new one as the old one "died")
	// But to create new one, permit must be available.
	c2, err := p.Get(ctx)
	assert.NoError(t, err, "Pool should recover after IsHealthy panic")
	if c2 != nil {
		p.Put(c2)
	}
}
