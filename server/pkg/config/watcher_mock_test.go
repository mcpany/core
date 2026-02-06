package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockWatcher(t *testing.T) {
	w := NewMockWatcher()

	calledWatch := false
	w.WatchFunc = func(_ []string, _ func()) {
		calledWatch = true
	}
	w.CloseFunc = func() {
	}

	// Test Watch
	assert.NotPanics(t, func() {
		_ = w.Watch([]string{"/tmp"}, func() {})
	})
	assert.True(t, calledWatch)

	// Test Close
	assert.NotPanics(t, func() {
		w.Close()
	})
}
