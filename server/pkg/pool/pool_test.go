package pool

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
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

func (c *mockClient) IsHealthy(_ context.Context) bool {
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

var clientIDCounter int32

func newMockClientFactory(healthy bool) func(ctx context.Context) (*mockClient, error) {
	return func(_ context.Context) (*mockClient, error) {
		id := atomic.AddInt32(&clientIDCounter, 1)
		return &mockClient{id: int(id), isHealthy: healthy}, nil
	}
}

// Helper to create an empty pool with buffered capacity.
func newEmptyBufferedPool(t *testing.T, factory func(context.Context) (*mockClient, error), maxIdle, maxActive int) Pool[*mockClient] {
	// initialSize=0 ensures it starts empty but has buffer capacity maxIdle.
	p, err := New(factory, 0, maxIdle, maxActive, 0, false)
	require.NoError(t, err)
	return p
}

func TestPool_New(t *testing.T) {
	t.Parallel()
	t.Run("valid config", func(t *testing.T) {
		t.Parallel()
		// initial=1, maxIdle=1, maxActive=5
		p, err := New(newMockClientFactory(true), 1, 1, 5, 100, false)
		require.NoError(t, err)
		assert.NotNil(t, p)
		assert.Equal(t, 1, p.Len())
		err = p.Close()
		require.NoError(t, err)
	})
	t.Run("invalid config", func(t *testing.T) {
		t.Parallel()
		testCases := []struct {
			name    string
			initial int
			min     int // maxIdle
			max     int // maxActive
			wantErr string
		}{
			{
				name:    "negative initial",
				initial: -1,
				min:     5,
				max:     5,
				wantErr: "initialSize must be between 0 and maxIdleSize",
			},
			{
				name:    "initial > maxIdle",
				initial: 6,
				min:     5,
				max:     5,
				wantErr: "initialSize must be between 0 and maxIdleSize",
			},
			{
				name:    "negative maxIdle",
				initial: 0,
				min:     -1,
				max:     5,
				wantErr: "invalid maxIdleSize/maxSize configuration",
			},
			{
				name:    "maxIdle > maxActive",
				initial: 0,
				min:     6,
				max:     5,
				wantErr: "invalid maxIdleSize/maxSize configuration",
			},
			{
				name:    "zero maxActive",
				initial: 0,
				min:     0,
				max:     0,
				wantErr: "maxSize must be positive",
			},
			{
				name:    "negative maxActive",
				initial: 0,
				min:     0,
				max:     -1,
				wantErr: "maxSize must be positive",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				_, err := New(newMockClientFactory(true), tc.initial, tc.min, tc.max, 0, false)
				assert.ErrorContains(t, err, tc.wantErr)
			})
		}
	})
}

func TestPool_GetPut(t *testing.T) {
	t.Parallel()
	// initial=1
	p, err := New(newMockClientFactory(true), 1, 1, 2, 100, false)
	require.NoError(t, err)
	defer func() {
		err := p.Close()
		require.NoError(t, err)
	}()
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
	t.Parallel()
	// Factory creates one unhealthy client, then healthy ones.
	var createdCount int32
	factory := func(_ context.Context) (*mockClient, error) {
		count := atomic.AddInt32(&createdCount, 1)
		// First client is unhealthy, subsequent are healthy.
		return &mockClient{isHealthy: count > 1}, nil
	}
	// Create pool with initial=1
	p, err := New(factory, 1, 1, 2, 100, false)
	require.NoError(t, err)
	defer func() {
		err := p.Close()
		require.NoError(t, err)
	}()
	// Get should discard the initial unhealthy client and create a new, healthy one.
	c, err := p.Get(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, c)
}

func TestPool_Put_Unhealthy(t *testing.T) {
	t.Parallel()
	// Needs buffered pool to store returned client.
	p := newEmptyBufferedPool(t, newMockClientFactory(true), 2, 2)
	defer func() {
		err := p.Close()
		require.NoError(t, err)
	}()
	// Get a client. This will acquire a semaphore.
	c, err := p.Get(context.Background())
	require.NoError(t, err)
	// Make it unhealthy.
	c.isHealthy = false
	// Put it back.
	p.Put(c)
	// Pool should store the client, even if unhealthy, because Put doesn't check health anymore.
	assert.Equal(t, 1, p.Len())
	// Client is NOT closed yet. It will be closed when retrieved by Get.
	assert.False(t, c.isClosed)

	// Verify Get closes it
	c2, err := p.Get(context.Background())
	require.NoError(t, err)
	// The bad client should have been closed during this Get call
	assert.True(t, c.isClosed)
	// And we should have received a new healthy client
	assert.NotEqual(t, c, c2)
	assert.True(t, c2.isHealthy)
}

