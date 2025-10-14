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
	ErrPoolClosed = fmt.Errorf("pool has been closed")
	ErrPoolFull   = fmt.Errorf("pool is full")
)

type ClosableClient interface {
	Close() error
	IsHealthy() bool
}

type Pool[T ClosableClient] interface {
	Get(context.Context) (T, error)
	Put(T)
	Close()
	Len() int
}

type poolImpl[T ClosableClient] struct {
	clients chan T
	factory func(context.Context) (T, error)
	sem     *semaphore.Weighted
	mu      sync.Mutex
	closed  bool
}

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

func (p *poolImpl[T]) Put(client T) {
	v := reflect.ValueOf(client)
	if (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) && v.IsNil() {
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

func (p *poolImpl[T]) Len() int {
	return len(p.clients)
}

type UntypedPool interface {
	Close()
	Len() int
}

type Manager struct {
	pools map[string]any
	mu    sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		pools: make(map[string]any),
	}
}

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
