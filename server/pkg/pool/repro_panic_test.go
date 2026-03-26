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

// TestPool_PanicInFactory ...
// Summary: TestPool_PanicInFactory
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Factory panics.
	factory := func(_ context.Context) (*mockClient, error) {
		panic("factory panic")
	}

	// We use initial=0. Unbuffered.
	p, err := New(factory, 0, 0, 1, 0, false)
	require.NoError(t, err)
	defer func() { _ = p.Close() }()

	// First Get should panic.
	assert.Panics(t, func() {
		_, _ = p.Get(context.Background())
	})
}

// TestPool_PanicInFactory_Recover ...
// Summary: TestPool_PanicInFactory_Recover
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	panicked := false
	factory := func(_ context.Context) (*mockClient, error) {
		if !panicked {
			panicked = true
			panic("factory panic")
		}
		return &mockClient{isHealthy: true}, nil
	}

	// initial=0. Unbuffered.
	p, err := New(factory, 0, 0, 1, 0, false)
	require.NoError(t, err)
	defer func() { _ = p.Close() }()

	// First Get panics
	assert.Panics(t, func() {
		_, _ = p.Get(context.Background())
	})

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

// TestPool_PanicInIsHealthy ...
// Summary: TestPool_PanicInIsHealthy
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Factory creates client that panics on IsHealthy
	factory := func(_ context.Context) (*mockClient, error) {
		return &mockClient{isHealthy: true}, nil // Initially healthy-ish
	}

	p, err := New(factory, 0, 0, 1, 0, false)
	require.NoError(t, err)
	defer func() { _ = p.Close() }()

	// Get client
	c, err := p.Get(context.Background())
	require.NoError(t, err)

	// Put it back.
	p.Put(c)
}

type panicMockClient struct {
	mockClient
	panicOnHealth bool
}

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
	if c.panicOnHealth {
		panic("health check panic")
	}
	return c.mockClient.IsHealthy(ctx)
}

// TestPool_PanicInIsHealthy_Recover ...
// Summary: TestPool_PanicInIsHealthy_Recover
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	factory := func(_ context.Context) (*panicMockClient, error) {
		return &panicMockClient{mockClient: mockClient{isHealthy: true}}, nil
	}

	// maxSize=1. Use maxIdleSize=1 to allow buffering.
	// Initial=1.
	p, err := New(factory, 1, 1, 1, 0, false)
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