func TestPool_Full(t *testing.T) {
	t.Parallel()
	// Max idle 1, Max active 1. Empty initially.
	p := newEmptyBufferedPool(t, newMockClientFactory(true), 1, 1)
	defer func() {
		err := p.Close()
		require.NoError(t, err)
	}()
	// Get one client, pool should now be at max capacity
	c1, err := p.Get(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, c1)
	// Try to get another, should be full
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, err = p.Get(ctx)
	assert.ErrorIs(t, err, context.DeadlineExceeded, "Getting from a full pool should block until context is cancelled")
}

func TestPool_Close(t *testing.T) {
	t.Parallel()
	client := &mockClient{isHealthy: true}
	factory := func(_ context.Context) (*mockClient, error) {
		return client, nil
	}
	// Initial=1
	p, err := New(factory, 1, 1, 1, 100, false)
	require.NoError(t, err)
	err = p.Close()
	require.NoError(t, err)
	assert.True(t, client.isClosed)
}

func TestManager(t *testing.T) {
	t.Parallel()
	m := NewManager()
	p, err := New(newMockClientFactory(true), 1, 1, 5, 100, false)
	require.NoError(t, err)
	m.Register("test_pool", p)
	retrievedPool, ok := Get[*mockClient](m, "test_pool")
	require.True(t, ok)
	assert.Equal(t, p, retrievedPool)
	_, ok = Get[*mockClient](m, "nonexistent_pool")
	assert.False(t, ok)
	m.CloseAll()
}

type mockCloser struct {
	closed bool
}

func (m *mockCloser) Close() error {
	m.closed = true
	return nil
}

func TestManager_RegisterOverwriteClosesOldPoolWithCloser(t *testing.T) {
	m := NewManager()
	closer1 := &mockCloser{}
	closer2 := &mockCloser{}
	m.Register("test_closer", closer1)
	m.Register("test_closer", closer2)
	assert.True(t, closer1.closed, "Expected old closer to be closed upon re-registration")
	assert.False(t, closer2.closed, "Expected new closer to not be closed")
}

type simpleMockPool struct {
	closed bool
}

func (p *simpleMockPool) Close() error {
	p.closed = true
	return nil
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
	// Use buffered pool
	p := newEmptyBufferedPool(t, newMockClientFactory(true), maxSize, maxSize)

	defer func() {
		err := p.Close()
		require.NoError(t, err)
	}()
	// Cycle through clients, marking them as unhealthy before returning them.
	// This should not exhaust the pool's semaphore.
	for i := 0; i < maxSize+1; i++ {
		// Get a client. This acquires a semaphore permit.
		c, err := p.Get(context.Background())
		require.NoError(t, err, "Pool should not be exhausted on iteration %d", i)
		require.NotNil(t, c)

		// If we retrieved an unhealthy client (from previous iteration), Get should have closed it.
		// The current 'c' should be a new, healthy one (because factory makes healthy ones).
		assert.True(t, c.isHealthy)

		// Mark it as unhealthy for the next iteration.
		c.isHealthy = false

		// Put it back.
		p.Put(c)

		// Client should NOT be closed yet.
		assert.False(t, c.isClosed)
		// The pool should store the unhealthy client.
		assert.Equal(t, 1, p.Len())
	}
}

func TestPool_PutOnClosedPool(t *testing.T) {
	const maxSize = 1
	p := newEmptyBufferedPool(t, newMockClientFactory(true), maxSize, maxSize)

	// Get the only client
	c, err := p.Get(context.Background())
	require.NoError(t, err)
	require.NotNil(t, c)
	// Close the pool while the client is checked out
	err = p.Close()
	require.NoError(t, err)
	// Now, put the client back into the closed pool.
	// This is where the semaphore leak occurs in the original code.
	p.Put(c)
	// The client should be closed because the pool is closed.
	assert.True(t, c.isClosed)
	// To verify the fix, we check if we can acquire the permit again.
	// In the buggy version, the permit is never released.
	poolImpl := p.(*poolImpl[*mockClient])
	success := poolImpl.tryAcquire(1)
	require.True(t, success, "Should be able to acquire permit after returning a client to a closed pool")
	if success {
		poolImpl.release(1)
	}
}

