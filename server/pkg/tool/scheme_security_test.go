// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
)

func TestIsVulnerableToSchemes(t *testing.T) {
	tests := []struct {
		command string
		want    bool
	}{
		// ImageMagick
		{"convert", true},
		{"mogrify", true},
		{"identify", true},
		{"magick", true},
		{"/usr/bin/convert", true},

		// FFmpeg
		{"ffmpeg", true},
		{"ffprobe", true},
		{"ffplay", true},

		// Git
		{"git", true},
		{"/usr/local/bin/git", true},

		// Safe/Unknown
		{"echo", false},
		{"ls", false},
		{"python", false},
		{"node", false},
		{"my-tool", false},
		{"converter", false}, // substring match check
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			got := isVulnerableToSchemes(tt.command)
			assert.Equal(t, tt.want, got, "isVulnerableToSchemes(%q)", tt.command)
		})
	}
}

func TestCheckForDangerousSchemes(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		// Safe inputs
		{"simple.txt", false},
		{"/path/to/file", false},
		{"http://example.com", false},
		{"https://example.com", false},
		{"ftp://example.com", true}, // ftp is blocked
		{"mailto:user@example.com", false}, // mailto not explicitly blocked
		{"just a colon:", false},
		{"12:34", false},

		// Dangerous schemes (Generic/Interpreter)
		{"file:///etc/passwd", true},
		{"FILE:///etc/passwd", true}, // Case insensitive
		{"gopher://127.0.0.1", true},
		{"expect://ls", true},
		{"php://input", true},
		{"zip://archive.zip", true},
		{"jar:file:///tmp/app.jar", true},
		{"war:file:///tmp/app.war", true},

		// ImageMagick schemes
		{"mvg:exploits.mvg", true},
		{"msl:exploits.msl", true},
		{"vid:exploits.vid", true},
		{"ephemeral:/tmp/test", true},
		{"label:Testing", true},
		{"text:Testing", true},
		{"info:Testing", true},
		{"pango:Testing", true},
		{"caption:Testing", true},
		{"plasma:Testing", true},
		{"xc:Testing", true},
		{"inline:base64data", true},
		{"gradient:red-blue", true},
		{"pattern:checkerboard", true},
		{"tile:pattern:checkerboard", true},
		{"read:image.png", true},

		// FFmpeg schemes
		{"concat:file1.ts|file2.ts", true},
		{"subfile:start=0:end=1", true},
		{"crypto:input.enc", true},
		{"data:image/png;base64,...", true},
		{"hls:http://example.com", true},
		{"rtmp://server/live", true},
		{"rtsp://server/live", true},

		// Git
		{"ext::sh -c touch%20/tmp/pwn", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			err := checkForDangerousSchemes(tt.input)
			if tt.wantErr {
				assert.Error(t, err, "checkForDangerousSchemes(%q) should error", tt.input)
				if err != nil {
					assert.Contains(t, err.Error(), "dangerous scheme detected", "Error message should mention dangerous scheme")
				}
			} else {
				assert.NoError(t, err, "checkForDangerousSchemes(%q) should not error", tt.input)
			}
		})
	}
}

func TestValidateSafePathAndInjection_Schemes(t *testing.T) {
	// Mock IsSafeURL to pass for http/https, fail for others, to isolate scheme check
	originalIsSafeURL := validation.IsSafeURL
	defer func() { validation.IsSafeURL = originalIsSafeURL }()
	validation.IsSafeURL = func(urlStr string) error {
		if strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://") {
			return nil
		}
		return fmt.Errorf("unsafe url")
	}

	tests := []struct {
		val         string
		command     string
		expectError bool
	}{
		// Vulnerable command + Dangerous scheme
		{"file:///etc/passwd", "convert", true},
		{"file:///etc/passwd", "ffmpeg", true},
		{"ext::sh -c whoami", "git", true},

		// Vulnerable command + Safe scheme
		{"https://example.com/image.png", "convert", false}, // Safe (http/https allowed, validated by IsSafeURL if ://)
		{"input.jpg", "convert", false},

		// Safe command + Dangerous scheme
		// Note: "file:///etc/passwd" contains "://" so validateSafePathAndInjection calls IsSafeURL.
		// Our mock IsSafeURL fails for file://
		{"file:///etc/passwd", "echo", true}, // Blocked by IsSafeURL

		// Dangerous scheme without :// (e.g. mvg:exploits.mvg)
		// IsSafeURL is NOT called because no "://".
		// Scheme check should block it for vulnerable command.
		{"mvg:exploits.mvg", "convert", true},
		{"mvg:exploits.mvg", "echo", false}, // Not blocked for echo

		// URL Encoded dangerous scheme
		{"file%3A%2F%2F%2Fetc%2Fpasswd", "convert", true},

		// For echo (safe command), "file%3A%2F%2F%2Fetc%2Fpasswd"
		// It fails because checkForLocalFileAccess blocks "file:" scheme globally, even if encoded.
		{"file%3A%2F%2F%2Fetc%2Fpasswd", "echo", true},

		// Encoded scheme without ://
		{"mvg%3Aexploits.mvg", "convert", true},
		{"mvg%3Aexploits.mvg", "echo", false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.command, tt.val), func(t *testing.T) {
			// We assume isDocker=false for strict checking
			err := validateSafePathAndInjection(tt.val, false, tt.command)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateSafePathAndInjection_SSRF(t *testing.T) {
	// Tests specifically for the SSRF protection logic triggered by "://"
	// We must mock IsSafeURL to enforce strict checking, overriding TestMain's permissive mock.
	originalIsSafeURL := validation.IsSafeURL
	defer func() { validation.IsSafeURL = originalIsSafeURL }()

	// Mock that blocks anything not http/https or local IPs
	validation.IsSafeURL = func(urlStr string) error {
		if strings.HasPrefix(urlStr, "ftp://") || strings.HasPrefix(urlStr, "gopher://") || strings.HasPrefix(urlStr, "file://") {
			return fmt.Errorf("bad scheme")
		}
		if strings.Contains(urlStr, "127.0.0.1") || strings.Contains(urlStr, "localhost") || strings.Contains(urlStr, "[::1]") || strings.Contains(urlStr, "169.254") {
			return fmt.Errorf("local/private ip")
		}
		return nil
	}

    tests := []struct {
        val         string
        expectError bool
    }{
        {"http://example.com", false},
        {"https://google.com", false},
        {"ftp://example.com", true}, // Scheme not allowed
        {"gopher://example.com", true}, // Scheme not allowed
        {"file:///etc/passwd", true}, // Scheme not allowed
        {"http://169.254.169.254/latest/meta-data/", true}, // Private IP (AWS metadata)
        {"http://127.0.0.1:8080", true}, // Loopback
        {"http://[::1]", true}, // IPv6 Loopback
        {"http://localhost", true}, // Localhost
    }

    for _, tt := range tests {
        t.Run(tt.val, func(t *testing.T) {
            // Command name doesn't matter for general SSRF check on ://
            err := validateSafePathAndInjection(tt.val, false, "echo")
            if tt.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
