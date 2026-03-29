package arh

import (
	"testing"
)

func TestNewAutomatedRemediationHub(t *testing.T) {
	hub := NewAutomatedRemediationHub()
	if !hub.Enabled {
		t.Errorf("Expected AutomatedRemediationHub to be enabled by default")
	}
}
