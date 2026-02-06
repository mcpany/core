// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package pool provides a generic connection pool implementation.
package pool

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/samber/lo"
)

var (
	// ErrPoolClosed is returned when an operation is attempted on a closed pool.
	ErrPoolClosed = fmt.Errorf("pool has been closed")
	// ErrPoolFull is returned when the pool has reached its maximum capacity and
	// cannot create new clients.
	ErrPoolFull = fmt.Errorf("pool is full")

	// retryBackoff is the duration to wait before retrying to create a new client
	// when the upstream is unhealthy.
	retryBackoff = 100 * time.Millisecond
)

// ClosableClient defines the interface for clients that can be managed by the
// connection pool. Implementations must provide methods for closing the
// connection and checking its health.
//
// Summary: Interface for poolable clients.
type ClosableClient interface {
	// Close terminates the client's connection.
	//
	// Summary: Closes the client connection.
	//
	// Returns:
	//   - error: An error if the operation fails.
	Close() error

	// IsHealthy returns true if the client's connection is active and usable.
	//
	// Summary: Checks client health.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the check.
	//
	// Returns:
	//   - bool: True if healthy.
	IsHealthy(ctx context.Context) bool
}

// Pool defines the interface for a generic connection pool.
//
// Summary: Interface for a connection pool.
type Pool[T ClosableClient] interface {
	// Get retrieves a client from the pool.
	//
	// Summary: Acquires a client from the pool.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//
	// Returns:
	//   - T: The acquired client.
	//   - error: An error if acquisition fails.
	Get(ctx context.Context) (T, error)

	// Put returns a client to the pool.
	//
	// Summary: Returns a client to the pool.
	//
	// Parameters:
	//   - client: T. The client to return.
	Put(client T)

	// Close terminates all clients in the pool.
	//
	// Summary: Closes the pool and all clients.
	//
	// Returns:
	//   - error: An error if closure fails.
	Close() error

	// Len returns the number of idle clients currently in the pool.
	//
	// Summary: Returns the number of idle clients.
	//
	// Returns:
	//   - int: The count of idle clients.
	Len() int
}

type poolItem[T any] struct {
	client T
	retry  bool
}

// poolImpl is the internal implementation of the Pool interface.
//
// Summary: Generic pool implementation.
type poolImpl[T ClosableClient] struct {
	clients            chan poolItem[T]
	factory            func(context.Context) (T, error)
	maxSize            int64
	activeCount        atomic.Int64
	mu                 sync.RWMutex
	closed             atomic.Bool
	disableHealthCheck bool
}

