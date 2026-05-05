package gc_immune

import "testing"

func TestAnchor(t *testing.T) {
	a := NewAnchor()
	if !a.Pin() {
		t.Errorf("Expected true")
	}
}
