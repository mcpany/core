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

func TestCheckForShellInjection_Comprehensive(t *testing.T) {
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
		// --- Shell Context (e.g. bash, sh) ---
		{"Shell Safe", "safe_value", "{{val}}", "{{val}}", "bash", true, false, ""},
		{"Shell Unquoted Semicolon", "ls;rm -rf /", "{{val}}", "{{val}}", "bash", true, true, "dangerous character"},
		{"Shell Unquoted Pipe", "ls|nc", "{{val}}", "{{val}}", "bash", true, true, "dangerous character"},
		{"Shell Unquoted Space", "ls -la", "{{val}}", "{{val}}", "bash", true, true, "dangerous character"}, // Space is blocked in unquoted shell
		{"Shell Double Quoted Safe", "safe value", `echo "{{val}}"`, "{{val}}", "bash", true, false, ""},
		{"Shell Double Quoted Dollar", "$VAR", `echo "{{val}}"`, "{{val}}", "bash", true, true, "dangerous character"},
		{"Shell Double Quoted Backtick", "`ls`", `echo "{{val}}"`, "{{val}}", "bash", true, true, "dangerous character"},
		{"Shell Double Quoted Quote", `foo"bar`, `echo "{{val}}"`, "{{val}}", "bash", true, true, "dangerous character"},
		// Updated expectation: Backticks are blocked in single quotes for strictness
		{"Shell Single Quoted Safe", "safe value $VAR `ls`", `echo '{{val}}'`, "{{val}}", "bash", true, true, "backtick"},
		{"Shell Single Quoted Quote", `foo'bar`, `echo '{{val}}'`, "{{val}}", "bash", true, true, "single quote"},

		// --- Python Interpreter ---
		{"Python Safe", "print('hello')", `python -c "{{val}}"`, "{{val}}", "python", false, false, ""},
		// Updated expectation: dangerous function call catches system before __import__ or generic check
		{"Python Dangerous Import", "__import__('os').system('ls')", `python -c "{{val}}"`, "{{val}}", "python", false, true, "dangerous function call"},
		{"Python Dangerous System", "import os; os.system('ls')", `python -c "{{val}}"`, "{{val}}", "python", false, true, "dangerous function call"},
		{"Python F-String Injection", "{self}", `python -c f"{{val}}"`, "{{val}}", "python", false, true, "f-string"},

		// --- Node.js Interpreter ---
		{"Node Safe", "console.log(1)", `node -e "{{val}}"`, "{{val}}", "node", false, false, ""},
		{"Node Dangerous ChildProcess", "require('child_process').exec('ls')", `node -e "{{val}}"`, "{{val}}", "node", false, true, "dangerous function call"},
		{"Node Template Literal", "${process.env}", `node -e ` + "`{{val}}`", "{{val}}", "node", false, true, "template literal"},

		// --- Ruby Interpreter ---
		{"Ruby Safe", "puts 1", `ruby -e "{{val}}"`, "{{val}}", "ruby", false, false, ""},
		{"Ruby Interpolation", "#{system('ls')}", `ruby -e "{{val}}"`, "{{val}}", "ruby", false, true, "ruby interpolation"},
		{"Ruby Open Injection", "|ls", `ruby -e "{{val}}"`, "{{val}}", "ruby", false, true, "ruby open injection"},

		// --- Perl Interpreter ---
		{"Perl Safe", "print 1", `perl -e "{{val}}"`, "{{val}}", "perl", false, false, ""},
		{"Perl qx Injection", "qx/ls/", `perl -e "{{val}}"`, "{{val}}", "perl", false, true, "perl qx"},
		{"Perl Array Interpolation", "@{ARGV}", `perl -e "{{val}}"`, "{{val}}", "perl", false, true, "perl array interpolation"},

		// --- PHP Interpreter ---
		{"PHP Safe", "echo 1;", `php -r "{{val}}"`, "{{val}}", "php", false, false, ""},
		// Updated expectation: generic double quote check catches $
		{"PHP Variable Interpolation", "$var", `php -r "{{val}}"`, "{{val}}", "php", false, true, "dangerous character"},
		{"PHP Exec", "exec('ls');", `php -r "{{val}}"`, "{{val}}", "php", false, true, "dangerous function call"},

		// --- AWK ---
		// Updated expectation: Awk with $ blocked by double quote check
		{"Awk Safe", "{print $1}", `awk "{{val}}"`, "{{val}}", "awk", false, true, "dangerous character"},
		{"Awk Pipe", "|sh", `awk "{{val}}"`, "{{val}}", "awk", false, true, "value contains '|'"},
		{"Awk System", "system('ls')", `awk "{{val}}"`, "{{val}}", "awk", false, true, "dangerous function call"},

		// --- SQL Clients (psql, mysql) ---
		{"SQL Safe", "SELECT * FROM users", `psql -c "{{val}}"`, "{{val}}", "psql", false, false, ""},
		{"SQL Injection Unquoted", "1 OR 1=1", `psql -c {{val}}`, "{{val}}", "psql", false, true, "SQL injection"},
		{"SQL Injection Comment", "admin' --", `psql -c {{val}}`, "{{val}}", "psql", false, true, "contains '--'"},

		// --- Tar ---
		{"Tar Safe", "archive.tar", `tar -cf {{val}}`, "{{val}}", "tar", false, false, ""},
		{"Tar Checkpoint Exec", "--checkpoint-action=exec=sh", `tar -cf {{val}}`, "{{val}}", "tar", false, true, "execution directive"},

		// --- Backtick Injection Generic ---
		{"Backtick Injection", "`ls`", `echo {{val}}`, "{{val}}", "bash", true, true, "dangerous character"},
		// Updated expectation
		{"Backtick Injection Quoted", "`ls`", "`{{val}}`", "{{val}}", "bash", true, true, "dangerous character"},

		// --- Obfuscation ---
		{"Obfuscated System", "sYsTeM ( 'ls' )", `python -c "{{val}}"`, "{{val}}", "python", false, true, "dangerous function call"},
		{"Obfuscated Comment", "system/*comment*/('ls')", `node -e "{{val}}"`, "{{val}}", "node", false, true, "dangerous function call"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkForShellInjection(tt.val, tt.template, tt.placeholder, tt.command, tt.isShell)
			if tt.wantErr {
				if assert.Error(t, err) {
					if tt.errContains != "" {
						assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.errContains))
					}
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateSafePathAndInjection(t *testing.T) {
	// Mock validation.IsSafeURL because external_test.go sabotages it globally
	originalIsSafeURL := validation.IsSafeURL
	defer func() { validation.IsSafeURL = originalIsSafeURL }()

	validation.IsSafeURL = func(urlStr string) error {
		if strings.HasPrefix(urlStr, "gopher://") {
			return fmt.Errorf("unsupported scheme: gopher")
		}
		if strings.HasPrefix(urlStr, "http://127.0.0.1") {
			return fmt.Errorf("loopback address is not allowed")
		}
		return nil
	}

	tests := []struct {
		name        string
		val         string
		isDocker    bool
		commandName string
		wantErr     bool
		errContains string
	}{
		{"Safe Path", "safe/path", false, "ls", false, ""},
		{"Traversal", "../etc/passwd", false, "ls", true, "path traversal"},
		{"Traversal Encoded", "%2e%2e/etc/passwd", false, "ls", true, "path traversal"},
		{"Absolute Path Local", "/etc/passwd", false, "ls", true, "absolute path detected"},
		{"Absolute Path Docker", "/etc/passwd", true, "ls", false, ""}, // Allowed in Docker
		{"File Scheme Local", "file:///etc/passwd", false, "ls", true, "file: scheme detected"},
		{"File Scheme Docker", "file:///etc/passwd", true, "ls", false, ""}, // Allowed in Docker? checkForLocalFileAccess is skipped if isDocker. But validateSafePathAndInjection calls it only if !isDocker.
		{"Arg Injection", "-rf", false, "rm", true, "argument injection"},
		{"Arg Injection Encoded", "%2drf", false, "rm", true, "argument injection"},
		{"Git Ext Scheme", "ext::sh -c touch%20/tmp/pwn", false, "git", true, "dangerous scheme"},
		{"Git Ext Encoded", "%65xt::sh", false, "git", true, "dangerous scheme"},
		{"ImageMagick Msl", "msl:/tmp/exploit.msl", false, "convert", true, "dangerous scheme"},
		{"Safe URL", "https://google.com", false, "curl", false, ""},
		{"Unsafe URL Scheme", "gopher://127.0.0.1:25", false, "curl", true, "unsafe url"},
		{"Unsafe URL IP", "http://127.0.0.1", false, "curl", true, "unsafe url"}, // Loopback blocked by IsSafeURL usually
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSafePathAndInjection(tt.val, tt.isDocker, tt.commandName)
			if tt.wantErr {
				if assert.Error(t, err) {
					if tt.errContains != "" {
						assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.errContains))
					}
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsDangerousEnvVar(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"LD_PRELOAD", true},
		{"DYLD_INSERT_LIBRARIES", true},
		{"GIT_CONFIG_PARAMETERS", true},
		{"PYTHONPATH", true},
		{"BASH_ENV", true},
		{"NODE_OPTIONS", true},
		{"JAVA_TOOL_OPTIONS", true},
		{"R_PROFILE_USER", true},
		{"ld_preload", true}, // Case insensitive
		{"GIT_CONFIG_KEY_0", true},
		{"SAFE_VAR", false},
		{"MY_APP_HOME", false},
		{"PATH", false}, // Usually allowed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDangerousEnvVar(tt.name)
			assert.Equal(t, tt.want, got, "isDangerousEnvVar(%q)", tt.name)
		})
	}
}

