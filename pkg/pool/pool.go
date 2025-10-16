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
	"reflect"
	"sync"

	"github.com/mcpxy/core/pkg/logging"
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
	IsHealthy() bool
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
	Close()
	// Len returns the number of idle clients currently in the pool.
	Len() int
}

// poolImpl is the internal implementation of the Pool interface. It manages a
// channel of clients, a factory for creating new clients, and a semaphore for
// controlling the pool size.
type poolImpl[T ClosableClient] struct {
	clients chan T
	factory func(context.Context) (T, error)
	sem     *semaphore.Weighted
	mu      sync.Mutex
	closed  bool
}

// New creates a new connection pool with the specified factory and size
// constraints.
//
// factory is a function that creates new clients.
// minSize is the initial number of clients to create.
// maxSize is the maximum number of clients the pool can hold.
// idleTimeout is not yet used but is intended for future implementation of
// idle connection handling.
func New[T ClosableClient](
	factory func(context.Context) (T, error),
	minSize, maxSize, idleTimeout int, // idleTimeout is not used yet
) (Pool[T], error) {
	if maxSize <= 0 {
		return nil, fmt.Errorf("maxSize must be positive")
	}
	if minSize < 0 || minSize > maxSize {
		return nil, fmt.Errorf("invalid minSize/maxSize configuration")
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
			return nil, fmt.Errorf("factory failed to create initial client: %w", err)
		}
		p.clients <- client
	}
	// Take permits for the initial clients
	if !p.sem.TryAcquire(int64(minSize)) {
		// This should not happen given the checks above
		return nil, fmt.Errorf("failed to acquire permits for initial clients")
	}

	return p, nil
}

// Get retrieves a client from the pool. It first attempts to fetch an idle
// client. If none are available, and the pool has not reached its maximum size,
// it creates a new client using the pool's factory. It ensures that any client
// returned is healthy.
func (p *poolImpl[T]) Get(ctx context.Context) (T, error) {
	var zero T

	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return zero, ErrPoolClosed
	}
	p.mu.Unlock()

	// Loop to ensure we return a healthy client
	for {
		select {
		case client := <-p.clients:
			if client.IsHealthy() {
				return client, nil
			}
			lo.Try(client.Close)
			// Unhealthy client found and closed. Release its permit and loop again.
			p.sem.Release(1)
		default:
			// No idle clients, try to create a new one.
			if err := p.sem.Acquire(ctx, 1); err != nil {
				return zero, err // Context canceled or pool is full and timeout exceeded.
			}

			client, err := p.factory(ctx)
			if err != nil {
				p.sem.Release(1) // Creation failed, release permit.
				return zero, err
			}
			return client, nil
		}
	}
}

// Put returns a client to the pool. If the pool is closed or the client is not
// healthy, the client is closed and discarded. If the pool's idle queue is
// full, the client is also discarded.
func (p *poolImpl[T]) Put(client T) {
	v := reflect.ValueOf(client)
	if (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) && v.IsNil() {
		p.sem.Release(1)
		return
	}

	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		lo.Try(client.Close)
		p.sem.Release(1) // Release permit as the client is discarded
		return
	}

	if !client.IsHealthy() {
		p.mu.Unlock()
		lo.Try(client.Close)
		p.sem.Release(1)
		return
	}

	select {
	case p.clients <- client:
		p.mu.Unlock()
	default:
		// Idle pool is full, discard client and release permit.
		p.mu.Unlock()
		lo.Try(client.Close)
		p.sem.Release(1)
	}
}

// Close shuts down the pool, closing all idle clients and preventing any new
// operations on the pool.
func (p *poolImpl[T]) Close() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	close(p.clients)
	p.mu.Unlock()

	// Drain the channel and close all idle clients
	for client := range p.clients {
		lo.Try(client.Close)
		p.sem.Release(1)
	}
}

// Len returns the current number of idle clients in the pool.
func (p *poolImpl[T]) Len() int {
	return len(p.clients)
}

// UntypedPool defines a non-generic interface for a pool, allowing for
// management of pools of different types in a single collection.
type UntypedPool interface {
	Close()
	Len() int
}

// Manager provides a way to manage multiple named connection pools. It allows
// for registering, retrieving, and closing pools in a centralized manner.
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

// Register adds a new pool to the manager under a given name. If a pool with
// the same name already exists, the old pool is closed before the new one is
// added.
func (m *Manager) Register(name string, pool any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if oldPool, ok := m.pools[name]; ok {
		if p, isPool := oldPool.(UntypedPool); isPool {
			logging.GetLogger().Info("Closing pool", "name", name)
			p.Close()
		}
	}
	m.pools[name] = pool
}

// Get retrieves a typed pool from the manager by name. It returns the pool and
// a boolean indicating whether the pool was found and of the correct type.
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
