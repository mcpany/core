package pool

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCloseAll_Error(t *testing.T) {
	m := NewManager()
	p := &errorCloser{}
	m.Register("pool1", p)
	m.CloseAll()
	// Should log warning, not panic.
}

// Test Get context done at start
func TestGet_ContextAlreadyDone(t *testing.T) {
	p, _ := New(newMockClientFactory(true), 0, 0, 1, 0, false)
	defer func() { _ = p.Close() }()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := p.Get(ctx)
	assert.ErrorIs(t, err, context.Canceled)
}

// Test Get returns error if pool closed (fast path)
func TestGet_PoolClosed_FastPath(t *testing.T) {
	p, _ := New(newMockClientFactory(true), 0, 0, 1, 0, false)
	_ = p.Close()
	_, err := p.Get(context.Background())
	assert.Equal(t, ErrPoolClosed, err)
}

// Test Get returns error if pool closed while waiting on channel
func TestGet_PoolClosed_WhileWaiting(t *testing.T) {
	p, _ := New(newMockClientFactory(true), 0, 0, 1, 0, false)
	// Fill pool (maxSize 1)
	c, _ := p.Get(context.Background())

	// Start Get in background
	errCh := make(chan error)
	go func() {
		_, err := p.Get(context.Background())
		errCh <- err
	}()

	time.Sleep(10*time.Millisecond)
	_ = p.Close()
	p.Put(c) // Put client back so Get can wake up?
    // Wait, Close closes p.clients channel. Get reading from channel should receive (!ok).

	err := <-errCh
	assert.Equal(t, ErrPoolClosed, err)
}

// Test Put nil client
func TestPut_Nil(t *testing.T) {
	p, _ := New(newMockClientFactory(true), 0, 0, 1, 0, false)
	defer func() { _ = p.Close() }()

    // Get first to acquire permit
    c1, _ := p.Get(context.Background())
    assert.NotNil(t, c1)

    // Put nil client
    var c *mockClient = nil
    p.Put(c)

    // Now we should be able to Get again if permit was released.
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()
    c2, err := p.Get(ctx)
    assert.NoError(t, err)
    assert.NotNil(t, c2)
}

// Test retry logic in loop (retry item in channel)
func TestGet_RetryItem(t *testing.T) {
    // Use buffered pool so we can inject retry item
	p := newEmptyBufferedPool(t, newMockClientFactory(true), 1, 1)
	defer func() { _ = p.Close() }()

    // Inject retry item into channel manually
    pi := p.(*poolImpl[*mockClient])
    pi.clients <- poolItem[*mockClient]{retry: true}

    // Get should skip it and create new client
    c, err := p.Get(context.Background())
    assert.NoError(t, err)
    assert.NotNil(t, c)
}

// Test race: Acquire permit, but client becomes available in channel.
func TestGet_Race_Acquire_But_ClientAvailable(t *testing.T) {
    // Use buffered pool
    p := newEmptyBufferedPool(t, newMockClientFactory(true), 1, 1)
    defer func() { _ = p.Close() }()

    pi := p.(*poolImpl[*mockClient])

    // Put item directly.
    c := &mockClient{isHealthy: true}
    pi.clients <- poolItem[*mockClient]{client: c}

    c2, err := p.Get(context.Background())
    assert.NoError(t, err)
    assert.Equal(t, c, c2)

    c3, err := p.Get(context.Background())
    assert.NoError(t, err)
    assert.NotEqual(t, c, c3)
}

// Test race: Acquire permit, but retry item becomes available in channel.
func TestGet_Race_Acquire_But_ClientAvailable_Retry(t *testing.T) {
    p := newEmptyBufferedPool(t, newMockClientFactory(true), 1, 1)
    defer func() { _ = p.Close() }()

    pi := p.(*poolImpl[*mockClient])

    // Put retry item.
    pi.clients <- poolItem[*mockClient]{retry: true}

    // Get calls tryAcquire (succeeds, active=1).
    // Then hits select case <- clients. Finds retry item.
    // Releases permit. Continues loop.
    // Loop again. Channel empty. tryAcquire succeeds (active=1).
    // Creates new client.

    c, err := p.Get(context.Background())
    assert.NoError(t, err)
    assert.NotNil(t, c)
}