func TestCheckForDangerousSchemes(t *testing.T) {
	tests := []struct {
		val         string
		wantErr     bool
		errContains string
	}{
		{"safe:value", false, ""},
		{"http://google.com", false, ""},
		{"https://google.com", false, ""},
		{"file:///etc/passwd", true, "dangerous scheme"},
		{"File:///etc/passwd", true, "dangerous scheme"}, // Case insensitive
		{"gopher://127.0.0.1", true, "dangerous scheme"},
		{"expect://ls", true, "dangerous scheme"},
		{"php://input", true, "dangerous scheme"},
		{"zip://archive.zip", true, "dangerous scheme"},
		{"jar://file.jar", true, "dangerous scheme"},
		// ImageMagick
		{"mvg:/tmp/exploit.mvg", true, "dangerous scheme"},
		{"msl:/tmp/exploit.msl", true, "dangerous scheme"},
		{"vid:xwd:/tmp/pwn", true, "dangerous scheme"},
		{"label:Hello", true, "dangerous scheme"},
		// FFmpeg
		{"concat:file1|file2", true, "dangerous scheme"},
		{"subfile:start=0:end=1", true, "dangerous scheme"},
		{"hls://playlist.m3u8", true, "dangerous scheme"},
		// Git
		{"ext::sh -c touch /tmp/pwn", true, "dangerous scheme"},
		// Not schemes
		{"just text", false, ""},
		{"text with : colon", false, ""}, // Scheme must be start
		{"123:456", false, ""}, // Not alpha scheme
	}

	for _, tt := range tests {
		t.Run(tt.val, func(t *testing.T) {
			err := checkForDangerousSchemes(tt.val)
			if tt.wantErr {
				if assert.Error(t, err) {
					if tt.errContains != "" {
						assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.errContains))
					}
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
