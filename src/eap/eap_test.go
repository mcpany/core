package eap

import "testing"

func TestProvider(t *testing.T) {
	p := NewProvider()
	if !p.Bind() {
		t.Errorf("Expected true")
	}
}
