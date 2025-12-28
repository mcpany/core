// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package pool provides a generic connection pool implementation.
package pool

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mcpany/core/pkg/logging"
	"github.com/samber/lo"
	"golang.org/x/sync/semaphore"
)

var (
	// ErrPoolClosed is returned when an operation is attempted on a closed pool.
	ErrPoolClosed = fmt.Errorf("pool has been closed")
	// ErrPoolFull is returned when the pool has reached its maximum capacity and
	// cannot create new clients.
	ErrPoolFull = fmt.Errorf("pool is full")
)

// ClosableClient defines the interface for clients that can be managed by the
// connection pool. Implementations must provide methods for closing the
// connection and checking its health.
type ClosableClient interface {
	// Close terminates the client's connection.
	Close() error
	// IsHealthy returns true if the client's connection is active and usable.
	IsHealthy(ctx context.Context) bool
}

// Pool defines the interface for a generic connection pool. It supports getting
// and putting clients, closing the pool, and querying its size. The type
// parameter T is constrained to types that implement the ClosableClient
// interface.
type Pool[T ClosableClient] interface {
	// Get retrieves a client from the pool. If no idle clients are available and
	// the pool is not full, it may create a new one.
	Get(ctx context.Context) (T, error)
	// Put returns a client to the pool, making it available for reuse.
	Put(T)
	// Close terminates all clients in the pool and prevents new ones from being
	// created.
	Close() error
	// Len returns the number of idle clients currently in the pool.
	Len() int
}

type poolItem[T any] struct {
	client T
	retry  bool
}

// poolImpl is the internal implementation of the Pool interface. It manages a
// channel of clients, a factory for creating new clients, and a semaphore for
// controlling the pool size.
type poolImpl[T ClosableClient] struct {
	clients            chan poolItem[T]
	factory            func(context.Context) (T, error)
	sem                *semaphore.Weighted
	mu                 sync.Mutex
	closed             atomic.Bool
	disableHealthCheck bool
}

// New creates a new connection pool with the specified factory and size
// constraints. The pool is initialized with a minimum number of clients and can
// grow up to a maximum size.
//
// The type parameter `T` is constrained to types that implement the
// `ClosableClient` interface.
//
// Parameters:
//   - factory: A function that creates new clients.
//   - minSize: The initial number of clients to create.
//   - maxSize: The maximum number of clients the pool can hold.
//   - idleTimeout: (Not yet used) Intended for future implementation of idle
//     connection handling.
//
// Returns a new `Pool` instance or an error if the configuration is invalid.
func New[T ClosableClient](
	factory func(context.Context) (T, error),
	minSize, maxSize int,
	_ time.Duration, // idleTimeout is not used yet
	disableHealthCheck bool,
) (Pool[T], error) {
	if maxSize <= 0 {
		return nil, fmt.Errorf("maxSize must be positive")
	}
	if minSize < 0 || minSize > maxSize {
		return nil, fmt.Errorf("invalid minSize/maxSize configuration")
	}

	p := &poolImpl[T]{
		clients:            make(chan poolItem[T], maxSize),
		factory:            factory,
		sem:                semaphore.NewWeighted(int64(maxSize)),
		disableHealthCheck: disableHealthCheck,
	}

	// If health checks are disabled, we can pre-fill the pool without checks.
	if disableHealthCheck {
		for i := 0; i < minSize; i++ {
			client, err := factory(context.Background())
			if err != nil {
				_ = p.Close()
				return nil, fmt.Errorf("factory failed to create initial client: %w", err)
			}
			v := reflect.ValueOf(client)
			if (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) && v.IsNil() {
				_ = p.Close()
				return nil, fmt.Errorf("factory returned nil client")
			}
			p.clients <- poolItem[T]{client: client}
		}
		if !p.sem.TryAcquire(int64(minSize)) {
			return nil, fmt.Errorf("failed to acquire permits for initial clients")
		}
		return p, nil
	}

	// With health checks enabled, we need to ensure clients are healthy before adding.
	for i := 0; i < minSize; i++ {
		client, err := factory(context.Background())
		if err != nil {
			_ = p.Close()
			return nil, fmt.Errorf("factory failed to create initial client: %w", err)
		}
		v := reflect.ValueOf(client)
		if (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) && v.IsNil() {
			_ = p.Close()
			return nil, fmt.Errorf("factory returned nil client")
		}
		p.clients <- poolItem[T]{client: client}
	}
	// Take permits for the initial clients
	if !p.sem.TryAcquire(int64(minSize)) {
		// This should not happen given the checks above
		return nil, fmt.Errorf("failed to acquire permits for initial clients")
	}

	return p, nil
}

