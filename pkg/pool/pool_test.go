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
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockClient struct {
	id        int
	isHealthy bool
	isClosed  bool
	closeErr  error
	mu        sync.RWMutex
}

func (c *mockClient) IsHealthy() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isHealthy
}

func (c *mockClient) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isClosed
}

func (c *mockClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.isClosed = true
	return c.closeErr
}

var (
	clientIDCounter int32
	ErrPoolFull     = fmt.Errorf("pool is full")
)

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

func TestPool_Full(t *testing.T) {
	p, err := New(newMockClientFactory(true), 0, 1, 100)
	require.NoError(t, err)
	defer p.Close()
	// Get one client, pool should now be at max capacity
	c1, err := p.Get(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, c1)
	// Try to get another, should be full
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, err = p.Get(ctx)
	assert.Error(t, err)
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

func TestPool_Put_UnhealthyClientDoesNotLeakSemaphore(t *testing.T) {
	const maxSize = 2
	p, err := New(newMockClientFactory(true), 0, maxSize, 100)
	require.NoError(t, err)
	defer p.Close()
	// Cycle through clients, marking them as unhealthy before returning them.
	// This should not exhaust the pool's semaphore.
	for i := 0; i < maxSize+1; i++ {
		// Get a client. This acquires a semaphore permit.
		c, err := p.Get(context.Background())
		require.NoError(t, err, "Pool should not be exhausted on iteration %d", i)
		require.NotNil(t, c)
		// Mark it as unhealthy.
		c.isHealthy = false
		// Put it back. This should release the semaphore permit.
		// The bug is that it doesn't, leading to a leak.
		p.Put(c)
		// The unhealthy client should have been closed.
		assert.True(t, c.isClosed, "Client should be closed after being put back as unhealthy")
		// The pool should not store the unhealthy client.
		assert.Equal(t, 0, p.Len(), "Pool should be empty after returning an unhealthy client")
	}
}

func TestPool_PutOnClosedPool(t *testing.T) {
	const maxSize = 1
	p, err := New(newMockClientFactory(true), 0, maxSize, 100)
	require.NoError(t, err)
	// Get the only client
	c, err := p.Get(context.Background())
	require.NoError(t, err)
	require.NotNil(t, c)
	// Close the pool while the client is checked out
	p.Close()
	// Now, put the client back into the closed pool.
	// This is where the semaphore leak occurs in the original code.
	p.Put(c)
	// The client should be closed because the pool is closed.
	assert.True(t, c.isClosed)
	// To verify the fix, we check if we can acquire and release the semaphore again.
	// In the buggy version, the semaphore is never released, so this would hang.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err = p.(*poolImpl[*mockClient]).sem.Acquire(ctx, 1)
	require.NoError(t, err, "Semaphore should not be locked after returning a client to a closed pool")
	p.(*poolImpl[*mockClient]).sem.Release(1)
}

func TestPool_ConcurrentGetPut(t *testing.T) {
	const (
		maxSize    = 20
		numClients = 100
		iterations = 200
	)
	p, err := New(newMockClientFactory(true), 10, maxSize, 100)
	require.NoError(t, err)
	defer p.Close()
	var wg sync.WaitGroup
	wg.Add(numClients)
	for i := 0; i < numClients; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				client, err := p.Get(ctx)
				if err != nil {
					cancel()
					// It's acceptable to get an error under high contention when the pool is full
					// and the context times out.
					time.Sleep(time.Duration(rand.Intn(20)) * time.Millisecond)
					continue
				}
				require.NotNil(t, client, "Goroutine %d received a nil client", goroutineID)
				// Simulate some work with random duration
				time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
				p.Put(client)
				cancel()
			}
		}(i)
	}
	wg.Wait()
}

func TestManager_Concurrent(t *testing.T) {
	m := NewManager()
	var wg sync.WaitGroup
	numRoutines := 100
	// Test concurrent registration
	wg.Add(numRoutines)
	for i := 0; i < numRoutines; i++ {
		go func(i int) {
			defer wg.Done()
			poolName := fmt.Sprintf("pool_%d", i%10) // Create contention on a few pool names
			p, err := New(newMockClientFactory(true), 1, 2, 100)
			require.NoError(t, err)
			m.Register(poolName, p)
		}(i)
	}
	wg.Wait()
	// Test concurrent Get
	wg.Add(numRoutines)
	for i := 0; i < numRoutines; i++ {
		go func(i int) {
			defer wg.Done()
			poolName := fmt.Sprintf("pool_%d", i%10)
			retrievedPool, ok := Get[*mockClient](m, poolName)
			assert.True(t, ok)
			assert.NotNil(t, retrievedPool)
		}(i)
	}
	wg.Wait()
	m.CloseAll()
}

func TestPool_ConcurrentClose(t *testing.T) {
	pool, _ := New(newMockClientFactory(true), 0, 10, 0)
	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			pool.Close()
		}()
	}
	wg.Wait()
}

func TestPool_GetWithUnhealthyClients(t *testing.T) {
	t.Parallel()
	maxSize := 5
	factory := newMockClientFactory(true)
	pool, err := New(factory, maxSize, maxSize, 0)
	require.NoError(t, err)
	// Invalidate all the initial clients in the pool
	for i := 0; i < maxSize; i++ {
		c, err := pool.Get(context.Background())
		require.NoError(t, err)
		c.mu.Lock()
		c.isHealthy = false
		c.mu.Unlock()
		pool.Put(c)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		client, err := pool.Get(ctx)
		require.NoError(t, err, "Get should not fail")
		require.NotNil(t, client, "Should have received a client")
		assert.True(t, client.IsHealthy(), "The new client should be healthy")
	}()
	wg.Wait()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatal("Test timed out, which indicates a likely deadlock due to semaphore leak.")
	}
}

func TestManager_ConcurrentAccess(t *testing.T) {
	m := NewManager()
	var wg sync.WaitGroup
	numGoroutines := 50
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			poolName := "test_pool"
			// Concurrently register and get pools
			p, err := New(newMockClientFactory(true), 1, 2, 100)
			require.NoError(t, err)
			m.Register(poolName, p)
			retrievedPool, ok := Get[*mockClient](m, poolName)
			assert.True(t, ok)
			assert.NotNil(t, retrievedPool)
		}(i)
	}
	wg.Wait()
	m.CloseAll()
}

func TestPool_ConcurrentGetAndClose(t *testing.T) {
	const (
		maxSize    = 10
		numGetters = 50
	)
	p, err := New(newMockClientFactory(true), 2, maxSize, 100)
	require.NoError(t, err)
	var wg sync.WaitGroup
	wg.Add(numGetters)
	// Start a goroutine to close the pool after a short delay.
	go func() {
		time.Sleep(15 * time.Millisecond)
		p.Close()
	}()
	for i := 0; i < numGetters; i++ {
		go func() {
			defer wg.Done()
			// Continuously get and put clients until the pool is closed.
			for {
				client, err := p.Get(context.Background())
				if err == ErrPoolClosed {
					break // Exit loop when pool is closed.
				}
				if err != nil {
					// Other errors are not expected in this test.
					require.NoError(t, err)
				}
				// Simulate some work.
				time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
				// Putting a client back into a closed pool is a valid operation.
				p.Put(client)
			}
		}()
	}
	wg.Wait()
	assert.True(t, p.(*poolImpl[*mockClient]).closed, "Pool should be closed")
	_, err = p.Get(context.Background())
	assert.Equal(t, ErrPoolClosed, err, "Getting from a closed pool should return ErrPoolClosed")
}
