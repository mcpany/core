package config

import (
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
	defer os.Remove(file.Name())

	// Create a new watcher.
	watcher, err := NewWatcher()
	if err != nil {
		t.Fatal(err)
	}
	defer watcher.Close()

	// Create a channel to signal when the file has been reloaded.
	reloaded := make(chan bool, 1)

	// Start watching the file.
	go watcher.Watch([]string{file.Name()}, func() {
		reloaded <- true
	})

	// Give the watcher a moment to start up.
	time.Sleep(100 * time.Millisecond)

	// Write to the file to trigger a reload.
	_, err = file.WriteString("test")
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	// Wait for the reload to complete.
	select {
	case <-reloaded:
		// The file was reloaded successfully.
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for file to be reloaded")
	}
}