// New creates a new connection pool with the specified factory and size
// constraints.
//
// Summary: Creates a new generic pool.
//
// Parameters:
//   - factory: func(context.Context) (T, error). The factory function.
//   - initialSize: int. Initial number of clients.
//   - maxIdleSize: int. Max idle clients.
//   - maxSize: int. Max total clients.
//   - idleTimeout: time.Duration. (Unused).
//   - disableHealthCheck: bool. Whether to skip health checks on creation.
//
// Returns:
//   - Pool[T]: The new pool.
//   - error: An error if configuration is invalid.
func New[T ClosableClient](
	factory func(context.Context) (T, error),
	initialSize, maxIdleSize, maxSize int,
	_ time.Duration, // idleTimeout is not used yet
	disableHealthCheck bool,
) (Pool[T], error) {
	if maxSize <= 0 {
		return nil, fmt.Errorf("maxSize must be positive")
	}
	if maxIdleSize < 0 || maxIdleSize > maxSize {
		return nil, fmt.Errorf("invalid maxIdleSize/maxSize configuration")
	}
	if initialSize < 0 || initialSize > maxIdleSize {
		return nil, fmt.Errorf("initialSize must be between 0 and maxIdleSize")
	}

	p := &poolImpl[T]{
		clients:            make(chan poolItem[T], maxIdleSize),
		factory:            factory,
		maxSize:            int64(maxSize),
		disableHealthCheck: disableHealthCheck,
	}

	// If health checks are disabled, we can pre-fill the pool without checks.
	if disableHealthCheck {
		for i := 0; i < initialSize; i++ {
			client, err := factory(context.Background())
			if err != nil {
				if closeErr := p.Close(); closeErr != nil {
					logging.GetLogger().Error("Failed to close pool after factory error", "error", closeErr)
				}
				return nil, fmt.Errorf("factory failed to create initial client: %w", err)
			}
			v := reflect.ValueOf(client)
			if (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) && v.IsNil() {
				_ = p.Close()
				return nil, fmt.Errorf("factory returned nil client")
			}
			p.clients <- poolItem[T]{client: client}
		}
		// Take permits for the initial clients. Guaranteed to succeed as initialSize <= maxIdleSize <= maxSize.
		p.activeCount.Add(int64(initialSize))
		return p, nil
	}

	// With health checks enabled, we need to ensure clients are healthy before adding.
	for i := 0; i < initialSize; i++ {
		client, err := factory(context.Background())
		if err != nil {
			if closeErr := p.Close(); closeErr != nil {
				logging.GetLogger().Error("Failed to close pool after factory error", "error", closeErr)
			}
			return nil, fmt.Errorf("factory failed to create initial client: %w", err)
		}
		v := reflect.ValueOf(client)
		if (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) && v.IsNil() {
			_ = p.Close()
			return nil, fmt.Errorf("factory returned nil client")
		}
		p.clients <- poolItem[T]{client: client}
	}
	// Take permits for the initial clients. Guaranteed to succeed as initialSize <= maxIdleSize <= maxSize.
	p.activeCount.Add(int64(initialSize))

	return p, nil
}

// tryAcquire attempts to acquire n permits.
// It returns true if successful, false if the pool would exceed maxSize.
// This uses atomic operations to avoid locking.
func (p *poolImpl[T]) tryAcquire(n int64) bool {
	for {
		cur := p.activeCount.Load()
		if cur+n > p.maxSize {
			return false
		}
		if p.activeCount.CompareAndSwap(cur, cur+n) {
			return true
		}
		// ⚡ BOLT: Yield processor to prevent potential livelock/starvation under high contention.
		// Randomized Selection from Top 5 High-Impact Targets
		runtime.Gosched()
	}
}

// release releases n permits.
// This uses atomic operations to avoid locking.
func (p *poolImpl[T]) release(n int64) {
	p.activeCount.Add(-n)
}

// Get retrieves a client from the pool.
//
// Summary: Acquires a client, creating one if necessary.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//
// Returns:
//   - T: The client.
//   - error: Error if pool closed or creation failed.
func (p *poolImpl[T]) Get(ctx context.Context) (T, error) {
	var zero T

	// Check if context is already cancelled before attempting to get a connection.
	// This is a fast path to avoid locking and other overhead.
	select {
	case <-ctx.Done():
		return zero, ctx.Err()
	default:
	}

	if p.closed.Load() {
		return zero, ErrPoolClosed
	}

	// Loop to ensure we return a healthy client
	for {
		// First, try a non-blocking retrieval of an existing client.
		select {
		case item, ok := <-p.clients:
			if !ok {
				return zero, ErrPoolClosed
			}
			if item.retry {
				continue
			}
			if p.isHealthySafe(ctx, item.client) {
				return item.client, nil
			}
			continue // Retry.
		default:
			// Channel is empty, proceed.
		}

		// If no idle client, try to create a new one without blocking.
		if p.tryAcquire(1) {
			// Acquired a permit, but double-check for a race condition where
			// a client was returned just before we acquired the permit.
			select {
			case item, ok := <-p.clients:
				// A client was available, so we use it and release the permit.
				p.release(1)
				if !ok {
					return zero, ErrPoolClosed
				}
				if item.retry {
					continue
				}
				if p.isHealthySafe(ctx, item.client) {
					return item.client, nil
				}
				continue // Retry.
			default:
				// No client, so create a new one.
				// We must check if the pool was closed *after* we acquired the permit.
				if p.closed.Load() {
					p.release(1) // Don't leak the permit
					return zero, ErrPoolClosed
				}

				client, err := p.factorySafe(ctx)
				if err != nil {
					return zero, err
				}

				// Check again if the pool was closed while we were creating a client.
				if p.closed.Load() {
					_ = lo.Try(func() error {
						return client.Close()
					})
					p.release(1)
					return zero, ErrPoolClosed
				}

				if p.isHealthySafe(ctx, client) {
					return client, nil
				}

				// Backoff before retrying to avoid busy loop when upstream is down
				select {
				case <-ctx.Done():
					return zero, ctx.Err()
				case <-time.After(retryBackoff):
				}

				continue // Retry.
			}
		}

		// Pool is full, so we must wait for a client to be returned.
		select {
		case item, ok := <-p.clients:
			if !ok {
				return zero, ErrPoolClosed
			}
			if item.retry {
				continue
			}
			if p.isHealthySafe(ctx, item.client) {
				return item.client, nil
			}
			continue // Retry.
		case <-ctx.Done():
			return zero, ctx.Err()
		}
	}
}

