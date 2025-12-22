package validation

import (
	"testing"
)

func TestIsSecurePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid file", "test.txt", false},
		{"valid path", "path/to/test.txt", false},
		{"valid with dots in name", "my..file.txt", false},
		{"valid with dots in dir name", "my..dir/file.txt", false},
		{"traversal up", "../test.txt", true},
		{"traversal up nested", "dir/../../test.txt", true},
		{"traversal up double", "../../test.txt", true},
		// Absolute paths are allowed, as long as they don't result in relative traversal up
		{"absolute path", "/etc/passwd", false},
		{"absolute path with resolved traversal", "/var/../etc/passwd", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IsSecurePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsSecurePath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}
