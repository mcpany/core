package validation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsValidURL(t *testing.T) {
	testCases := []struct {
		name     string
		rawURL   string
		expected bool
	}{
		// Valid URLs
		{"valid http", "http://example.com", true},
		{"valid https", "https://example.com/path?query=value#fragment", true},
		{"valid ftp", "ftp://user:pass@example.com/resource", true},
		{"valid 127.0.0.1 http", "http://127.0.0.1:8080", true},
		{"valid ip address http", "http://127.0.0.1/test", true},
		{"valid dns scheme simple", "dns:example.com", true},
		{"valid dns scheme full", "dns:///resolver.example.com/example.com", true},   // Host is "resolver.example.com"
		{"valid unix scheme with path", "unix:/tmp/socket.sock", true},               // No host, but path is present
		{"valid unix scheme with slashes and path", "unix:///tmp/socket.sock", true}, // No host, path is present
		{"valid passthrough scheme with path", "passthrough:///service-name", true},  // No host, path is present
		{"valid passthrough scheme opaque", "passthrough:service-name", true},        // No host, opaque is present
		{"valid mailto", "mailto:user@example.com", true},
		{"valid data url", "data:text/plain;base64,SGVsbG8sIFdvcmxkIQ==", true},

		// Invalid URLs
		{"empty string", "", false},
		{"only spaces", "   ", false},
		{"just scheme http", "http:", false},
		{"just scheme https with slashes", "https://", false}, // common web scheme, requires host
		{"missing scheme", "example.com", false},
		{"missing host for http", "http://", false}, // common web scheme, requires host
		{"http scheme with empty authority", "http:///", false},
		{"http scheme with empty authority and path", "http:///path", false},
		{"http scheme with port but no host", "http://:8080", false},
		{"scheme only custom", "customscheme:", false}, // No host, no opaque, no path
		{"url with internal spaces", "http://example.com/path with spaces", false},
		{"url with leading space", " http://example.com", false},
		{"url with trailing space", "http://example.com ", false},
		{"url with internal and surrounding spaces", " http://example.com/path with spaces ", false},
		// Base length of "http://example.com/" is 19. 2048 - 19 = 2029.
		{"very long url (just at limit)", "http://example.com/" + strings.Repeat("a", 2029), true},           // Total 2048
		{"excessively long url (just over limit)", "http://example.com/" + strings.Repeat("a", 2030), false}, // Total 2049
		{"no scheme but has slashes", "//example.com/path", false},
		{"invalid scheme chars", "ht!tp://example.com", false},
		{"dns scheme malformed (empty opaque/path)", "dns:", false},

		// Specific gRPC target cases
		{"grpc target no scheme", "127.0.0.1:50051", false},                       // Invalid: No scheme
		{"grpc target with scheme", "grpc://127.0.0.1:50051", true},               // Valid: Has scheme and host
		{"dns target for grpc", "dns:resolver.example.com/service", true},         // Valid: dns scheme with opaque part
		{"dns target for grpc full", "dns:///resolver.example.com/service", true}, // Valid: dns scheme with host and path
		{"unix target for grpc", "unix:/var/run/service.sock", true},              // Valid: unix scheme with path
		{"unix target for grpc full", "unix:///var/run/service.sock", true},       // Valid: unix scheme with path
		{"passthrough target for grpc", "passthrough:///my_service", true},        // Valid: passthrough scheme with path
		{"passthrough target opaque for grpc", "passthrough:my_service", true},    // Valid: passthrough scheme with opaque part
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := IsValidURL(tc.rawURL)
			if actual != tc.expected {
				t.Errorf("IsValidURL(%q) = %v; want %v", tc.rawURL, actual, tc.expected)
			}
		})
	}
}

func TestIsValidBindAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{"valid", "127.0.0.1:8080", false},
		{"no port", "127.0.0.1", true},
		{"no host", ":8080", false},
		{"empty", "", true},
		{"just colon", ":", true},
		{"multiple colons", "127.0.0.1:8080:8080", true},
		{"ipv6", "[::1]:8080", false},
		{"port only", "50050", false},
		{"invalid port negative", "127.0.0.1:-1", true},
		{"invalid port negative only", "-1", true},
		{"invalid port too large", "127.0.0.1:65536", true},
		{"invalid port too large only", "65536", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IsValidBindAddress(tt.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsValidBindAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	// Create a temporary file to test the case where the file exists.
	file, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer func() { _ = os.Remove(file.Name()) }()

	// Test case where the file exists.
	assert.NoError(t, FileExists(file.Name()))

	// Test case where the file does not exist.
	assert.Error(t, FileExists("non-existent-file"))
}

func TestValidateHTTPServiceDefinition(t *testing.T) {
	methodPost := configv1.HttpCallDefinition_HTTP_METHOD_POST
	methodGet := configv1.HttpCallDefinition_HTTP_METHOD_GET
	methodUnspecified := configv1.HttpCallDefinition_HTTP_METHOD_UNSPECIFIED

	testCases := []struct {
		name          string
		def           *configv1.HttpCallDefinition
		expectedError string
	}{
		{
			name: "valid definition",
			def: configv1.HttpCallDefinition_builder{
				EndpointPath: lo.ToPtr("/v1/users"),
				Method:       &methodPost,
			}.Build(),
			expectedError: "",
		},
		{
			name:          "nil definition",
			def:           nil,
			expectedError: "http call definition cannot be nil",
		},
		{
			name: "missing path",
			def: configv1.HttpCallDefinition_builder{
				Method: &methodGet,
			}.Build(),
			expectedError: "path is required",
		},
		{
			name: "path with only whitespace",
			def: configv1.HttpCallDefinition_builder{
				EndpointPath: lo.ToPtr("   "),
				Method:       &methodGet,
			}.Build(),
			expectedError: "path is required",
		},
		{
			name: "path does not start with slash",
			def: configv1.HttpCallDefinition_builder{
				EndpointPath: lo.ToPtr("v1/users"),
				Method:       &methodGet,
			}.Build(),
			expectedError: "path must start with a '/'",
		},
		{
			name: "unspecified method",
			def: configv1.HttpCallDefinition_builder{
				EndpointPath: lo.ToPtr("/v1/users"),
				Method:       &methodUnspecified,
			}.Build(),
			expectedError: "method is required",
		},
		{
			name: "path with query string",
			def: configv1.HttpCallDefinition_builder{
				EndpointPath: lo.ToPtr("/v1/users/{userId}?filter=name eq 'test'"),
				Method:       &methodGet,
			}.Build(),
			expectedError: "path must not contain query parameters",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateHTTPServiceDefinition(tc.def)
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestIsAllowedPath_AllowedDotDotPrefix(t *testing.T) {
	// Create a temporary directory for our "CWD"
	cwd, err := os.MkdirTemp("", "mcpany-cwd")
	require.NoError(t, err)
	defer os.RemoveAll(cwd)

	// Change CWD to the temp dir
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)
	err = os.Chdir(cwd)
	require.NoError(t, err)

	// Create a file named "..foo" in the CWD
	// This is a valid filename on most FS, and technically starts with ".." string
	// but is NOT a parent directory reference.
	err = os.Mkdir(filepath.Join(cwd, "..foo"), 0755)
	require.NoError(t, err)

	// This should PASS, but due to bug it FAILS
	err = IsAllowedPath("..foo")
	require.NoError(t, err, "IsAllowedPath should allow filenames starting with '..'")
}

func TestIsAllowedPath_Symlinks(t *testing.T) {
tmpDir := t.TempDir()

// Create a safe directory
safeDir := filepath.Join(tmpDir, "safe")
err := os.Mkdir(safeDir, 0755)
require.NoError(t, err)

// Create a secret directory outside safeDir
secretDir := filepath.Join(tmpDir, "secret")
err = os.Mkdir(secretDir, 0755)
require.NoError(t, err)

// Create a secret file
secretFile := filepath.Join(secretDir, "password.txt")
err = os.WriteFile(secretFile, []byte("s3cr3t"), 0600)
require.NoError(t, err)

// Create a symlink inside safeDir pointing to secretDir
symlinkPath := filepath.Join(safeDir, "link_to_secret")
err = os.Symlink(secretDir, symlinkPath)
require.NoError(t, err)

// Set allowed paths to ONLY safeDir
// Note: We need to temporarily set allowedPaths global variable
// We can use SetAllowedPaths for this.
defer SetAllowedPaths(nil)
SetAllowedPaths([]string{safeDir})

// 1. Accessing safeDir should be allowed
err = IsAllowedPath(safeDir)
require.NoError(t, err, "safeDir should be allowed")

// 2. Accessing a file in safeDir should be allowed
safeFile := filepath.Join(safeDir, "test.txt")
err = os.WriteFile(safeFile, []byte("test"), 0600)
require.NoError(t, err)
err = IsAllowedPath(safeFile)
require.NoError(t, err, "file in safeDir should be allowed")

// 3. Accessing secretDir should be DENIED (it's not in allowed paths)
err = IsAllowedPath(secretDir)
require.Error(t, err, "secretDir should be denied")

// 4. Accessing symlink inside safeDir pointing to secretDir should be DENIED
// even though the symlink is inside safeDir, it resolves to secretDir which is outside.
err = IsAllowedPath(symlinkPath)
require.Error(t, err, "symlink to secretDir should be denied")

// 5. Accessing file via symlink should be DENIED
fileViaSymlink := filepath.Join(symlinkPath, "password.txt")
err = IsAllowedPath(fileViaSymlink)
require.Error(t, err, "file via symlink should be denied")
}

func TestIsAllowedPath_SymlinkLoop(t *testing.T) {
tmpDir := t.TempDir()

// Create a loop
link1 := filepath.Join(tmpDir, "link1")
link2 := filepath.Join(tmpDir, "link2")

_ = os.Symlink(link2, link1)
_ = os.Symlink(link1, link2)

defer SetAllowedPaths(nil)
SetAllowedPaths([]string{tmpDir})

// Accessing loop should fail or timeout, or return error from EvalSymlinks
// We just want to make sure it doesn't crash or hang indefinitely (though EvalSymlinks handles loops)
err := IsAllowedPath(link1)
require.Error(t, err)
}