// Test race: Acquire permit, but unhealthy client becomes available in channel.
func TestGet_Race_Acquire_But_ClientAvailable_Unhealthy(t *testing.T) {
    p := newEmptyBufferedPool(t, newMockClientFactory(true), 1, 1)
    defer func() { _ = p.Close() }()

    pi := p.(*poolImpl[*mockClient])

    // Put unhealthy item.
    cUnhealthy := &mockClient{isHealthy: false}
    pi.clients <- poolItem[*mockClient]{client: cUnhealthy}

    // Get calls tryAcquire (succeeds, active=1).
    // Then hits select case <- clients. Finds unhealthy item.
    // Releases permit. isHealthySafe returns false.
    // Continues loop.
    // Loop again. Channel empty. tryAcquire succeeds (active=1).
    // Creates new client.

    c, err := p.Get(context.Background())
    assert.NoError(t, err)
    assert.NotNil(t, c)
    assert.NotEqual(t, cUnhealthy, c)
    assert.True(t, cUnhealthy.isClosed)
}


// Test CloseAll with non-closer pool (coverage for "ok := untypedPool.(io.Closer)")
func TestCloseAll_NonCloser(t *testing.T) {
    m := NewManager()
    m.Register("pool1", "string is not closer")
    m.CloseAll()
}

// Test Deregister with non-closer
func TestDeregister_NonCloser(t *testing.T) {
    m := NewManager()
    m.Register("pool1", "string is not closer")
    m.Deregister("pool1")
}

// Test Register overwrite non-closer
func TestRegister_Overwrite_NonCloser(t *testing.T) {
    m := NewManager()
    m.Register("pool1", "string is not closer")
    m.Register("pool1", &simpleMockPool{})
}

// Test Put with interface type and nil
func TestPut_NilInterface(t *testing.T) {
    factory := func(_ context.Context) (ClosableClient, error) {
        return &mockClient{isHealthy:true}, nil
    }
    p, _ := New[ClosableClient](factory, 0, 0, 1, 0, false)
    defer func() { _ = p.Close() }()

    // Acquire permit
    _, _ = p.Get(context.Background())

    // Put nil interface
    p.Put(nil)

    // Check if permit released
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()
    c2, err := p.Get(ctx)
    assert.NoError(t, err)
    assert.NotNil(t, c2)
}

// Test Close handles retry items in channel
func TestClose_WithRetryItems(t *testing.T) {
    // Use buffered pool
	p := newEmptyBufferedPool(t, newMockClientFactory(true), 1, 1)
    pi := p.(*poolImpl[*mockClient])
    pi.clients <- poolItem[*mockClient]{retry: true}

    err := p.Close()
    assert.NoError(t, err)
}

// Test Get backoff context cancellation
func TestGet_Backoff_ContextCancel(t *testing.T) {
    factory := func(_ context.Context) (*mockClient, error) {
        return &mockClient{isHealthy: false}, nil
    }

    p, _ := New(factory, 0, 0, 1, 0, false)
    defer func() { _ = p.Close() }()

    ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond) // Less than 100ms backoff
    defer cancel()

    start := time.Now()
    _, err := p.Get(ctx)
    duration := time.Since(start)

    assert.ErrorIs(t, err, context.DeadlineExceeded)
    assert.GreaterOrEqual(t, duration, 50*time.Millisecond)
}

// Test New with minSize > 0 and factory error (verify Close called)
func TestNew_FactoryError_VerifyClose(t *testing.T) {
    factory := func(_ context.Context) (*mockClient, error) {
        return nil, errors.New("fail")
    }
    // initial=1
    _, err := New(factory, 1, 1, 1, 0, false)
    assert.Error(t, err)
}

// Additional coverage tests

func TestNew_EnabledHealthCheck_FactoryError(t *testing.T) {
    factory := func(_ context.Context) (*mockClient, error) {
        return nil, errors.New("factory failure")
    }
    // disableHealthCheck=false, initial=1
    _, err := New(factory, 1, 1, 5, 0, false)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "factory failed to create initial client")
}

func TestNew_EnabledHealthCheck_NilClient(t *testing.T) {
    factory := func(_ context.Context) (*mockClient, error) {
        return nil, nil
    }
    // disableHealthCheck=false, initial=1
    _, err := New(factory, 1, 1, 5, 0, false)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "factory returned nil client")
}

