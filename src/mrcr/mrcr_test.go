package mrcr

import "testing"

func TestResolver(t *testing.T) {
	r := NewResolver()
	if !r.Resolve() {
		t.Errorf("Expected true")
	}
}
