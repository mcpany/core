// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"
)

func TestWatcher(t *testing.T) {
	// Create a temporary file to watch.
	file, err := os.CreateTemp("", "watcher_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(file.Name()) }()

	// Create a new watcher.
	w, err := NewWatcher()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	// Create a channel to signal when the file has been reloaded.
	var reloaded bool
	done := make(chan bool, 1)

	// Start watching the file.
	go func() {
		if err := w.Watch([]string{file.Name()}, func() {
			reloaded = true
			done <- true
		}); err != nil {
			// We might get "watch closed" error if test ends early, but ideally not
			t.Logf("Watch exited: %v", err)
		}
	}()

	// Give the watcher a moment to start up.
	time.Sleep(100 * time.Millisecond)

	// Write to the file to trigger a reload.
	if _, err := file.WriteString("test"); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	// Wait for the reload to complete.
	select {
	case <-done:
		// The file was reloaded successfully.
		if !reloaded {
			t.Error("expected reloaded to be true")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for file to be reloaded")
	}
}

func TestWatcher_AddError(t *testing.T) {
	w, err := NewWatcher()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	// Watching a non-existent file should error
	err = w.Watch([]string{"/non/existent/file"}, func() {})
	if err == nil {
		t.Fatal("Expected error watching non-existent file")
	}
}

func TestWatcher_URL(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	t.Setenv("MCPANY_TEST_MODE", "true")

	// Use a dynamic port to avoid conflicts
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close() // Close it, we just wanted a free port (or we could keep it open if we wanted to simulate a server)
	// Actually, Watch expects to fetch. If port is closed, it errors but shouldn't block.
	// That's what we want to test (non-blocking).

	w, err := NewWatcher()
	if err != nil {
		t.Fatal(err)
	}
	// We don't defer Close here because we want to test explicit Close affects Watch return

	done := make(chan bool)
	go func() {
		// Use local address to avoid external network dependency
		url := fmt.Sprintf("http://127.0.0.1:%d/config", port)
		_ = w.Watch([]string{url}, func() {})
		close(done)
	}()

	// Give it time to process (and potentially fail if it didn't skip URL)
	time.Sleep(100 * time.Millisecond)

	// Close should unblock Watch
	w.Close()

	select {
	case <-done:
		// Success: Watch returned
	case <-time.After(5 * time.Second):
		t.Fatal("Watch blocked after Close")
	}
}