func TestPool_ConcurrentGetPut(t *testing.T) {
	const (
		maxSize    = 20
		numClients = 100
		iterations = 200
	)
	// Using New() with initial=10
	p, err := New(newMockClientFactory(true), 10, 10, maxSize, 100, false)
	require.NoError(t, err)
	defer func() {
		err := p.Close()
		require.NoError(t, err)
	}()
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
					time.Sleep(time.Duration(secureRandomInt(20)) * time.Millisecond)
					continue
				}
				require.NotNil(t, client, "Goroutine %d received a nil client", goroutineID)
				// Simulate some work with random duration
				time.Sleep(time.Duration(secureRandomInt(10)) * time.Millisecond)
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
			p, err := New(newMockClientFactory(true), 1, 1, 2, 100, false)
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

func TestManager_Deregister(t *testing.T) {
	m := NewManager()
	p, err := New(newMockClientFactory(true), 1, 1, 5, 100, false)
	require.NoError(t, err)
	m.Register("test_pool", p)

	retrievedPool, ok := Get[*mockClient](m, "test_pool")
	require.True(t, ok)
	assert.NotNil(t, retrievedPool)

	m.Deregister("test_pool")
	_, ok = Get[*mockClient](m, "test_pool")
	assert.False(t, ok, "Pool should have been deregistered")
}

func TestPool_New_FactoryError(t *testing.T) {
	factory := func(_ context.Context) (*mockClient, error) {
		return nil, fmt.Errorf("factory error")
	}
	// initial=1
	_, err := New(factory, 1, 1, 1, 0, false)
	assert.Error(t, err)
}

func TestPool_New_DisableHealthCheck_WithUnhealthyClient(t *testing.T) {
	factory := newMockClientFactory(false) // Creates unhealthy clients
	p, err := New(factory, 1, 1, 1, 0, true)  // disableHealthCheck = true
	require.NoError(t, err)
	defer func() {
		err := p.Close()
		require.NoError(t, err)
	}()

	// The pool should contain the unhealthy client.
	assert.Equal(t, 1, p.Len())

	// Getting the client should not fail, even though it's unhealthy.
	c, err := p.Get(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, c)
	assert.False(t, c.IsHealthy(context.Background()))
}

func TestPool_Get_FactoryError(t *testing.T) {
	var callCount int32
	factory := func(_ context.Context) (*mockClient, error) {
		if atomic.AddInt32(&callCount, 1) > 1 {
			return nil, fmt.Errorf("factory error")
		}
		return &mockClient{isHealthy: true}, nil
	}

	// initial=1
	p, err := New(factory, 1, 1, 2, 0, false)
	require.NoError(t, err)
	defer func() {
		err := p.Close()
		require.NoError(t, err)
	}()

	// Get the first client, which should succeed.
	c1, err := p.Get(context.Background())
	require.NoError(t, err)
	require.NotNil(t, c1)

	// Try to get a second client. This should call the factory and fail.
	_, err = p.Get(context.Background())
	assert.Error(t, err)
}

func TestPool_DisableHealthCheck(t *testing.T) {
	factory := newMockClientFactory(false) // Creates unhealthy clients
	p, err := New(factory, 1, 1, 1, 0, true)  // disableHealthCheck = true
	require.NoError(t, err)
	defer func() {
		err := p.Close()
		require.NoError(t, err)
	}()

	// Get a client. Since health checks are disabled, it should return the unhealthy one.
	c, err := p.Get(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, c)
	assert.False(t, c.IsHealthy(context.Background()))

	// Put it back and get it again, should be the same one.
	p.Put(c)
	c2, err := p.Get(context.Background())
	require.NoError(t, err)
	assert.Same(t, c, c2)
}

