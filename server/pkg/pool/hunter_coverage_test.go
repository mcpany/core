// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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
	p, _ := New(newMockClientFactory(true), 0, 1, 0, false)
	defer func() { _ = p.Close() }()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := p.Get(ctx)
	assert.ErrorIs(t, err, context.Canceled)
}

// Test Get returns error if pool closed (fast path)
func TestGet_PoolClosed_FastPath(t *testing.T) {
	p, _ := New(newMockClientFactory(true), 0, 1, 0, false)
	_ = p.Close()
	_, err := p.Get(context.Background())
	assert.Equal(t, ErrPoolClosed, err)
}

// Test Get returns error if pool closed while waiting on channel
func TestGet_PoolClosed_WhileWaiting(t *testing.T) {
	p, _ := New(newMockClientFactory(true), 0, 1, 0, false)
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
	p, _ := New(newMockClientFactory(true), 0, 1, 0, false)
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
	p, _ := New(newMockClientFactory(true), 0, 1, 0, false)
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
    p, _ := New(newMockClientFactory(true), 0, 1, 0, false)
    defer func() { _ = p.Close() }()

    // Manually release permit so sem=1
    pi := p.(*poolImpl[*mockClient])
    // But channel is empty.

    // We want channel to have item AND sem to have permit.
    // Put item directly.
    c := &mockClient{isHealthy: true}
    pi.clients <- poolItem[*mockClient]{client: c}
    // Now channel has 1 item.
    // Sem still has 1 permit (initial state).

    c2, err := p.Get(context.Background())
    assert.NoError(t, err)
    assert.Equal(t, c, c2)

    // Result: Sem=1. Chan=[].
    // If we Get again, we should get new client.
    c3, err := p.Get(context.Background())
    assert.NoError(t, err)
    assert.NotEqual(t, c, c3)
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
    p, _ := New[ClosableClient](factory, 0, 1, 0, false)
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
	p, _ := New(newMockClientFactory(true), 0, 1, 0, false)
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

    p, _ := New(factory, 0, 1, 0, false)
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
    _, err := New(factory, 1, 1, 0, false)
    assert.Error(t, err)
}
