// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateSafePathAndInjection_Extended(t *testing.T) {
	// Cannot use t.Parallel() because we modify global state (validation.IsSafeURL)

	// Mock IsSafeURL with a strict implementation for this test,
	// because TestMain mocks it to be permissive globally.
	originalIsSafeURL := validation.IsSafeURL
	defer func() {
		validation.IsSafeURL = originalIsSafeURL
	}()

	validation.IsSafeURL = func(urlStr string) error {
		u, err := url.Parse(urlStr)
		if err != nil {
			return err
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return fmt.Errorf("unsupported scheme: %s", u.Scheme)
		}
		if strings.Contains(u.Host, "localhost") || strings.Contains(u.Host, "127.0.0.1") || strings.Contains(u.Host, "169.254.169.254") {
			return fmt.Errorf("unsafe host: %s", u.Host)
		}
		return nil
	}

	// Ensure secure defaults for testing, overriding any environment variables set by CI/Makefile
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "false")
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "false")
	t.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES", "false")

	tests := []struct {
		name        string
		input       string
		isDocker    bool
		commandName string
		wantErr     bool
		errContains string
	}{
		// Whitespace Bypass
		{
			name:        "Whitespace Bypass Path Traversal",
			input:       "  /../etc/passwd",
			isDocker:    false,
			commandName: "cat",
			wantErr:     true,
			errContains: "path traversal attempt detected",
		},
		{
			name:        "Whitespace Bypass Argument Injection",
			input:       "  -rf",
			isDocker:    false,
			commandName: "rm",
			wantErr:     true,
			errContains: "argument injection detected",
		},
		// SSRF / URL Schemes
		{
			name:        "SSRF File Scheme",
			input:       "file:///etc/passwd",
			isDocker:    false,
			commandName: "curl",
			wantErr:     true,
			// Caught by IsSafeURL because it contains ://
			errContains: "unsupported scheme: file",
		},
		{
			name:        "SSRF Gopher Scheme",
			input:       "gopher://127.0.0.1:25/",
			isDocker:    false,
			commandName: "curl",
			wantErr:     true,
			errContains: "unsafe url argument",
		},
		{
			name:        "SSRF FTP Scheme",
			input:       "ftp://127.0.0.1/",
			isDocker:    false,
			commandName: "wget",
			wantErr:     true,
			errContains: "unsafe url argument",
		},
		{
			name:        "SSRF Localhost",
			input:       "http://localhost:8080",
			isDocker:    false,
			commandName: "curl",
			wantErr:     true,
			errContains: "unsafe url argument",
		},
		{
			name:        "SSRF Localhost IP",
			input:       "http://127.0.0.1:8080",
			isDocker:    false,
			commandName: "curl",
			wantErr:     true,
			errContains: "unsafe url argument",
		},
		{
			name:        "SSRF Metadata IP",
			input:       "http://169.254.169.254/latest/meta-data/",
			isDocker:    false,
			commandName: "curl",
			wantErr:     true,
			errContains: "unsafe url argument",
		},
		// Dangerous Schemes for Vulnerable Tools
		{
			name:        "Git Ext Protocol",
			input:       "ext::sh -c touch /tmp/pwned",
			isDocker:    false,
			commandName: "git",
			wantErr:     true,
			errContains: "dangerous scheme detected: ext",
		},
		{
			name:        "ImageMagick Label Scheme",
			input:       "label:@/etc/passwd",
			isDocker:    false,
			commandName: "convert",
			wantErr:     true,
			errContains: "dangerous scheme detected: label",
		},
		{
			name:        "FFmpeg Concat Scheme",
			input:       "concat:http://attacker.com/file1.ts|file2.ts",
			isDocker:    false,
			commandName: "ffmpeg",
			wantErr:     true,
			// Caught by IsSafeURL because it contains :// (http://)
			// But note: IsSafeURL parses the scheme as "concat".
			errContains: "unsupported scheme: concat",
		},
		// Path Traversal Variants
		{
			name:        "Encoded Path Traversal",
			input:       "%2e%2e%2fetc%2fpasswd",
			isDocker:    false,
			commandName: "cat",
			wantErr:     true,
			errContains: "path traversal attempt detected",
		},
		{
			name:        "Double Encoded Path Traversal",
			input:       "%252e%252e%252fetc%252fpasswd",
			isDocker:    false,
			commandName: "cat",
			wantErr:     true,
			errContains: "path traversal attempt detected", // validateSafePathAndInjection decodes once, checkPathTraversal handles re-encoding or we catch decoded
		},
		{
			name:        "Mixed Encoding Path Traversal",
			input:       "..%2fetc%2fpasswd",
			isDocker:    false,
			commandName: "cat",
			wantErr:     true,
			errContains: "path traversal attempt detected",
		},
		// Argument Injection Variants
		{
			name:        "Argument Injection Single Dash",
			input:       "-option",
			isDocker:    false,
			commandName: "ls",
			wantErr:     true,
			errContains: "argument injection detected",
		},
		{
			name:        "Argument Injection Double Dash",
			input:       "--option",
			isDocker:    false,
			commandName: "ls",
			wantErr:     true,
			errContains: "argument injection detected",
		},
		{
			name:        "Argument Injection Encoded Dash",
			input:       "%2doption",
			isDocker:    false,
			commandName: "ls",
			wantErr:     true,
			errContains: "argument injection detected",
		},
		{
			name:        "Argument Injection Plus",
			input:       "+option",
			isDocker:    false,
			commandName: "vim", // Some tools treat + as flag
			wantErr:     true,
			errContains: "argument injection detected",
		},
		{
			name:        "Safe Number Argument",
			input:       "-10",
			isDocker:    false,
			commandName: "math",
			wantErr:     false,
		},
		{
			name:        "Safe Positive Number Argument",
			input:       "+10",
			isDocker:    false,
			commandName: "math",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSafePathAndInjection(tt.input, tt.isDocker, tt.commandName)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckForShellInjection_Extended(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		val         string
		template    string
		placeholder string
		command     string
		isShell     bool
		wantErr     bool
		errContains string
	}{
		// Obfuscation / Polyglot
		{
			name:        "Null Byte Injection",
			val:         "safe\x00; rm -rf /",
			template:    "echo {{input}}",
			placeholder: "{{input}}",
			command:     "sh",
			isShell:     true,
			wantErr:     true,
			// \x00 is not explicitly in dangerousChars for Unquoted check, but ; is.
			// Let's test standard shell metachars.
			errContains: "shell injection detected",
		},
		{
			name:        "New Line Injection",
			val:         "safe\nrm -rf /",
			template:    "echo {{input}}",
			placeholder: "{{input}}",
			command:     "sh",
			isShell:     true,
			wantErr:     true,
			errContains: "shell injection detected",
		},
		{
			name:        "Tab Injection",
			val:         "safe\trm -rf /",
			template:    "echo {{input}}",
			placeholder: "{{input}}",
			command:     "sh",
			isShell:     true,
			wantErr:     true,
			errContains: "shell injection detected",
		},
		// Python Interpreter Injection
		{
			name:        "Python __import__",
			val:         "__import__('os')",
			template:    "print('{{input}}')",
			placeholder: "{{input}}",
			command:     "python",
			isShell:     false, // Interpreter
			wantErr:     true,
			errContains: "interpreter injection detected: value contains '__import__'",
		},
		{
			name:        "Python exec()",
			val:         "exec('import os; os.system(\"id\")')",
			template:    "print('{{input}}')",
			placeholder: "{{input}}",
			command:     "python",
			isShell:     false,
			wantErr:     true,
			errContains: "interpreter injection detected: dangerous keyword \"exec\" followed by",
		},
		{
			name:        "Python eval()",
			val:         "eval('__import__(\"os\").system(\"id\")')",
			template:    "print('{{input}}')",
			placeholder: "{{input}}",
			command:     "python",
			isShell:     false,
			wantErr:     true,
			errContains: "interpreter injection detected: dangerous keyword \"eval\" followed by",
		},
		{
			name:        "Python F-String Injection",
			val:         "{__import__('os').system('id')}",
			template:    "print(f'Hello {{input}}')",
			placeholder: "{{input}}",
			command:     "python",
			isShell:     false,
			wantErr:     true,
			errContains: "python f-string injection detected",
		},
		// Ruby Interpreter Injection
		{
			name:        "Ruby System Call",
			val:         "system('id')",
			template:    "puts '{{input}}'",
			placeholder: "{{input}}",
			command:     "ruby",
			isShell:     false,
			wantErr:     true,
			// Updated expectation: The code now detects "system" as a dangerous keyword first.
			errContains: "interpreter injection detected: dangerous keyword \"system\" found (unquoted)",
		},
		{
			name:        "Ruby Backtick",
			val:         "`id`",
			template:    "puts '{{input}}'",
			placeholder: "{{input}}",
			command:     "ruby",
			isShell:     false,
			wantErr:     true,
			errContains: "shell injection detected: value contains backtick inside single-quoted argument",
		},
		{
			name:        "Ruby Interpolation",
			val:         "#{system('id')}",
			template:    "puts \"{{input}}\"",
			placeholder: "{{input}}",
			command:     "ruby",
			isShell:     false,
			wantErr:     true,
			errContains: "ruby interpolation injection detected",
		},
		// Perl Interpreter Injection
		{
			name:        "Perl System Call",
			val:         "system('id')",
			template:    "print '{{input}}'",
			placeholder: "{{input}}",
			command:     "perl",
			isShell:     false,
			wantErr:     true,
			// Updated expectation: The code now detects "system" as a dangerous keyword first.
			errContains: "interpreter injection detected: dangerous keyword \"system\" found (unquoted)",
		},
		{
			name:        "Perl qx Operator",
			val:         "qx/id/",
			template:    "print \"{{input}}\"",
			placeholder: "{{input}}",
			command:     "perl",
			isShell:     false,
			wantErr:     true,
			errContains: "shell injection detected: perl qx execution",
		},
		// Nested Interpreter (Shell -> Python)
		{
			name:        "Nested Python Injection",
			val:         "__import__('os').system('id')",
			template:    "python -c 'print(\"{{input}}\")'",
			placeholder: "{{input}}",
			command:     "sh",
			isShell:     true,
			wantErr:     true,
			errContains: "argument interpreter injection detected (python)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkForShellInjection(tt.val, tt.template, tt.placeholder, tt.command, tt.isShell)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFastJSONNumber_LargeIntegers(t *testing.T) {
	t.Parallel()

	// JSON with a large integer that exceeds float64 precision (max integer exact is 2^53 - 1 approx 9e15)
	// 9223372036854775807 is MaxInt64 (approx 9e18)
	// float64 only has 53 bits of mantissa.
	jsonInput := []byte(`{"large_int": 9223372036854775807}`)

	var result map[string]interface{}
	err := fastJSONNumber.Unmarshal(jsonInput, &result)
	require.NoError(t, err)

	val, ok := result["large_int"]
	require.True(t, ok)

	// Since UseNumber is true, it should be json.Number
	num, ok := val.(json.Number)
	require.True(t, ok, "Expected json.Number, got %T", val)

	// Verify we can parse it as int64 correctly
	intVal, err := num.Int64()
	require.NoError(t, err)
	assert.Equal(t, int64(9223372036854775807), intVal)

	// Verify string representation matches exactly
	assert.Equal(t, "9223372036854775807", num.String())
}
