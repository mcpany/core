package tool

import (
	"testing"
)

func TestCheckContextualKeywords(t *testing.T) {
	// Testing internal function via public behavior or direct call since in same package
	tests := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		{"Safe import word", "importing", false},
		{"Dangerous import", "import os", true},
		{"Safe os word", "oscar", false},
		{"Dangerous os.system", "os.system('ls')", true},
		{"Dangerous os . system", "os . system", true},
		{"Safe string literal", "'import os'", false},
		{"Unsafe mixed", "print('hello'); import os", true},
		{"Dangerous subprocess", "subprocess.call", true},
		{"Safe subprocess word", "subprocess_call", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We simulate "python" which triggers checkInterpreterFunctionCalls -> checkContextualKeywords
			err := checkInterpreterFunctionCalls(tt.input, "python")
			if tt.shouldErr && err == nil {
				t.Errorf("Expected error for input %q, got nil", tt.input)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error for input %q: %v", tt.input, err)
			}
		})
	}
}
