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

func TestWatcher_FileRemoved(t *testing.T) {
	// Create a temporary file to watch.
	file, err := os.CreateTemp("", "watcher_test")
	if err != nil {
		t.Fatal(err)
	}

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

	// Remove the file.
	err = os.Remove(file.Name())
	if err != nil {
		t.Fatal(err)
	}

	// Wait for a moment to see if a reload is triggered.
	select {
	case <-reloaded:
		t.Fatal("file was reloaded after being removed")
	case <-time.After(2 * time.Second):
		// The file was not reloaded, which is the expected behavior.
	}
}