// factorySafe invokes the factory with panic protection.
// It releases a semaphore permit if the factory fails or panics.
func (p *poolImpl[T]) factorySafe(ctx context.Context) (T, error) {
	var client T
	var err error
	panicked := true
	func() {
		defer func() {
			if r := recover(); r != nil {
				p.release(1)
				panic(r)
			}
		}()
		client, err = p.factory(ctx)
		panicked = false
	}()

	// If not panicked, but error
	if !panicked && err != nil {
		p.release(1)
		return client, err // client is zero
	}

	return client, err
}

// isHealthySafe checks if the client is healthy with panic protection.
// It releases a semaphore permit and closes the client if the check fails or panics.
// Returns true if healthy.
func (p *poolImpl[T]) isHealthySafe(ctx context.Context, client T) bool {
	healthy := false
	panicked := true
	func() {
		defer func() {
			if r := recover(); r != nil {
				_ = lo.Try(func() error {
					return client.Close()
				})
				p.release(1)
				panic(r)
			}
		}()
		if p.disableHealthCheck || client.IsHealthy(ctx) {
			healthy = true
		}
		panicked = false
	}()

	// If we didn't panic, but unhealthy, cleanup.
	if !panicked && !healthy {
		_ = lo.Try(func() error {
			return client.Close()
		})
		p.release(1)
	}

	return healthy
}

// Put returns a client to the pool for reuse.
//
// Summary: Returns a client to the pool.
//
// Parameters:
//   - client: T. The client to return.
func (p *poolImpl[T]) Put(client T) {
	v := reflect.ValueOf(client)
	if !v.IsValid() || ((v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) && v.IsNil()) {
		p.release(1)
		return
	}

	if p.closed.Load() {
		_ = lo.Try(func() error {
			return client.Close()
		})
		p.release(1) // Release permit as the client is discarded
		return
	}

	// We do NOT check health here to avoid blocking the caller.
	// IsHealthy() can be slow (network calls), and Put is often used in defer.
	// Any unhealthy client returned here will be detected and discarded by Get()
	// when it is next retrieved.

	// Use RLock to allow concurrent Puts. We only need to ensure the pool isn't closed
	// *during* the send. Close() takes a Write lock.
	p.mu.RLock()
	if p.closed.Load() {
		p.mu.RUnlock()
		_ = lo.Try(client.Close)
		p.release(1)
		return
	}

	select {
	case p.clients <- poolItem[T]{client: client}:
		p.mu.RUnlock()
	default:
		// Idle pool is full, discard client.
		p.mu.RUnlock()
		_ = lo.Try(client.Close)
		p.release(1)
	}
}

