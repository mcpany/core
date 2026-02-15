package resource

import (
	"testing"
	"time"
)

func TestAddResourceDeadlock(t *testing.T) {
	m := NewManager()

	// Channel to signal that the callback completed
	done := make(chan bool)

	m.OnListChanged(func() {
		// Attempt to acquire a read lock inside the callback
		// This should cause a deadlock if the lock is held during the callback
		m.ListResources()
		done <- true
	})

	go func() {
		m.AddResource(&mockResource{uri: "test://resource"})
	}()

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Deadlock detected: OnListChanged callback failed to complete")
	}
}
