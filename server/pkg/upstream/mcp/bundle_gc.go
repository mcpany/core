package mcp

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mcpany/core/server/pkg/logging"
)

var (
	// activeBundles tracks the service IDs of currently active bundles.
	// This prevents the garbage collector from deleting directories that are in use.
	activeBundles sync.Map

	// lastGCTimestamp stores the Unix timestamp of the last garbage collection run.
	lastGCTimestamp atomic.Int64

	// gcInterval defines how often the garbage collector should run.
	gcInterval = 1 * time.Hour

	// bundleBaseDir is the base directory for MCP bundles.
	// It is variable to allow overriding in tests or via environment.
	bundleBaseDir = func() string {
		if dir := os.Getenv("MCP_BUNDLE_DIR"); dir != "" {
			return dir
		}
		return filepath.Join(os.TempDir(), "mcp-bundles")
	}()
)

// trackBundle marks a service ID as active.
func trackBundle(serviceID string) {
	activeBundles.Store(serviceID, true)
}

// untrackBundle marks a service ID as inactive.
func untrackBundle(serviceID string) {
	activeBundles.Delete(serviceID)
}

// triggerGC checks if it's time to run garbage collection and starts it in a
// background goroutine if necessary.
// triggerGC checks if it's time to run garbage collection and starts it in a
// background goroutine if necessary.
func triggerGC() {
	now := time.Now().Unix()
	last := lastGCTimestamp.Load()

	if now-last > int64(gcInterval.Seconds()) {
		if lastGCTimestamp.CompareAndSwap(last, now) {
			// Capture the current base directory to avoid race conditions
			// if the global variable is changed (e.g., in tests).
			dir := bundleBaseDir
			go runGC(dir)
		}
	}
}

// runGC scans the bundle base directory and removes any directories that do not
// correspond to an active service ID.
func runGC(baseDir string) {
	log := logging.GetLogger()
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		// If directory doesn't exist, nothing to clean
		if !os.IsNotExist(err) {
			log.Error("GC: Failed to read bundle directory", "dir", baseDir, "error", err)
		}
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		serviceID := entry.Name()
		if _, ok := activeBundles.Load(serviceID); !ok {
			// Found an orphaned directory
			path := filepath.Join(baseDir, serviceID)
			log.Info("GC: Removing orphaned bundle directory", "path", path)
			if err := os.RemoveAll(path); err != nil {
				log.Error("GC: Failed to remove orphaned directory", "path", path, "error", err)
			}
		}
	}
}