// Close shuts down the pool, closing all idle clients.
//
// Summary: Closes the pool.
//
// Returns:
//   - error: Error if close fails (usually nil).
func (p *poolImpl[T]) Close() error {
	// We use the mutex here to ensure that we don't close the channel multiple times
	// or have races with other Close calls. Get/Put check p.closed via atomic which is fast.
	p.mu.Lock()
	if p.closed.Load() {
		p.mu.Unlock()
		return nil
	}
	p.closed.Store(true)
	close(p.clients)
	p.mu.Unlock()

	// Drain the channel and close all idle clients
	for item := range p.clients {
		if item.retry {
			continue
		}
		client := item.client
		v := reflect.ValueOf(client)
		if (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) && v.IsNil() {
			continue
		}
		_ = lo.Try(func() error {
			return client.Close()
		})
		p.release(1)
	}
	return nil
}

// Len returns the current number of idle clients in the pool.
//
// Summary: Returns idle client count.
//
// Returns:
//   - int: Idle count.
func (p *poolImpl[T]) Len() int {
	return len(p.clients)
}

// UntypedPool defines a non-generic interface for a pool.
//
// Summary: Interface for untyped pool management.
type UntypedPool interface {
	io.Closer
	// Len returns the number of idle clients currently in the pool.
	//
	// Summary: Returns idle client count.
	//
	// Returns:
	//   - int: Idle count.
	Len() int
}

// Manager provides a way to manage multiple named connection pools.
//
// Summary: Manages a collection of pools.
type Manager struct {
	pools map[string]any
	mu    sync.RWMutex
}

// NewManager creates and returns a new pool Manager.
//
// Summary: Initializes a new Pool Manager.
//
// Returns:
//   - *Manager: The initialized manager.
func NewManager() *Manager {
	return &Manager{
		pools: make(map[string]any),
	}
}

// Register adds a new pool to the manager under a given name.
//
// Summary: Registers a pool by name.
//
// Parameters:
//   - name: string. The pool name.
//   - pool: any. The pool instance.
func (m *Manager) Register(name string, pool any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if oldPool, ok := m.pools[name]; ok {
		if p, isCloser := oldPool.(io.Closer); isCloser {
			logging.GetLogger().Info("Closing old entry", "name", name)
			if err := p.Close(); err != nil {
				logging.GetLogger().Warn("Failed to close old pool", "name", name, "error", err)
			}
		}
	}
	m.pools[name] = pool
}

// Deregister closes and removes a pool from the manager.
//
// Summary: Removes a pool by name.
//
// Parameters:
//   - name: string. The pool name.
func (m *Manager) Deregister(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if oldPool, ok := m.pools[name]; ok {
		if p, isCloser := oldPool.(io.Closer); isCloser {
			logging.GetLogger().Info("Closing old entry", "name", name)
			if err := p.Close(); err != nil {
				logging.GetLogger().Warn("Failed to close pool", "name", name, "error", err)
			}
		}
		delete(m.pools, name)
	}
}

// Get retrieves a typed pool from the manager by name.
//
// Summary: Retrieves a pool by name and type.
//
// Parameters:
//   - m: *Manager. The manager.
//   - name: string. The pool name.
//
// Returns:
//   - Pool[T]: The typed pool.
//   - bool: True if found and type matches.
func Get[T ClosableClient](m *Manager, name string) (Pool[T], bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	untypedPool, ok := m.pools[name]
	if !ok {
		return nil, false
	}
	pool, ok := untypedPool.(Pool[T])
	return pool, ok
}

// CloseAll iterates through all registered pools in the manager and closes them.
//
// Summary: Closes all managed pools.
func (m *Manager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, untypedPool := range m.pools {
		logging.GetLogger().Info("Closing pool", "name", name)
		if p, ok := untypedPool.(io.Closer); ok {
			if err := p.Close(); err != nil {
				logging.GetLogger().Warn("Failed to close pool in CloseAll", "name", name, "error", err)
			}
		}
	}
}
