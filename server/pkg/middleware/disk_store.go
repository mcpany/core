// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/eko/gocache/lib/v4/store"
	"github.com/spf13/afero"
)

// DiskStoreType is the constant for the disk store type.
const DiskStoreType = "disk"

// DiskStore implements the store.StoreInterface for filesystem-based caching.
type DiskStore struct {
	fs   afero.Fs
	path string
}

// NewDiskStore creates a new DiskStore.
func NewDiskStore(fs afero.Fs, path string) *DiskStore {
	if fs == nil {
		fs = afero.NewOsFs()
	}
	// Ensure directory exists
	_ = fs.MkdirAll(path, 0755)
	return &DiskStore{
		fs:   fs,
		path: path,
	}
}

type diskCacheEntry struct {
	Value     json.RawMessage `json:"value"`
	ExpiresAt time.Time       `json:"expires_at"`
}

// Get retrieves a value from the cache.
func (s *DiskStore) Get(_ context.Context, key any) (any, error) {
	strKey := fmt.Sprintf("%v", key)
	data, err := afero.ReadFile(s.fs, filepath.Join(s.path, strKey))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, store.NotFound{}
		}
		return nil, err
	}

	var entry diskCacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}

	if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
		// Can't delete easily without context if Delete signature required it?
		// Delete requires context, but Get signature also has it.
		// Wait, I replaced ctx with _ in signature to satisfy unused-parameter.
		// But I need it for Delete?
		// Actually Delete doesn't use ctx either in my impl.
		// So I can pass nil or context.Background() internally or just keep ctx named _ and pass it?
		// If I rename ctx to _, I can't use it.
		// But Delete(ctx) uses it? Let's check Delete implementation.
		// Delete implementation below has unused ctx too.
		// So I can just ignore it.
		_ = s.Delete(context.Background(), key)
		return nil, store.NotFound{}
	}

	// Return the raw value (bytes) so the caller can unmarshal it
	return []byte(entry.Value), nil
}

// GetWithTTL retrieves a value from the cache with its TTL.
func (s *DiskStore) GetWithTTL(_ context.Context, key any) (any, time.Duration, error) {
	strKey := fmt.Sprintf("%v", key)
	data, err := afero.ReadFile(s.fs, filepath.Join(s.path, strKey))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, 0, store.NotFound{}
		}
		return nil, 0, err
	}

	var entry diskCacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, 0, err
	}

	if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
		_ = s.Delete(context.Background(), key)
		return nil, 0, store.NotFound{}
	}

	ttl := time.Until(entry.ExpiresAt)
	return []byte(entry.Value), ttl, nil
}

// Set sets a value in the cache.
func (s *DiskStore) Set(_ context.Context, key any, value any, options ...store.Option) error {
	strKey := fmt.Sprintf("%v", key)
	opts := store.ApplyOptions(options...)

	// Marshal the value to JSON
	valBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	entry := diskCacheEntry{
		Value: valBytes,
	}
	if opts.Expiration > 0 {
		entry.ExpiresAt = time.Now().Add(opts.Expiration)
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	return afero.WriteFile(s.fs, filepath.Join(s.path, strKey), data, 0644)
}

// Delete removes a value from the cache.
func (s *DiskStore) Delete(_ context.Context, key any) error {
	strKey := fmt.Sprintf("%v", key)
	return s.fs.Remove(filepath.Join(s.path, strKey))
}

// Invalidate invalidates cache entries (not implemented).
func (s *DiskStore) Invalidate(_ context.Context, _ ...store.InvalidateOption) error {
	return nil
}

// Clear clears the cache.
func (s *DiskStore) Clear(_ context.Context) error {
	return s.fs.RemoveAll(s.path)
}

// GetType returns the cache type.
func (s *DiskStore) GetType() string {
	return DiskStoreType
}
