// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resource

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

func TestManager_ConcurrentAccess(t *testing.T) {
	t.Parallel()
	rm := NewManager()
	const numGoroutines = 50
	const numIterations = 100

	g, ctx := errgroup.WithContext(context.Background())

	// Adders
	for i := 0; i < numGoroutines; i++ {
		i := i
		g.Go(func() error {
			for j := 0; j < numIterations; j++ {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					uri := fmt.Sprintf("res://%d-%d", i, j)
					rm.AddResource(&mockResource{uri: uri, service: "s1"})
					// Small sleep to yield
					time.Sleep(time.Microsecond)
				}
			}
			return nil
		})
	}

	// Listers
	for i := 0; i < numGoroutines; i++ {
		g.Go(func() error {
			for j := 0; j < numIterations; j++ {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					_ = rm.ListResources()
					time.Sleep(time.Microsecond)
				}
			}
			return nil
		})
	}

	// Removers
	for i := 0; i < numGoroutines; i++ {
		i := i
		g.Go(func() error {
			for j := 0; j < numIterations; j++ {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					uri := fmt.Sprintf("res://%d-%d", i, j)
					rm.RemoveResource(uri)
					time.Sleep(time.Microsecond)
				}
			}
			return nil
		})
	}

	assert.NoError(t, g.Wait())
}

func TestManager_CallbackConcurrency(t *testing.T) {
	t.Parallel()
	rm := NewManager()
	var callbackCount int64

	// Register a callback that takes some time
	rm.OnListChanged(func() {
		atomic.AddInt64(&callbackCount, 1)
		time.Sleep(time.Microsecond * 10)
	})

	const numGoroutines = 20
	const numIterations = 50

	g, ctx := errgroup.WithContext(context.Background())

	for i := 0; i < numGoroutines; i++ {
		i := i
		g.Go(func() error {
			for j := 0; j < numIterations; j++ {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					if j%2 == 0 {
						rm.AddResource(&mockResource{uri: fmt.Sprintf("res://%d-%d", i, j)})
					} else {
						rm.RemoveResource(fmt.Sprintf("res://%d-%d", i, j-1))
					}
				}
			}
			return nil
		})
	}

	assert.NoError(t, g.Wait())
	assert.Greater(t, atomic.LoadInt64(&callbackCount), int64(0))
}

func TestManager_ClearResourcesConcurrency(t *testing.T) {
	t.Parallel()
	rm := NewManager()
	const numGoroutines = 10
	const numIterations = 100

	g, ctx := errgroup.WithContext(context.Background())

	// Add resources for two services
	for i := 0; i < numGoroutines; i++ {
		i := i
		g.Go(func() error {
			for j := 0; j < numIterations; j++ {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					serviceID := "s1"
					if j%2 == 0 {
						serviceID = "s2"
					}
					rm.AddResource(&mockResource{
						uri:     fmt.Sprintf("res://%s-%d-%d", serviceID, i, j),
						service: serviceID,
					})
				}
			}
			return nil
		})
	}

	// Concurrently clear resources for s1
	for i := 0; i < 5; i++ {
		g.Go(func() error {
			for j := 0; j < numIterations/10; j++ {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					rm.ClearResourcesForService("s1")
					time.Sleep(time.Millisecond)
				}
			}
			return nil
		})
	}

	// Concurrently list resources
	for i := 0; i < 5; i++ {
		g.Go(func() error {
			for j := 0; j < numIterations; j++ {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					_ = rm.ListResources()
				}
			}
			return nil
		})
	}

	assert.NoError(t, g.Wait())
}
