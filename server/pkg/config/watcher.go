// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/mcpany/core/pkg/logging"
)

// Watcher watches for changes in configuration files and triggers a callback.
type Watcher struct {
	watcher *fsnotify.Watcher
	// watchedFiles maps absolute file path to boolean (true).
	watchedFiles map[string]bool
	// watchedDirs maps directory path to count of watched files in it.
	watchedDirs map[string]int
	mu          sync.Mutex
	done        chan struct{}
}

// NewWatcher creates a new Watcher.
func NewWatcher() (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	return &Watcher{
		watcher:      w,
		watchedFiles: make(map[string]bool),
		watchedDirs:  make(map[string]int),
		done:         make(chan struct{}),
	}, nil
}

// Add adds a file to the watcher.
func (w *Watcher) Add(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Resolve absolute path to be safe
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for %s: %w", path, err)
	}

	if w.watchedFiles[absPath] {
		return nil // Already watched
	}

	dir := filepath.Dir(absPath)

	// If we aren't watching this directory yet, add it
	if w.watchedDirs[dir] == 0 {
		if err := w.watcher.Add(dir); err != nil {
			return fmt.Errorf("failed to watch directory %s: %w", dir, err)
		}
	}

	w.watchedDirs[dir]++
	w.watchedFiles[absPath] = true
	return nil
}

// Start starts the watcher loop. It runs until the context is canceled or Close is called.
// The onChange callback is called when a write event is detected.
// debounceDuration is the time to wait after the last event before triggering the callback,
// to avoid multiple reloads for a single save.
func (w *Watcher) Start(ctx context.Context, debounceDuration time.Duration, onChange func() error) {
	go func() {
		var (
			timer *time.Timer
			mu    sync.Mutex
		)

		for {
			select {
			case <-ctx.Done():
				return
			case <-w.done:
				return
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}

				// Check if event affects one of our watched files
				w.mu.Lock()
				isWatched := w.watchedFiles[event.Name]
				// Also check if it's a rename/create/chmod of the file we care about
				// event.Name should be the absolute path if we added absolute path?
				// fsnotify typically returns the path relative to the watch point or absolute if absolute was added.
				// Since we added absolute path of DIR, let's assume event.Name matches.
				// But just in case, let's check absolute.
				if !isWatched {
					absEventName, err := filepath.Abs(event.Name)
					if err == nil {
						isWatched = w.watchedFiles[absEventName]
					}
				}
				w.mu.Unlock()

				if !isWatched {
					continue
				}

				// event.Has(fsnotify.Write) - standard edit
				// event.Has(fsnotify.Create) - atomic save (new file replaces old)
				// event.Has(fsnotify.Rename) - atomic save (old file moved away? not relevant for new content)
				// event.Has(fsnotify.Chmod) - permission change

				// For atomic save (vim):
				// 1. Rename config.yaml -> config.yaml~ (Rename)
				// 2. Create config.yaml (Create)
				// 3. Write config.yaml (Write)
				// 4. Chmod config.yaml (Chmod)

				// We care if the file we are watching is Created or Written to.
				// Rename might indicate it's gone, but if it's replaced immediately, we catch the Create.

				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Chmod) {
					logging.GetLogger().Info("Config file changed", "event", event.String())

					mu.Lock()
					if timer != nil {
						timer.Stop()
					}
					timer = time.AfterFunc(debounceDuration, func() {
						logging.GetLogger().Info("Triggering config reload")
						if err := onChange(); err != nil {
							logging.GetLogger().Error("Failed to reload config", "error", err)
						}
					})
					mu.Unlock()
				}

			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				logging.GetLogger().Error("Watcher error", "error", err)
			}
		}
	}()
}

// Close closes the watcher.
func (w *Watcher) Close() error {
	close(w.done)
	return w.watcher.Close()
}
