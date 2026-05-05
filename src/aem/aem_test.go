package aem

import "testing"

func TestMonitor(t *testing.T) {
	m := NewMonitor()
	if m.Score() != 100 {
		t.Errorf("Expected 100")
	}
}
