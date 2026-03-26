// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package pool

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// White-box testing for coverage

// TestNew_DisableHealthCheck_FactoryError_Coverage ...
// Summary: TestNew_DisableHealthCheck_FactoryError_Coverage
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	factory := func(ctx context.Context) (*mockClient, error) {
		return nil, errors.New("factory failure")
	}
	// Initial=1
	_, err := New(factory, 1, 1, 5, 0, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "factory failed to create initial client")
}

// TestNew_DisableHealthCheck_NilClient_Coverage ...
// Summary: TestNew_DisableHealthCheck_NilClient_Coverage
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	factory := func(ctx context.Context) (*mockClient, error) {
		return nil, nil // Return nil client without error
	}
	// Initial=1
	_, err := New(factory, 1, 1, 5, 0, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "factory returned nil client")
}

// TestGet_RetryItem_Coverage_Correct ...
// Summary: TestGet_RetryItem_Coverage_Correct
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Create a pool with larger capacity.
	p := newEmptyBufferedPool(t, newMockClientFactory(true), 2, 2)
	require.NotNil(t, p)
	defer func() { _ = p.Close() }()

	pool := p.(*poolImpl[*mockClient])

	// Inject retry item
	pool.clients <- poolItem[*mockClient]{retry: true}

	// Inject healthy client
	expectedClient := &mockClient{isHealthy: true}
	pool.clients <- poolItem[*mockClient]{client: expectedClient}

	// Get should skip retry item and return healthy client
	c, err := p.Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, expectedClient, c)
}

// TestGet_Unhealthy_InFirstLoop_Coverage ...
// Summary: TestGet_Unhealthy_InFirstLoop_Coverage
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Need capacity >= 2 to hold both unhealthy and healthy without blocking
	p := newEmptyBufferedPool(t, newMockClientFactory(true), 2, 2)
	require.NotNil(t, p)
	defer func() { _ = p.Close() }()

	pool := p.(*poolImpl[*mockClient])

	unhealthy := &mockClient{isHealthy: false}
	expected := &mockClient{isHealthy: true}

	// Manually acquire permits for the injected items because Get/Release logic expects permits to be held for items in channel
	// activeCount=0 initially (newEmptyBufferedPool).
	// We inject 2 items. We need 2 permits.
	// tryAcquire(2) should succeed (0+2 <= 2).
	require.True(t, pool.tryAcquire(2))

	// Inject unhealthy client
	pool.clients <- poolItem[*mockClient]{client: unhealthy}
	// Inject healthy client
	pool.clients <- poolItem[*mockClient]{client: expected}

	c, err := p.Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, expected, c)

	// Verify unhealthy one was closed
	// Unhealthy one is pulled, checked, closed, and permit released.
	assert.True(t, unhealthy.isClosed)
}

// TestGet_Race_ClosedAfterAcquire_Coverage ...
// Summary: TestGet_Race_ClosedAfterAcquire_Coverage
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// This attempts to hit the path where TryAcquire succeeds but then we find the pool closed
	// unbuffered (maxIdle=0)
	p, err := New(newMockClientFactory(true), 0, 0, 1, 0, false)
	require.NoError(t, err)

	factory := func(ctx context.Context) (*mockClient, error) {
		// Close the pool inside the factory
		_ = p.Close()
		return &mockClient{isHealthy: true}, nil
	}

	p.(*poolImpl[*mockClient]).factory = factory

	c, err := p.Get(context.Background())
	// Should return ErrPoolClosed because we closed it in factory
	assert.Equal(t, ErrPoolClosed, err)
	if c != nil {
		_ = c.Close()
	}
}

// TestDeregister_CloseError_Coverage ...
// Summary: TestDeregister_CloseError_Coverage
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m := NewManager()

	mockP := &mockPoolWithCloseErr{err: errors.New("close error")}
	m.Register("test", mockP)

	// Deregister should log the error but not fail/panic
	m.Deregister("test")
}

type mockPoolWithCloseErr struct {
	err error
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
	return m.err
}

// Len ...
// Summary: Len
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return 0
}

// TestPut_Closed_DoubleCheck_Coverage ...
// Summary: TestPut_Closed_DoubleCheck_Coverage
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Pre-fill 1
	p, err := New(newMockClientFactory(true), 1, 1, 1, 0, false)
	require.NoError(t, err)

	// Create a healthy client
	c, err := p.Get(context.Background())
	require.NoError(t, err)

	_ = p.Close()

	// Attempt to put the client back into the closed pool
	// This should trigger the closed check within Put (or the initial one)
	p.Put(c)

	// Ensure client is closed
	assert.True(t, c.isClosed)
}

// TestClose_AlreadyClosed_Coverage ...
// Summary: TestClose_AlreadyClosed_Coverage
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	p, err := New(newMockClientFactory(true), 0, 0, 1, 0, false)
	require.NoError(t, err)
	err = p.Close()
	require.NoError(t, err)
	// Close again
	err = p.Close()
	require.NoError(t, err)
}

// TestGet_Wait_Retry_Coverage ...
// Summary: TestGet_Wait_Retry_Coverage
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// This targets the second select statement where we wait for a client.
	// maxIdle=1. Buffered.
	p := newEmptyBufferedPool(t, newMockClientFactory(true), 1, 1)
	require.NotNil(t, p)
	defer func() { _ = p.Close() }()

	// Pool has 1 item and size 1.
	// Get the item to exhaust pool? No, it's empty but buffered.
	// activeCount=0.
	// Get(c1) -> tryAcquire(1) -> Success. active=1.

	c1, err := p.Get(context.Background())
	require.NoError(t, err)

	// Now pool is empty and all permits taken (active=1, max=1).
	// Next Get will block on second select.

	done := make(chan struct{})
	go func() {
		defer close(done)
		c2, err := p.Get(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, c2)
	}()

	// Wait a bit to ensure Get is blocking.
	time.Sleep(10 * time.Millisecond)

	// Inject retry item.
	// Channel size is 1.
	p.(*poolImpl[*mockClient]).clients <- poolItem[*mockClient]{retry: true}

	// The Get loop should consume this retry item and continue loop.
	// Since permits are still exhausted (c1 still out), it will go back to blocking?
	// Wait, retry item doesn't release permit?
	// In Get loop:
	// case item...: if item.retry { continue }
	// It loops back to top.
	// First select: empty? (we just consumed the retry item).
	// TryAcquire: fails (c1 still out).
	// Second select: blocks again.

	// So now we return c1.
	p.Put(c1)

	<-done
}

// TestGet_Wait_Unhealthy_Coverage ...
// Summary: TestGet_Wait_Unhealthy_Coverage
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// maxIdle=1. Buffered.
	p := newEmptyBufferedPool(t, newMockClientFactory(true), 1, 1)
	require.NotNil(t, p)
	defer func() { _ = p.Close() }()

	c1, err := p.Get(context.Background())
	require.NoError(t, err)

	done := make(chan struct{})
	go func() {
		defer close(done)
		c2, err := p.Get(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, c2)
		// Should be a new client
		assert.NotEqual(t, c1, c2)
	}()

	time.Sleep(10 * time.Millisecond)

	// Make c1 unhealthy and put it back
	c1.mu.Lock()
	c1.isHealthy = false
	c1.mu.Unlock()
	p.Put(c1)

	<-done
}
