package mcpserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsTextMime(t *testing.T) {
	tests := []struct {
		mimeType string
		want     bool
	}{
		{"text/plain", true},
		{"text/html", true},
		{"text/plain; charset=utf-8", true},
		{"application/json", true},
		{"application/json; charset=utf-8", true},
		{"application/xml", true},
		{"application/octet-stream", false},
		{"image/png", false},
		{"application/pdf", false},
		{"application/x-yaml", true},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			assert.Equal(t, tt.want, isTextMime(tt.mimeType))
		})
	}
}
