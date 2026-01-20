package util

import (
	"net"
	"net/http"
	"testing"
)

func TestGetClientIP_XFFWithBrackets(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "[::1]")

	// trustProxy = true
	got := GetClientIP(req, true)
	expected := "::1"

	if got != expected {
		t.Errorf("GetClientIP() = %q, want %q", got, expected)
	}

	if net.ParseIP(got) == nil {
		t.Errorf("net.ParseIP(%q) returned nil, indicating invalid IP format returned by GetClientIP", got)
	}
}

func TestGetClientIP_XFFWithPort(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4:1234")

	// trustProxy = true
	got := GetClientIP(req, true)
	expected := "1.2.3.4"

	if got != expected {
		t.Errorf("GetClientIP() = %q, want %q", got, expected)
	}

	if net.ParseIP(got) == nil {
		t.Errorf("net.ParseIP(%q) returned nil, indicating invalid IP format returned by GetClientIP", got)
	}
}

func TestExtractIP_Zone(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"fe80::1%eth0", "fe80::1"},
		{"[fe80::1%eth0]", "fe80::1"},
		{"[fe80::1%eth0]:12345", "fe80::1"},
	}

	for _, tt := range tests {
		got := ExtractIP(tt.input)
		if got != tt.expected {
			t.Errorf("ExtractIP(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
