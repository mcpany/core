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

const DiskStoreType = "disk"

type DiskStore struct {
	fs   afero.Fs
	path string
}

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

func (s *DiskStore) Get(ctx context.Context, key any) (any, error) {
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
		_ = s.Delete(ctx, key)
		return nil, store.NotFound{}
	}

	// Return the raw value (bytes) so the caller can unmarshal it
	return []byte(entry.Value), nil
}

func (s *DiskStore) GetWithTTL(ctx context.Context, key any) (any, time.Duration, error) {
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
		_ = s.Delete(ctx, key)
		return nil, 0, store.NotFound{}
	}

	ttl := time.Until(entry.ExpiresAt)
	return []byte(entry.Value), ttl, nil
}

func (s *DiskStore) Set(ctx context.Context, key any, value any, options ...store.Option) error {
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

func (s *DiskStore) Delete(ctx context.Context, key any) error {
	strKey := fmt.Sprintf("%v", key)
	return s.fs.Remove(filepath.Join(s.path, strKey))
}

func (s *DiskStore) Invalidate(ctx context.Context, options ...store.InvalidateOption) error {
	return nil
}

func (s *DiskStore) Clear(ctx context.Context) error {
	return s.fs.RemoveAll(s.path)
}

func (s *DiskStore) GetType() string {
	return DiskStoreType
}
