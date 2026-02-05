package tool

import (
	"testing"
)

func TestInterpreterBypass(t *testing.T) {
	tests := []struct {
		name      string
		command   string
		template  string
		input     string
		shouldErr bool
	}{
		{
			name:      "Shell wrapping Python with double quotes",
			command:   "sh", // Treating it as shell command
			template:  "python -c 'print(\"{{input}}\")'",
			input:     "\"); import subprocess; subprocess.call([\"ls\"]); print(\"",
			shouldErr: true,
		},
		{
			name:      "Shell wrapping Python with double quotes - dangerous call",
			command:   "sh",
			template:  "python -c 'print(\"{{input}}\")'",
			input:     "\"); import os; os.system(\"ls\"); print(\"",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			placeholder := "{{input}}"

			err := checkForShellInjection(tt.input, tt.template, placeholder, tt.command)
			if tt.shouldErr && err == nil {
				t.Errorf("Expected error for input %q, but got nil", tt.input)
			} else if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error for input %q: %v", tt.input, err)
			}
		})
	}
}
