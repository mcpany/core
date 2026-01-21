// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package pool

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPool_DoubleRelease_ExceedsMaxSize(t *testing.T) {
	// We want to simulate a scenario where Get fails (releasing permit),
	// and user calls Put(nil) (releasing permit again), causing negative activeCount.
	// Then we verify we can create MORE than maxSize clients.

	maxSize := 1
	var failFactory atomic.Bool
	failFactory.Store(true)

	factory := func(_ context.Context) (*mockClient, error) {
		if failFactory.Load() {
			return nil, errors.New("factory failed")
		}
		return &mockClient{isHealthy: true}, nil
	}

	p, err := New(factory, 0, maxSize, 0, false)
	require.NoError(t, err)
	defer func() { _ = p.Close() }()

	// 1. Trigger Get failure
	// This will try to acquire permit -> success (activeCount=1)
	// Factory fails -> release permit (activeCount=0)
	// Returns error.
	c, err := p.Get(context.Background())
	require.Error(t, err)
	require.Nil(t, c)

	// 2. User mistakenly Puts nil (e.g. via defer)
	// Put(nil) -> releases permit (activeCount=-1)
	p.Put(nil)

	// Now activeCount should be -1.
	// We enable factory.
	failFactory.Store(false)

	// 3. Get client 1 (activeCount -> 0)
	c1, err := p.Get(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, c1)

	// 4. Get client 2 (activeCount -> 1)
	// If activeCount was correctly 0 before step 3, then step 3 made it 1.
	// This step would try to make it 2, which > maxSize(1), so it should fail.
	// But if activeCount was -1, step 3 made it 0. This step makes it 1. Success.

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	c2, err := p.Get(ctx)

	// If bug is present, we succeeded in getting c2.
	// We want the test to fail if the bug is present.
	// So we assert that we EXPECT an error (DeadlineExceeded).
	assert.ErrorIs(t, err, context.DeadlineExceeded, "Should not be able to get more clients than maxSize")

	if err == nil {
		p.Put(c2)
	}
	p.Put(c1)
}