func TestPool_GetWithAlreadyCanceledContext(t *testing.T) {
	p := newEmptyBufferedPool(t, newMockClientFactory(true), 1, 1)
	defer func() {
		err := p.Close()
		require.NoError(t, err)
	}()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := p.Get(ctx)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestPool_PutToFullPool(t *testing.T) {
	p := newEmptyBufferedPool(t, newMockClientFactory(true), 1, 1)
	defer func() {
		err := p.Close()
		require.NoError(t, err)
	}()

	// Manually inject a client to make it full (simulate active client returning)
	// We need to acquire permit first.
	// Since we used newEmptyBufferedPool, activeCount=0.

	c1, err := p.Get(context.Background()) // active=1
	require.NoError(t, err)
	p.Put(c1) // active=1, chan=1. Full.

	// The pool is now full with one idle client.
	require.Equal(t, 1, p.Len())

	// Create an extra client and try to put it in the full pool.
	extraClient := &mockClient{isHealthy: true}
	p.Put(extraClient)

	// The extra client should be closed, and the pool size should not increase.
	assert.True(t, extraClient.isClosed)
	assert.Equal(t, 1, p.Len())
}

type anotherMockClient struct {
	mockClient
}

func TestManager_Get_TypeSafety(t *testing.T) {
	m := NewManager()
	p, err := New(newMockClientFactory(true), 1, 1, 1, 0, false)
	require.NoError(t, err)

	m.Register("test_pool", p)

	// Correct type, should succeed.
	_, ok := Get[*mockClient](m, "test_pool")
	assert.True(t, ok)

	// Incorrect type, should fail.
	_, ok = Get[*anotherMockClient](m, "test_pool")
	assert.False(t, ok)
}

func TestPool_GetPrefersIdleClientsOverCreatingNew(t *testing.T) {
	const maxSize = 2
	var factoryCallCount int32
	factory := func(_ context.Context) (*mockClient, error) {
		atomic.AddInt32(&factoryCallCount, 1)
		return &mockClient{isHealthy: true}, nil
	}

	p := newEmptyBufferedPool(t, factory, maxSize, maxSize)
	defer func() {
		err := p.Close()
		require.NoError(t, err)
	}()

	// Reset counter.
	atomic.StoreInt32(&factoryCallCount, 0)

	// 1. Fill the pool up to its max size.
	clients := make([]*mockClient, maxSize)
	for i := 0; i < maxSize; i++ {
		c, err := p.Get(context.Background())
		require.NoError(t, err)
		clients[i] = c
	}
	require.Equal(t, int32(maxSize), atomic.LoadInt32(&factoryCallCount), "Factory should be called max_size times to fill the pool")

	// 2. Return one client to the pool, making it idle.
	p.Put(clients[0])
	require.Equal(t, 1, p.Len(), "Pool should have one idle client")

	// 3. Request a client. It should reuse the idle one, not create a new one.
	reusedClient, err := p.Get(context.Background())
	require.NoError(t, err)
	require.Equal(t, int32(maxSize), atomic.LoadInt32(&factoryCallCount), "Factory should not be called again when an idle client is available")
	assert.Same(t, clients[0], reusedClient, "Should have reused the same client instance")

	// 4. Clean up
	p.Put(reusedClient)
	p.Put(clients[1])
}

func TestPool_ConcurrentGetAndClose(t *testing.T) {
	const (
		maxSize    = 20
		numGetters = 50
	)

	p, err := New(newMockClientFactory(true), 5, 5, maxSize, 100, false)
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(numGetters)

	// Start a goroutine to close the pool after a short delay.
	go func() {
		time.Sleep(10 * time.Millisecond)
		err := p.Close()
		if err != nil {
			// Ignore error in async close during test if it's already closed
		}
	}()

	for i := 0; i < numGetters; i++ {
		go func() {
			defer wg.Done()
			client, err := p.Get(context.Background())
			if err == ErrPoolClosed {
				return
			}
			if err != nil {
				// Under high contention, it's possible to get a context deadline exceeded error
				// if the test machine is slow. This is acceptable.
				return
			}
			time.Sleep(time.Duration(secureRandomInt(20)) * time.Millisecond)
			p.Put(client)
		}()
	}

	wg.Wait()

	_, err = p.Get(context.Background())
	assert.Equal(t, ErrPoolClosed, err, "Getting from a closed pool should return ErrPoolClosed")
}

func TestPool_PutNilClientReleasesSemaphore(t *testing.T) {
	const maxSize = 1
	p := newEmptyBufferedPool(t, newMockClientFactory(true), maxSize, maxSize)
	defer func() {
		err := p.Close()
		require.NoError(t, err)
	}()

	// Get the only client, acquiring the only permit.
	client, err := p.Get(context.Background())
	require.NoError(t, err)
	require.NotNil(t, client)

	// Put a nil client back. The bug is that this does not release the permit.
	p.Put(nil)

	// Try to get a client again. With the bug, this will hang until timeout
	// because the permit was never released. With the fix, it should succeed.
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = p.Get(ctx)
	assert.NoError(t, err, "Should be able to get a client after putting nil, but it timed out, indicating a semaphore leak.")
}

func TestPool_ConcurrentClose(_ *testing.T) {
	pool, _ := New(newMockClientFactory(true), 0, 0, 10, 0, false)
	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			_ = pool.Close()
		}()
	}
	wg.Wait()
}

