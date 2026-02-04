// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package gc provides a background worker for global garbage collection of
// temporary files and directories.
package gc

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/mcpany/core/server/pkg/logging"
)

// Config defines the configuration for the GC Worker.
type Config struct {
	Enabled  bool
	Interval time.Duration
	TTL      time.Duration
	Paths    []string
}

// Worker implements a background worker for garbage collection.
type Worker struct {
	config Config
}

// New creates a new GC Worker.
//
// config holds the configuration settings.
//
// Returns the result.
func New(config Config) *Worker {
	if config.Interval <= 0 {
		config.Interval = 1 * time.Hour // Default 1 hour
	}
	if config.TTL <= 0 {
		config.TTL = 24 * time.Hour // Default 24 hours
	}
	return &Worker{
		config: config,
	}
}

// Start runs the GC worker in the background.
// It returns immediately and runs cleanup periodically until the context is canceled.
func (w *Worker) Start(ctx context.Context) {
	if !w.config.Enabled {
		logging.GetLogger().Info("Global GC worker is disabled")
		return
	}

	logging.GetLogger().Info("Starting Global GC worker",
		"interval", w.config.Interval,
		"ttl", w.config.TTL,
		"paths", len(w.config.Paths),
	)

	go func() {
		ticker := time.NewTicker(w.config.Interval)
		defer ticker.Stop()

		// Run immediately on startup
		w.runCleanup(ctx)

		for {
			select {
			case <-ctx.Done():
				logging.GetLogger().Info("Stopping Global GC worker")
				return
			case <-ticker.C:
				w.runCleanup(ctx)
			}
		}
	}()
}

// runCleanup executes a single GC cycle.
func (w *Worker) runCleanup(ctx context.Context) {
	log := logging.GetLogger()
	cutoff := time.Now().Add(-w.config.TTL)

	for _, basePath := range w.config.Paths {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if basePath == "" {
			continue
		}

		// Robustness check: Ensure we don't scan root or dangerous paths if misconfigured
		cleanPath := filepath.Clean(basePath)
		if cleanPath == "/" || cleanPath == "." {
			log.Warn("Skipping dangerous GC path", "path", basePath)
			continue
		}

		log.Debug("GC: Scanning directory", "path", cleanPath)

		// ⚡ BOLT: Optimization to stream directory entries instead of loading all at once.
		// Randomized Selection from Top 5 High-Impact Targets
		// os.ReadDir reads the whole directory into a slice, which can cause OOM for large temporary directories.
		// We use os.Open + ReadDir(n) to process in chunks.

		f, err := os.Open(cleanPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			log.Error("GC: Failed to open directory", "path", cleanPath, "error", err)
			continue
		}

		// Ensure we close the file handle.
		// We use a closure to ensure Close is called before moving to the next path.
		func() {
			defer f.Close()

			const batchSize = 100
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				entries, err := f.ReadDir(batchSize)

				// Process entries even if there was an error (e.g. partial read)
				if len(entries) > 0 {
					for _, entry := range entries {
						fullPath := filepath.Join(cleanPath, entry.Name())
						info, err := entry.Info()
						if err != nil {
							continue
						}

						if info.ModTime().Before(cutoff) {
							log.Info("GC: Removing old item", "path", fullPath, "age", time.Since(info.ModTime()))
							if err := os.RemoveAll(fullPath); err != nil {
								log.Error("GC: Failed to remove item", "path", fullPath, "error", err)
							}
						}
					}
				}

				if err != nil {
					if err != io.EOF {
						log.Error("GC: Failed to read directory chunk", "path", cleanPath, "error", err)
					}
					return // Stop processing this directory on error or EOF
				}
			}
		}()
	}
}
