// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package gc provides a background worker for global garbage collection of
// temporary files and directories.
package gc

import (
	"context"
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

		// We only scan the top-level directories in the list, but we recurse?
		// Usually temporary directories have a flat structure or predictable depth.
		// For safety, let's use WalkDir but be careful about crossing boundaries.
		// We WILL delete subdirectories if they are old.

		entries, err := os.ReadDir(cleanPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			log.Error("GC: Failed to read directory", "path", cleanPath, "error", err)
			continue
		}

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
}