func TestPool_GetWithUnhealthyClients(t *testing.T) {
	t.Parallel()
	maxSize := 5
	factory := newMockClientFactory(true)
	// New() pre-fills. We can use that.
	pool, err := New(factory, maxSize, maxSize, maxSize, 0, false)
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
	}()
	wg.Wait()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatal("Test timed out, which indicates a likely deadlock due to semaphore leak.")
	}
}

func TestPool_GetRetriesWhenClientIsNil(t *testing.T) {
	p := newEmptyBufferedPool(t, newMockClientFactory(true), 1, 1)
	defer func() {
		err := p.Close()
		require.NoError(t, err)
	}()

	poolImpl := p.(*poolImpl[*mockClient])
	poolImpl.clients <- poolItem[*mockClient]{client: nil, retry: true}

	// Now, try to get a client. It should NOT return error, but retry and create a new one.
	c, err := p.Get(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, c)
}

func TestPool_Get_RaceWithClose(t *testing.T) {
	const maxSize = 1
	factoryStarted := make(chan struct{})
	factoryProceed := make(chan struct{})

	// Initial factory for pre-fill
	p := newEmptyBufferedPool(t, newMockClientFactory(true), maxSize, maxSize)

	// Blocking factory for the test
	factory := func(_ context.Context) (*mockClient, error) {
		close(factoryStarted) // Signal that the factory has been entered
		<-factoryProceed      // Wait for the signal to proceed
		return &mockClient{isHealthy: true}, nil
	}
	p.(*poolImpl[*mockClient]).factory = factory

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		// This will acquire the semaphore and then block in the factory
		_, _ = p.Get(context.Background())
	}()

	// Wait for the Get goroutine to enter the factory
	<-factoryStarted

	// At this point, the semaphore is acquired, but the client is not created yet.
	// Now, close the pool.
	err := p.Close()
	require.NoError(t, err)

	// Allow the factory to proceed. The Get call should now fail because the pool is closed.
	close(factoryProceed)

	// Wait for the Get goroutine to finish
	wg.Wait()

	// To verify the fix, we check if we can acquire the permit.
	poolImpl := p.(*poolImpl[*mockClient])
	success := poolImpl.tryAcquire(1)
	require.True(t, success, "Should be able to acquire permit")
	if success {
		poolImpl.release(1)
	}
}

func TestPool_Get_RaceWithPut(t *testing.T) {
	const maxSize = 1
	var factoryCallCount int32
	factoryWillBlock := make(chan struct{})
	factoryProceed := make(chan struct{})

	// Initial factory
	p := newEmptyBufferedPool(t, newMockClientFactory(true), maxSize, maxSize)
	defer func() {
		err := p.Close()
		require.NoError(t, err)
	}()

	// Blocking factory
	factory := func(_ context.Context) (*mockClient, error) {
		atomic.AddInt32(&factoryCallCount, 1)
		close(factoryWillBlock) // Signal that we are in the factory
		<-factoryProceed        // Wait for the test to allow us to proceed
		return &mockClient{isHealthy: true}, nil
	}
	p.(*poolImpl[*mockClient]).factory = factory

	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine 1: Tries to get a client, will block in the factory.
	go func() {
		defer wg.Done()
		_, _ = p.Get(context.Background())
	}()

	// Wait for Goroutine 1 to enter the factory, ensuring it has acquired the only semaphore permit.
	<-factoryWillBlock

	// Goroutine 2: Tries to get a client, but the pool is full, so it will wait.
	getCtx, cancelGet := context.WithCancel(context.Background())
	go func() {
		defer wg.Done()
		_, _ = p.Get(getCtx)
	}()

	// Give Goroutine 2 a moment to start waiting.
	time.Sleep(10 * time.Millisecond)

	// Cancel the context for Goroutine 2.
	cancelGet()

	// Now, create a new client and put it into the pool. This happens while Goroutine 2 is waiting.
	p.Put(&mockClient{isHealthy: true})

	// Allow the factory in Goroutine 1 to proceed.
	close(factoryProceed)

	// Wait for all goroutines to finish.
	wg.Wait()

	// Assert that the factory was only called once, and the pool has one idle client.
	assert.Equal(t, int32(1), atomic.LoadInt32(&factoryCallCount), "Factory should have been called only once")
	assert.Equal(t, 1, p.Len(), "Pool should have one idle client")
}

