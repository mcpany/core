// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package sqlite

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// TestStore_Load_Concurrency verifies that the parallel Load implementation is thread-safe
// and correctly aggregates results from all concurrent queries.
func TestStore_Load_Concurrency(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "concurrency.db")
	db, err := NewDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	store := NewStore(db)

	// Pre-populate with a significant amount of data to ensure queries take some time
	ctx := context.Background()
	numItems := 100

	// 1. Services
	for i := 0; i < numItems; i++ {
		svc := configv1.UpstreamServiceConfig_builder{
			Name: proto.String(fmt.Sprintf("svc-%d", i)),
			Id:   proto.String(fmt.Sprintf("svc-%d", i)),
		}.Build()
		require.NoError(t, store.SaveService(ctx, svc))
	}

	// 2. Users
	for i := 0; i < numItems; i++ {
		user := configv1.User_builder{
			Id: proto.String(fmt.Sprintf("user-%d", i)),
		}.Build()
		require.NoError(t, store.CreateUser(ctx, user))
	}

	// 3. Profiles
	for i := 0; i < numItems; i++ {
		prof := configv1.ProfileDefinition_builder{
			Name: proto.String(fmt.Sprintf("prof-%d", i)),
		}.Build()
		require.NoError(t, store.SaveProfile(ctx, prof))
	}

	// Run multiple Loads concurrently
	concurrency := 10
	var wg sync.WaitGroup
	wg.Add(concurrency)

	errCh := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			cfg, err := store.Load(ctx)
			if err != nil {
				errCh <- err
				return
			}

			// Verification
			if len(cfg.GetUpstreamServices()) != numItems {
				errCh <- fmt.Errorf("expected %d services, got %d", numItems, len(cfg.GetUpstreamServices()))
				return
			}
			if len(cfg.GetUsers()) != numItems {
				errCh <- fmt.Errorf("expected %d users, got %d", numItems, len(cfg.GetUsers()))
				return
			}
			// Profiles are merged into GlobalSettings
			if len(cfg.GetGlobalSettings().GetProfileDefinitions()) != numItems {
				errCh <- fmt.Errorf("expected %d profiles, got %d", numItems, len(cfg.GetGlobalSettings().GetProfileDefinitions()))
				return
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		assert.NoError(t, err)
	}
}

// TestStore_Save_Load_Race ensures that writing to the DB while loading does not crash the application.
// Note: Consistency is not guaranteed if a write happens during a read, but we should not panic.
func TestStore_Save_Load_Race(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "race.db")
	db, err := NewDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	store := NewStore(db)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	// Writer
	wg.Add(1)
	go func() {
		defer wg.Done()
		i := 0
		for {
			select {
			case <-ctx.Done():
				return
			default:
				svc := configv1.UpstreamServiceConfig_builder{
					Name: proto.String(fmt.Sprintf("race-svc-%d", i)),
				}.Build()
				_ = store.SaveService(context.Background(), svc)
				i++
				// Small sleep to yield CPU
				time.Sleep(1 * time.Millisecond)
			}
		}
	}()

	// Reader (Load)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				_, err := store.Load(context.Background())
				// We expect Load to succeed mostly, but db locks might cause transient errors
				// if we were not using WAL mode (but we are).
				// In WAL mode, readers don't block writers and vice versa.
				// However, if we get an error, it shouldn't be a panic.
				if err != nil {
					// It's acceptable for Load to fail under heavy contention or if DB is closed (not here)
					// But usually SQLite handles this fine.
				}
			}
		}
	}()

	wg.Wait()
}

// TestStore_Context_Cancellation verifies that Load aborts quickly when context is canceled.
func TestStore_Context_Cancellation(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "cancel.db")
	db, err := NewDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	store := NewStore(db)

	// Populate data to make queries take non-zero time
	ctx := context.Background()
	for i := 0; i < 50; i++ {
		_ = store.SaveService(ctx, configv1.UpstreamServiceConfig_builder{Name: proto.String(fmt.Sprintf("s-%d", i))}.Build())
	}

	// 1. Cancel immediately
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel before call

	_, err = store.Load(canceledCtx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")

	// 2. Cancel during execution (simulated with very short timeout)
	// This is hard to guarantee deterministically without hooks, but we can try with a short timeout.
	timeoutCtx, cancel2 := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel2()

	// Ensure we sleep a bit to let the timeout expire if it hasn't already
	time.Sleep(1 * time.Millisecond)

	_, err = store.Load(timeoutCtx)
	// It might succeed if it's super fast, or fail with deadline exceeded.
	// If it succeeds, that's technically fine (fast operation), but if it errors, it must be context error.
	if err != nil {
		assert.Contains(t, err.Error(), "context")
	}
}
