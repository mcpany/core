package pool

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPoolClient struct {
	closeErr error
	health   bool
	isClosed bool
}

func (m *mockPoolClient) Close() error {
	m.isClosed = true
	return m.closeErr
}

func (m *mockPoolClient) IsHealthy(ctx context.Context) bool {
	return m.health
}

func TestPool_New_FactoryErrorInitialClientCloseErr(t *testing.T) {
	// Tests line 164-165
	var closeCalled bool
	returnErr := errors.New("factory failed to create initial client")
	factory := func(ctx context.Context) (*mockPoolClient, error) {
		return nil, returnErr
	}

	p, err := New(factory, 1, 1, 1, time.Second, false)
	_ = closeCalled
	require.Error(t, err)
	assert.Nil(t, p)
	assert.Contains(t, err.Error(), "factory failed to create initial client")
}

func TestPool_New_FactoryErrorInitialClientCloseErr_HealthCheck(t *testing.T) {
	// Tests line 185-186
	returnErr := errors.New("factory failed to create initial client")
	factory := func(ctx context.Context) (*mockPoolClient, error) {
		return nil, returnErr
	}

	// healthCheckInterval > 0 enables health checks
	p, err := New(factory, 1, 1, 1, time.Second, true)
	require.Error(t, err)
	assert.Nil(t, p)
	assert.Contains(t, err.Error(), "factory failed to create initial client")
}


func TestPool_New_FactoryNilClient(t *testing.T) {
	// Trigger the "v.IsNil()" check
	factory := func(ctx context.Context) (*mockPoolClient, error) {
		return nil, nil // Return a typed nil
	}

	p, err := New(factory, 1, 1, 1, time.Second, false)
	require.Error(t, err)
	assert.Nil(t, p)
	assert.Contains(t, err.Error(), "factory returned nil client")
}

func TestPool_Get_PoolClosedWhileAcquiringPermit(t *testing.T) {
    // Tests: if p.closed.Load() after selecting default branch
    factory := func(ctx context.Context) (*mockPoolClient, error) {
		return &mockPoolClient{health: true}, nil
	}
	p, err := New(factory, 0, 1, 1, time.Second, false)
	require.NoError(t, err)

	// Close the pool manually.
	p.Close()

	_, err = p.Get(context.Background())
	assert.ErrorIs(t, err, ErrPoolClosed)
}

func TestPool_Put_PoolClosed(t *testing.T) {
    // Tests line: if p.closed.Load() in Put
    factory := func(ctx context.Context) (*mockPoolClient, error) {
		return &mockPoolClient{health: true}, nil
	}
	p, err := New(factory, 0, 1, 1, time.Second, false)
	require.NoError(t, err)

	c, err := p.Get(context.Background())
	require.NoError(t, err)

	p.Close()

	p.Put(c) // This should hit the if p.closed.Load() block and call Close on the client
    assert.True(t, c.(*mockPoolClient).isClosed)
}

func TestPool_Get_PoolClosedWhileAcquiringPermit_NoClients(t *testing.T) {
    factory := func(ctx context.Context) (*mockPoolClient, error) {
		return &mockPoolClient{health: true}, nil
	}
	p, err := New(factory, 0, 1, 1, time.Second, false)
	require.NoError(t, err)

	impl := p.(*poolImpl[*mockPoolClient])
	impl.factory = func(ctx context.Context) (*mockPoolClient, error) {
		impl.Close()
		return &mockPoolClient{health: true}, nil
	}

	_, err = impl.Get(context.Background())
	assert.ErrorIs(t, err, ErrPoolClosed)
}

func TestPool_Put_PoolClosedWhilePutting(t *testing.T) {
    factory := func(ctx context.Context) (*mockPoolClient, error) {
		return &mockPoolClient{health: true}, nil
	}
	p, err := New(factory, 0, 1, 1, time.Second, false)
	require.NoError(t, err)

	impl := p.(*poolImpl[*mockPoolClient])
	c, err := impl.Get(context.Background())
	require.NoError(t, err)

	impl.closed.Store(true) // simulate closure right before lock
	impl.Put(c)
    assert.True(t, c.(*mockPoolClient).isClosed)
}

func TestPool_Get_FactoryError(t *testing.T) {
    factory := func(ctx context.Context) (*mockPoolClient, error) {
		return nil, errors.New("get factory err")
	}
	p, err := New(factory, 0, 1, 1, time.Second, false)
	require.NoError(t, err)

	_, err = p.Get(context.Background())
	require.Error(t, err)
    assert.Contains(t, err.Error(), "get factory err")
}

func TestPool_Get_NilClient(t *testing.T) {
    factory := func(ctx context.Context) (*mockPoolClient, error) {
		return nil, nil
	}
	p, err := New(factory, 0, 1, 1, time.Second, false)
	require.NoError(t, err)

	_, err = p.Get(context.Background())
	require.Error(t, err)
    assert.Contains(t, err.Error(), "factory returned nil client")
}
