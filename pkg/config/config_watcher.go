// Copyright 2025 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/mcpany/core/pkg/logging"
)

// ConfigWatcher monitors configuration files for changes and triggers a reload.
type ConfigWatcher struct {
	watcher *fsnotify.Watcher
	reload  func() error
	done    chan struct{}
}

// NewConfigWatcher creates a new ConfigWatcher.
func NewConfigWatcher(paths []string, reload func() error) (*ConfigWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	for _, path := range paths {
		err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				if err := watcher.Add(path); err != nil {
					return fmt.Errorf("failed to add path to watcher: %w", err)
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return &ConfigWatcher{
		watcher: watcher,
		reload:  reload,
		done:    make(chan struct{}),
	}, nil
}

// Start starts the file watcher.
func (cw *ConfigWatcher) Start(ctx context.Context) {
	log := logging.GetLogger()
	go func() {
		for {
			select {
			case event, ok := <-cw.watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Info("Configuration file modified, reloading...", "file", event.Name)
					if err := cw.reload(); err != nil {
						log.Error("Failed to reload configuration", "error", err)
					}
				}
			case err, ok := <-cw.watcher.Errors:
				if !ok {
					return
				}
				log.Error("File watcher error", "error", err)
			case <-cw.done:
				return
			}
		}
	}()
}

// Stop stops the file watcher.
func (cw *ConfigWatcher) Stop() {
	close(cw.done)
	if err := cw.watcher.Close(); err != nil {
		logging.GetLogger().Error("Failed to close file watcher", "error", err)
	}
}