// Get retrieves a client from the pool. It first attempts to fetch an idle
// client from the channel. If none are available and the pool has not reached
// its maximum size, it creates a new client using the pool's factory.
//
// The method ensures that any client returned is healthy by checking
// `IsHealthy()`. If an unhealthy client is found, it is closed, and the process
// is retried. If the pool is full, the method will block until a client is
// returned to the pool or the context is canceled.
//
// Parameters:
//   - ctx: The context for the operation, which can be used to cancel the wait
//     for a client.
//
// Returns a client from the pool or an error if the pool is closed, the context
// is canceled, or the factory fails to create a new client.
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
			client := item.client

			if p.disableHealthCheck || client.IsHealthy(ctx) {
				return client, nil
			}
			_ = lo.Try(client.Close)
			p.sem.Release(1)
			// Small backoff to prevent tight loops when service is down
			time.Sleep(100 * time.Millisecond)
			continue // Retry.
		default:
			// Channel is empty, proceed.
		}

		// If no idle client, try to create a new one without blocking.
		if p.sem.TryAcquire(1) {
			// Acquired a permit, but double-check for a race condition where
			// a client was returned just before we acquired the permit.
			select {
			case item, ok := <-p.clients:
				// A client was available, so we use it and release the permit.
				p.sem.Release(1)
				if !ok {
					return zero, ErrPoolClosed
				}
				if item.retry {
					continue
				}
				client := item.client

				if p.disableHealthCheck || client.IsHealthy(ctx) {
					return client, nil
				}
				_ = lo.Try(client.Close)
				p.sem.Release(1)
				// Small backoff to prevent tight loops when service is down
				time.Sleep(100 * time.Millisecond)
				continue // Retry.
			default:
				// No client, so create a new one.
				// We must check if the pool was closed *after* we acquired the permit.
				if p.closed.Load() {
					p.sem.Release(1) // Don't leak the permit
					return zero, ErrPoolClosed
				}
				client, err := p.factory(ctx)
				if err != nil {
					p.sem.Release(1)
					return zero, err
				}

				// Check again if the pool was closed while we were creating a client.
				if p.closed.Load() {
					_ = lo.Try(client.Close)
					p.sem.Release(1)
					return zero, ErrPoolClosed
				}

				return client, nil
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
			client := item.client

			if p.disableHealthCheck || client.IsHealthy(ctx) {
				return client, nil
			}
			_ = lo.Try(client.Close)
			p.sem.Release(1)
			// Small backoff to prevent tight loops when service is down
			time.Sleep(100 * time.Millisecond)
			continue // Retry.
		case <-ctx.Done():
			return zero, ctx.Err()
		}
	}
}

// Put returns a client to the pool for reuse. If the pool is closed or the
// client is not healthy, the client is closed and discarded. If the pool's idle
// queue is full, the client is also discarded to prevent blocking. In all cases
// where a client is discarded, a permit is released to allow a new client to be
// created if needed.
//
// Parameters:
//   - client: The client to return to the pool.
func (p *poolImpl[T]) Put(client T) {
	v := reflect.ValueOf(client)
	if (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) && v.IsNil() {
		p.sem.Release(1)
		return
	}

	if p.closed.Load() {
		_ = lo.Try(client.Close)
		p.sem.Release(1) // Release permit as the client is discarded
		return
	}

	// We do NOT check health here to avoid blocking the caller.
	// IsHealthy() can be slow (network calls), and Put is often used in defer.
	// Any unhealthy client returned here will be detected and discarded by Get()
	// when it is next retrieved.

	p.mu.Lock()
	if p.closed.Load() {
		p.mu.Unlock()
		_ = lo.Try(client.Close)
		p.sem.Release(1)
		return
	}

	select {
	case p.clients <- poolItem[T]{client: client}:
		p.mu.Unlock()
	default:
		// Idle pool is full, discard client. The permit for this client is
		// effectively leaked, but this is the only safe option. Releasing
		// the permit would allow the pool to create more clients than
		// maxSize, and we can't tell if this client came from the pool
		// in the first place.
		p.mu.Unlock()
		_ = lo.Try(client.Close)
	}
}

// Close shuts down the pool, closing all idle clients and preventing any new
// operations. Any subsequent calls to `Get` will return `ErrPoolClosed`.
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
		_ = lo.Try(client.Close)
		p.sem.Release(1)
	}
	return nil
}

// Len returns the current number of idle clients in the pool.
func (p *poolImpl[T]) Len() int {
	return len(p.clients)
}

// UntypedPool defines a non-generic interface for a pool, allowing for
// management of pools of different types in a single collection.
type UntypedPool interface {
	io.Closer
	// Len returns the number of idle clients currently in the pool.
	Len() int
}

// Manager provides a way to manage multiple named connection pools. It allows
// for registering, retrieving, and closing pools in a centralized manner, which
// is useful for applications that need to connect to multiple upstream
// services.
type Manager struct {
	pools map[string]any
	mu    sync.RWMutex
}

// NewManager creates and returns a new pool Manager for managing multiple named
// connection pools.
func NewManager() *Manager {
	return &Manager{
		pools: make(map[string]any),
	}
}

// Register adds a new pool to the manager under a given name. If a pool with
// the same name already exists, the old pool is closed before the new one is
// registered. This ensures that there are no resource leaks from orphaned
// pools.
//
// Parameters:
//   - name: The name to register the pool under.
//   - pool: The pool to register.
func (m *Manager) Register(name string, pool any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if oldPool, ok := m.pools[name]; ok {
		if p, isCloser := oldPool.(io.Closer); isCloser {
			logging.GetLogger().Info("Closing old entry", "name", name)
			_ = lo.Try(p.Close)
		}
	}
	m.pools[name] = pool
}

// Deregister closes and removes a pool from the manager.
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

// Get retrieves a typed pool from the manager by name. It uses a type parameter
// `T` to ensure that the returned pool is of the expected type.
//
// Parameters:
//   - m: The Manager instance.
//   - name: The name of the pool to retrieve.
//
// Returns the typed `Pool` and a boolean indicating whether the pool was found
// and of the correct type.
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

// CloseAll iterates through all registered pools in the manager and closes them,
// releasing all their associated resources. This is typically called during a
// graceful shutdown of the application.
func (m *Manager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, untypedPool := range m.pools {
		logging.GetLogger().Info("Closing pool", "name", name)
		if p, ok := untypedPool.(io.Closer); ok {
			_ = lo.Try(p.Close)
		}
	}
}
