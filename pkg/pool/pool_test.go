/*
 * Copyright 2025 Author(s) of MCP-XY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package pool

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockClient struct {
	id        int
	isHealthy bool
	isClosed  bool
	closeErr  error
}

func (c *mockClient) IsHealthy() bool {
	return c.isHealthy
}

func (c *mockClient) Close() error {
	c.isClosed = true
	return c.closeErr
}

var clientIDCounter int32

func newMockClientFactory(healthy bool) func(ctx context.Context) (*mockClient, error) {
	return func(ctx context.Context) (*mockClient, error) {
		id := atomic.AddInt32(&clientIDCounter, 1)
		return &mockClient{id: int(id), isHealthy: healthy}, nil
	}
}

func TestPool_New(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		p, err := New(newMockClientFactory(true), 1, 5, 100)
		require.NoError(t, err)
		assert.NotNil(t, p)
		assert.Equal(t, 1, p.Len())
		p.Close()
	})

	t.Run("invalid config", func(t *testing.T) {
		_, err := New(newMockClientFactory(true), -1, 5, 100)
		assert.Error(t, err)
	})
}

func TestPool_GetPut(t *testing.T) {
	p, err := New(newMockClientFactory(true), 1, 2, 100)
	require.NoError(t, err)
	defer p.Close()

	// Get initial client
	c1, err := p.Get(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, c1)
	assert.Equal(t, 0, p.Len())

	// Put it back
	p.Put(c1)
	assert.Equal(t, 1, p.Len())

	// Get it again
	c2, err := p.Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, c1, c2)
}

func TestPool_Get_Unhealthy(t *testing.T) {
	// Factory creates one unhealthy client, then healthy ones.
	var createdCount int32
	factory := func(ctx context.Context) (*mockClient, error) {
		count := atomic.AddInt32(&createdCount, 1)
		// First client is unhealthy, subsequent are healthy.
		return &mockClient{isHealthy: count > 1}, nil
	}

	// Create pool with minSize=1, so it starts with one (unhealthy) client.
	p, err := New(factory, 1, 2, 100)
	require.NoError(t, err)
	defer p.Close()

	// Get should discard the initial unhealthy client and create a new, healthy one.
	c, err := p.Get(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, c)
	assert.True(t, c.IsHealthy())
}

func TestPool_Put_Unhealthy(t *testing.T) {
	p, err := New(newMockClientFactory(true), 0, 2, 100)
	require.NoError(t, err)
	defer p.Close()

	// Get a client. This will acquire a semaphore.
	c, err := p.Get(context.Background())
	require.NoError(t, err)

	// Make it unhealthy.
	c.isHealthy = false

	// Put it back. This will release the semaphore.
	p.Put(c)

	// Pool should not store the unhealthy client.
	assert.Equal(t, 0, p.Len())
	assert.True(t, c.isClosed)
}

func TestPool_PutUnhealthyClientReleasesSemaphore(t *testing.T) {
	// This test ensures that putting an unhealthy client back into the pool releases
	// the semaphore permit, preventing a leak.
	p, err := New(newMockClientFactory(true), 0, 1, 100)
	require.NoError(t, err)
	defer p.Close()

	// Get a client. This acquires the only permit.
	c, err := p.Get(context.Background())
	require.NoError(t, err)
	require.NotNil(t, c)

	// Make the client unhealthy.
	c.isHealthy = false

	// Put the unhealthy client back.
	// Before the fix, this does not release the semaphore.
	p.Put(c)

	// Now, try to get a client again.
	// Before the fix, this will fail with ErrPoolFull because the semaphore was not released.
	// After the fix, this should succeed by creating a new client.
	c2, err := p.Get(context.Background())
	require.NoError(t, err, "Should be able to get a new client after returning an unhealthy one")
	assert.NotNil(t, c2)
}

func TestPool_Full(t *testing.T) {
	p, err := New(newMockClientFactory(true), 0, 1, 100)
	require.NoError(t, err)
	defer p.Close()

	// Get one client, pool should now be at max capacity
	c1, err := p.Get(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, c1)

	// Try to get another, should be full
	_, err = p.Get(context.Background())
	assert.Equal(t, ErrPoolFull, err)
}

func TestPool_Close(t *testing.T) {
	client := &mockClient{isHealthy: true}
	factory := func(ctx context.Context) (*mockClient, error) {
		return client, nil
	}

	p, err := New(factory, 1, 1, 100)
	require.NoError(t, err)
	p.Close()

	assert.True(t, client.isClosed)
}

func TestManager(t *testing.T) {
	m := NewManager()
	p, err := New(newMockClientFactory(true), 1, 5, 100)
	require.NoError(t, err)

	m.Register("test_pool", p)

	retrievedPool, ok := Get[*mockClient](m, "test_pool")
	require.True(t, ok)
	assert.Equal(t, p, retrievedPool)

	_, ok = Get[*mockClient](m, "nonexistent_pool")
	assert.False(t, ok)

	m.CloseAll()
	// After closing, the pool should be empty, assuming Close() works
}

type simpleMockPool struct {
	closed bool
}

func (p *simpleMockPool) Close() {
	p.closed = true
}

func (p *simpleMockPool) Len() int {
	return 0
}

func TestManager_RegisterOverwriteClosesOldPool(t *testing.T) {
	m := NewManager()
	pool1 := &simpleMockPool{}
	pool2 := &simpleMockPool{}

	m.Register("test_pool", pool1)
	m.Register("test_pool", pool2) // This should close pool1

	assert.True(t, pool1.closed, "Expected old pool to be closed upon re-registration")
	assert.False(t, pool2.closed, "Expected new pool to not be closed")
}
