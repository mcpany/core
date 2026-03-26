package browser

import (

	"testing"
)

// We won't test playwrightFetcher directly if we don't have playwright installed
// in this environment. But we can write tests for NewProvider.
func TestNewProvider(t *testing.T) {
	p := NewProvider()
	if p == nil {
		t.Errorf("NewProvider returned nil")
	}
}
