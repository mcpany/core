// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher monitors configuration files for changes and triggers a reload.
//
// Summary: Watches the parent directory of specified files to handle atomic saves (rename/move) commonly used by text editors.
type Watcher struct {
	watcher *fsnotify.Watcher
	done    chan bool
	mu      sync.Mutex
	timer   *time.Timer
}

// NewWatcher creates a new file watcher.
//
// Summary: Initializes a new file system watcher.
//
// Returns:
//   - *Watcher: A pointer to a new Watcher instance.
//   - error: An error if the watcher creation fails.
//
// Errors/Throws:
//   - Returns error if fsnotify.NewWatcher() fails.
func NewWatcher() (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Watcher{
		watcher: watcher,
		done:    make(chan bool),
	}, nil
}

// Watch starts monitoring the specified configuration paths.
//
// Summary: Registers the specified paths for monitoring and triggers the reload callback on changes.
//
// Parameters:
//   - paths: []string. A slice of file or directory paths to watch.
//   - reloadFunc: func(). The function to call when a change is detected.
//
// Returns:
//   - error: An error if adding paths to the watcher fails.
//
// Errors/Throws:
//   - Returns error if resolving absolute paths fails.
//   - Returns error if adding a directory to the watcher fails.
//
// Side Effects:
//   - Starts a goroutine to listen for file system events.
//   - Logs errors and informational messages.
func (w *Watcher) Watch(paths []string, reloadFunc func()) error {
	// Map of parent directory -> list of filenames to watch in that directory
	watchedFiles := make(map[string][]string)

	for _, path := range paths {
		if isURL(path) {
			continue
		}
		absPath, err := filepath.Abs(path)
		if err != nil {
			log.Printf("Failed to get absolute path for %s: %v", path, err)
			continue
		}

		// Since we want to handle atomic saves (rename), we MUST watch the parent directory of files.
		parent := filepath.Dir(absPath)
		filename := filepath.Base(absPath)

		if _, exists := watchedFiles[parent]; !exists {
			watchedFiles[parent] = []string{}
		}
		watchedFiles[parent] = append(watchedFiles[parent], filename)
	}

	go func() {
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}

				// Check if this event is relevant
				relevant := false

				// Logic:
				// 1. If we are watching the directory of this event
				// 2. And the event name matches one of the files we are interested in.

				parent := filepath.Dir(event.Name)
				filename := filepath.Base(event.Name)

				if files, ok := watchedFiles[parent]; ok {
					for _, f := range files {
						if f == filename {
							relevant = true
							break
						}
					}
				}

				// Also check absolute path matching
				if !relevant {
					absName, _ := filepath.Abs(event.Name)
					parent = filepath.Dir(absName)
					filename = filepath.Base(absName)
					if files, ok := watchedFiles[parent]; ok {
						for _, f := range files {
							if f == filename {
								relevant = true
								break
							}
						}
					}
				}

				// Handle Vim backup files (ends with ~)
				if strings.HasSuffix(filename, "~") {
					relevant = false
				}

				if relevant {
					// Trigger on Write, Create, Rename, Chmod
					// Atomic save: Create (new file) -> Rename (to old file).
					// So we get Create (tmp) -> Rename (tmp->target).
					// Or Rename (target->backup).
					// If we see any change to the target filename, we reload.
					if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename|fsnotify.Chmod) != 0 {
						w.mu.Lock()
						if w.timer != nil {
							w.timer.Stop()
						}
						// Debounce for 500ms to avoid multiple reloads for a single save event
						w.timer = time.AfterFunc(500*time.Millisecond, func() {
							log.Println("Reloading configuration...")
							reloadFunc()
						})
						w.mu.Unlock()
					}
				}

			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			case <-w.done:
				return
			}
		}
	}()

	for parent := range watchedFiles {
		if err := w.watcher.Add(parent); err != nil {
			return err
		}
	}

	<-w.done
	return nil
}

// Close stops the file watcher and releases resources.
//
// Summary: Signals the watcher to stop and closes the underlying fsnotify watcher.
//
// Side Effects:
//   - Closes the done channel.
//   - Closes the fsnotify watcher.
func (w *Watcher) Close() {
	close(w.done)
	_ = w.watcher.Close()
}