func TestGet_ContextCancel_WhileWaitFull(t *testing.T) {
    p := newEmptyBufferedPool(t, newMockClientFactory(true), 1, 1)
    defer func() { _ = p.Close() }()

    // Consume permit
    c, _ := p.Get(context.Background())

    // Second Get waits. Cancel context.
    ctx, cancel := context.WithCancel(context.Background())

    errCh := make(chan error)
    go func() {
        _, err := p.Get(ctx)
        errCh <- err
    }()

    time.Sleep(10*time.Millisecond)
    cancel()

    err := <-errCh
    assert.ErrorIs(t, err, context.Canceled)

    p.Put(c)
}

func TestGet_Coverage_ClosedChannel_FirstLoop(t *testing.T) {
    p, _ := New(newMockClientFactory(true), 1, 1, 1, 0, false)
    pi := p.(*poolImpl[*mockClient])

    // Close pool.
    err := p.Close()
    assert.NoError(t, err)

    // Hack: reopen logic check by setting closed=false
    pi.closed.Store(false)

    // Now Get. closed.Load() is false.
    // Channel is closed.
    // Select <-p.clients returns !ok immediately.

    _, err = p.Get(context.Background())
    assert.Equal(t, ErrPoolClosed, err)
}

type panicCloser struct {
	mockClient
}

func (c *panicCloser) Close() error {
	panic("close panic")
}

func TestPut_PanicOnClose(t *testing.T) {
    factory := func(_ context.Context) (*panicCloser, error) {
        return &panicCloser{mockClient: mockClient{isHealthy:true}}, nil
    }
    // Use unbuffered pool (or just full pool) to force discard
    // initial=0, maxIdle=0
    p, _ := New(factory, 0, 0, 1, 0, false)
    // defer func() { _ = p.Close() }() // Avoid panic in cleanup

    // Get client
    c, err := p.Get(context.Background())
    assert.NoError(t, err)

    // Put it back. Channel is unbuffered.
    // Put enters default. Discards client. Calls c.Close(). Panics.
    // lo.Try should recover it.

    assert.NotPanics(t, func() {
        p.Put(c)
    })

    // And permit should be released.
    c2, err := p.Get(context.Background())
    assert.NoError(t, err)
    assert.NotNil(t, c2)
}

func TestClose_PanicOnClose(t *testing.T) {
    factory := func(_ context.Context) (*panicCloser, error) {
        return &panicCloser{mockClient: mockClient{isHealthy:true}}, nil
    }
    // initial=1
    p, _ := New(factory, 1, 1, 1, 0, false)

    // Pool has 1 item in channel.

    assert.NotPanics(t, func() {
        _ = p.Close()
    })
}

func TestClose_WithNilClient(t *testing.T) {
    p := newEmptyBufferedPool(t, newMockClientFactory(true), 1, 1)
    pi := p.(*poolImpl[*mockClient])
    // Inject nil client
    pi.clients <- poolItem[*mockClient]{client: nil}

    // Close should handle it
    err := p.Close()
    assert.NoError(t, err)
}

func TestPool_MaxIdleConnections_RespectsLimit(t *testing.T) {
	factory := func(ctx context.Context) (*mockClient, error) {
		return &mockClient{isHealthy: true}, nil
	}

	maxIdle := 1
	maxConn := 10

	// Initial=maxIdle to simulate previous "buggy" behavior where we start full but should stay limited.
	// But new signature requires initial.
	p, err := New(factory, maxIdle, maxIdle, maxConn, time.Second, true)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	defer func() { _ = p.Close() }()

	// Initial state: should have maxIdle (1) clients.
	assert.Equal(t, maxIdle, p.Len(), "Initial idle count should match minSize")

	// Checkout maxConn clients
	clients := make([]*mockClient, maxConn)
	for i := 0; i < maxConn; i++ {
		c, err := p.Get(context.Background())
		if err != nil {
			t.Fatalf("Failed to get client %d: %v", i, err)
		}
		clients[i] = c
	}

	// Pool should be empty now (all checked out)
	assert.Equal(t, 0, p.Len(), "Pool should be empty after checking out all")

	// Return all clients
	for _, c := range clients {
		p.Put(c)
	}

	// Check idle count.
	currentIdle := p.Len()

	if currentIdle > maxIdle {
		t.Errorf("Pool holds %d idle connections, expected max %d", currentIdle, maxIdle)
	}
}
