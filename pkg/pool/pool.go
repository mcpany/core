/*
 * Copyright 2025 Author(s) of MCPX
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
	"reflect"
	"sync"

	"github.com/mcpxy/mcpx/pkg/logging"
	"github.com/samber/lo"
	"golang.org/x/sync/semaphore"
)

var (
	// ErrPoolClosed is returned when an operation is attempted on a closed pool.
	ErrPoolClosed = fmt.Errorf("pool has been closed")
	// ErrPoolFull is returned when a new connection is requested, but the pool has
	// reached its maximum capacity.
	ErrPoolFull = fmt.Errorf("pool is full")
)

// ClosableClient defines the interface for clients that can be managed by a
// connection pool.
type ClosableClient interface {
	// Close terminates the client connection.
	Close() error
	// IsHealthy checks if the client connection is still active and usable.
	IsHealthy() bool
}

// Pool defines the interface for a generic, thread-safe connection pool.
// It manages a set of clients of type T, where T must implement the
// ClosableClient interface.
type Pool[T ClosableClient] interface {
	// Get retrieves a client from the pool. If no idle clients are available and
	// the pool is not full, a new client may be created. It blocks until a
	// client is available or the context is canceled.
	Get(context.Context) (T, error)
	// Put returns a client to the pool, making it available for reuse.
	Put(T)
	// Close shuts down the pool, closing all idle clients and preventing new
	// clients from being retrieved.
	Close()
	// Len returns the number of idle clients currently in the pool.
	Len() int
}

// poolImpl is the internal implementation of the Pool interface.
type poolImpl[T ClosableClient] struct {
	clients chan T
	factory func(context.Context) (T, error)
	sem     *semaphore.Weighted
	mu      sync.Mutex
	closed  bool
}

// New creates a new connection pool for clients of type T.
//
// factory is a function that creates a new client.
// minSize is the number of clients to create initially.
// maxSize is the maximum number of clients the pool can hold.
// idleTimeout is not currently used but is intended for future implementation.
// It returns a new Pool or an error if the parameters are invalid or the
// initial clients cannot be created.
func New[T ClosableClient](
	factory func(context.Context) (T, error),
	minSize, maxSize, idleTimeout int,
) (Pool[T], error) {
	if minSize < 0 || maxSize <= 0 || minSize > maxSize {
		return nil, fmt.Errorf("invalid pool size parameters")
	}

	p := &poolImpl[T]{
		clients: make(chan T, maxSize),
		factory: factory,
		sem:     semaphore.NewWeighted(int64(maxSize)),
	}

	for i := 0; i < minSize; i++ {
		client, err := factory(context.Background())
		if err != nil {
			p.Close()
			return nil, fmt.Errorf("failed to create initial clients: %w", err)
		}
		p.clients <- client
	}

	return p, nil
}

// Get retrieves a client from the pool. It first tries to get an idle client.
// If none are available and the pool has not reached its maximum size, it
// creates a new one using the factory function.
func (p *poolImpl[T]) Get(ctx context.Context) (T, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		var zero T
		return zero, ErrPoolClosed
	}
	p.mu.Unlock()

	select {
	case client := <-p.clients:
		if client.IsHealthy() {
			if !p.sem.TryAcquire(1) {
				// Should not happen if pool is managed correctly
				p.clients <- client // put it back
				var zero T
				return zero, ErrPoolFull
			}
			return client, nil
		}
		lo.Try(client.Close)
		// Unhealthy client, fall through to create a new one.
	default:
		// No idle clients.
	}

	if !p.sem.TryAcquire(1) {
		var zero T
		return zero, ErrPoolFull
	}

	client, err := p.factory(ctx)
	if err != nil {
		p.sem.Release(1)
		var zero T
		return zero, err
	}
	return client, nil
}

// Put returns a client to the pool. If the client is not healthy, it is closed
// and discarded. If the pool is full, the client is also closed.
func (p *poolImpl[T]) Put(client T) {
	v := reflect.ValueOf(client)
	if (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) && v.IsNil() {
		return
	}

	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		lo.Try(client.Close)
		return
	}
	p.mu.Unlock()

	if !client.IsHealthy() {
		lo.Try(client.Close)
		return
	}

	select {
	case p.clients <- client:
		// Returned to idle queue.
	default:
		// Idle queue is full, discard client.
		lo.Try(client.Close)
	}
	p.sem.Release(1)
}

// Close closes the pool and all its underlying client connections.
// Once closed, the pool cannot be used anymore.
func (p *poolImpl[T]) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}
	p.closed = true

	close(p.clients)
	for client := range p.clients {
		lo.Try(client.Close)
	}
}

// Len returns the current number of idle clients in the pool.
func (p *poolImpl[T]) Len() int {
	return len(p.clients)
}

// UntypedPool is an interface that represents a pool without exposing its generic
// client type. This is useful for managing multiple pools of different types in
// a single container.
type UntypedPool interface {
	// Close closes the pool.
	Close()
	// Len returns the number of idle clients in the pool.
	Len() int
}

// Manager provides a way to manage multiple named connection pools. It allows
// for registering pools and closing them all at once.
type Manager struct {
	pools map[string]any
	mu    sync.RWMutex
}

// NewManager creates and returns a new pool Manager.
func NewManager() *Manager {
	return &Manager{
		pools: make(map[string]any),
	}
}

// Register adds a new pool to the manager with a given name. If a pool with the
// same name already exists, it is overwritten.
//
// name is the name to associate with the pool.
// pool is the pool to be registered. It must be of a type that can be asserted
// to a Pool[T] or UntypedPool.
func (m *Manager) Register(name string, pool any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pools[name] = pool
}

// Get retrieves a typed pool from the manager by its name. It returns the pool
// and a boolean indicating whether the pool was found and of the correct type.
//
// T is the type of the client managed by the pool.
// m is the manager instance.
// name is the name of the pool to retrieve.
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

// CloseAll iterates through all registered pools and closes them.
func (m *Manager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, untypedPool := range m.pools {
		logging.GetLogger().Info("Closing pool", "name", name)
		if p, ok := untypedPool.(UntypedPool); ok {
			p.Close()
		}
	}
}