func TestPool_Get_Full_ContextCanceled(t *testing.T) {
	p := newEmptyBufferedPool(t, newMockClientFactory(true), 1, 1)
	defer func() {
		err := p.Close()
		require.NoError(t, err)
	}()

	// Get the only client to make the pool full.
	c, err := p.Get(context.Background())
	require.NoError(t, err)
	require.NotNil(t, c)

	// Now try to get another one, with a context that will be canceled.
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := p.Get(ctx)
		assert.ErrorIs(t, err, context.Canceled)
	}()

	// Give the goroutine a moment to start waiting, then cancel.
	time.Sleep(10 * time.Millisecond)
	cancel()
	wg.Wait()

	// Return the client to the pool.
	p.Put(c)
}

func TestPool_Get_WaitsForClientThenGetsUnhealthy(t *testing.T) {
	p := newEmptyBufferedPool(t, newMockClientFactory(true), 1, 1)
	defer func() {
		err := p.Close()
		require.NoError(t, err)
	}()

	// Goroutine 1: Tries to get a client, will wait because the pool is empty.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		c, err := p.Get(context.Background())
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.True(t, c.IsHealthy(context.Background()))
	}()

	// Give Goroutine 1 a moment to start waiting.
	time.Sleep(10 * time.Millisecond)

	// Goroutine 2: Puts an unhealthy client into the pool.
	unhealthyClient := &mockClient{isHealthy: false}
	p.Put(unhealthyClient)

	// Goroutine 1 should discard the unhealthy client and create a new healthy one.
	wg.Wait()
}

func TestManager_ConcurrentAccess(t *testing.T) {
	m := NewManager()
	var wg sync.WaitGroup
	numGoroutines := 50
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(_ int) {
			defer wg.Done()
			poolName := "test_pool"
			// Concurrently register and get pools
			p, err := New(newMockClientFactory(true), 1, 1, 2, 100, false)
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
func secureRandomInt(maxVal int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(maxVal)))
	if err != nil {
		// Fallback or panic in test
		return 0
	}
	return int(n.Int64())
}

func TestPool_Starvation(t *testing.T) {
	// Max size 1.
	p := newEmptyBufferedPool(t, newMockClientFactory(true), 1, 1)
	defer func() { _ = p.Close() }()

	// 1. Get client C1.
	c1, err := p.Get(context.Background())
	require.NoError(t, err)

	// 2. Start goroutine to Get C2. It should block.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Use timeout to prevent indefinite hang if bug exists
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		c2, err := p.Get(ctx)
		// If starvation occurs, we get DeadlineExceeded
		assert.NoError(t, err, "Get timed out, starvation occurred")
		assert.NotNil(t, c2)
	}()

	// Sleep to let C2 block
	time.Sleep(50 * time.Millisecond)

	// 3. Mark C1 unhealthy and Put it.
	// This releases permit, but p.clients is empty.
	c1.mu.Lock()
	c1.isHealthy = false
	c1.mu.Unlock()
	p.Put(c1)

	// 4. Wait for C2.
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(3 * time.Second):
		t.Fatal("Starvation detected: Get() did not return")
	}
}

func TestPool_New_DisableHealthCheck_FactoryError(t *testing.T) {
	factory := func(_ context.Context) (*mockClient, error) {
		return nil, fmt.Errorf("factory error")
	}
	_, err := New(factory, 1, 1, 1, 0, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "factory failed to create initial client")
}

type errorCloser struct{}

func (e *errorCloser) Close() error {
	return fmt.Errorf("close error")
}

func TestManager_RegisterOverwrite_CloseError(t *testing.T) {
	m := NewManager()
	pool1 := &errorCloser{}
	pool2 := &simpleMockPool{}
	m.Register("test_pool", pool1)

	// This should trigger the warning log, but not panic
	m.Register("test_pool", pool2)

	// Verify the new pool is registered
	assert.Equal(t, pool2, m.pools["test_pool"])
}

func TestPool_New_DisableHealthCheck_NilClient(t *testing.T) {
	factory := func(_ context.Context) (*mockClient, error) {
		return nil, nil // Return nil client with no error
	}
	_, err := New(factory, 1, 1, 1, 0, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "factory returned nil client")
}
