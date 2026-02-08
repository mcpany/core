package main

import (
	"os"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "MissingName",
			args:    []string{"connector-runtime"},
			wantErr: true,
		},
		{
			name:    "ValidName",
			args:    []string{"connector-runtime", "-name", "test-connector"},
			wantErr: false,
		},
		{
			name:    "ValidNameSidecar",
			args:    []string{"connector-runtime", "-name", "test-connector", "-sidecar=true"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stop := make(chan os.Signal, 1)

			// Simulate signal after short delay if expecting success
			if !tt.wantErr {
				go func() {
					time.Sleep(10 * time.Millisecond)
					stop <- os.Interrupt
				}()
			}

			err := run(tt.args, stop)
			if (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
