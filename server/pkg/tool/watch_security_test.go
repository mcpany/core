package tool

import (
	"testing"
)

func TestWatchIsDetectedAsShell(t *testing.T) {
	if !isShellCommand("watch") {
		t.Errorf("VULNERABILITY: 'watch' is not detected as a shell command/runner. It executes arguments using sh -c, allowing Command Injection via arguments like '; id'.")
	}
	if !isShellCommand("tmux") {
		t.Errorf("VULNERABILITY: 'tmux' is not detected as a shell command/runner.")
	}
	if !isShellCommand("screen") {
		t.Errorf("VULNERABILITY: 'screen' is not detected as a shell command/runner.")
	}
}
