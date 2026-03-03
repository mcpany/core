// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"
)

func TestCheckAwkInjection(t *testing.T) {
	tests := []struct {
		name    string
		val     string
		base    string
		wantErr bool
	}{
		{
			name:    "Safe awk input",
			val:     "print $1",
			base:    "awk",
			wantErr: false,
		},
		{
			name:    "Awk with pipe",
			val:     "print $1 | \"sh\"",
			base:    "awk",
			wantErr: true,
		},
		{
			name:    "Awk with redirection >",
			val:     "print $1 > \"/etc/passwd\"",
			base:    "gawk",
			wantErr: true,
		},
		{
			name:    "Awk with redirection <",
			val:     "print $1 < \"/etc/passwd\"",
			base:    "nawk",
			wantErr: true,
		},
		{
			name:    "Awk with getline",
			val:     "getline var < \"/etc/passwd\"",
			base:    "mawk",
			wantErr: true,
		},
		{
			name:    "Awk with getline inside string",
			val:     "print \"getline is fine here\"",
			base:    "awk",
			wantErr: false,
		},
		{
			name:    "Awk with getline inside comment",
			val:     "print $1 # getline here is fine\n print $2",
			base:    "awk",
			wantErr: false,
		},
		{
			name:    "Awk with indirect function call",
			val:     "BEGIN { f=\"system\"; @f(\"id\") }",
			base:    "awk",
			wantErr: true,
		},
		{
			name:    "Awk with @ inside string",
			val:     "print \"user@example.com\"",
			base:    "awk",
			wantErr: true, // Function logic in checkAwkInjection explicitly blocks any "@" anywhere currently (strings.Contains)
		},
		{
			name:    "Awk with @ inside comment",
			val:     "print $1 # contact user@example.com",
			base:    "awk",
			wantErr: true, // Function logic in checkAwkInjection explicitly blocks any "@" anywhere currently (strings.Contains)
		},
		{
			name:    "Awk with escaped quote and getline",
			val:     "print \"\\\"\" getline",
			base:    "awk",
			wantErr: true, // Quote is escaped, so getline is outside string
		},
		{
			name:    "Awk with escaped backslash before quote",
			val:     "print \"\\\\\" | \"sh\"",
			base:    "awk",
			wantErr: true, // \\ is escaped backslash, so quote ends string, pipe is outside
		},
		{
			name:    "Non-awk base command",
			val:     "system(\"ls\") | @",
			base:    "bash",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkAwkInjection(tt.val, tt.base); (err != nil) != tt.wantErr {
				t.Errorf("checkAwkInjection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckBacktickInjection(t *testing.T) {
	tests := []struct {
		name    string
		val     string
		command string
		wantErr bool
	}{
		{
			name:    "Safe node input without backticks",
			val:     "console.log('Hello')",
			command: "node",
			wantErr: false, // Node is safe, doesn't contain backticks
		},
		{
			name:    "Safe node input with backticks",
			val:     "console.log(`Hello`)",
			command: "node",
			wantErr: true, // Node is safe, but string contains backticks! breaking out of backticks!
		},
		{
			name:    "Unsafe bash input with backticks",
			val:     "echo `ls -la`",
			command: "bash",
			wantErr: true,
		},
		{
			name:    "Unsafe bash input with dangerous chars but no backticks",
			val:     "echo hello world", // contains space
			command: "bash",
			wantErr: true,
		},
		{
			name:    "Unsafe bash input completely safe chars",
			val:     "safeword",
			command: "bash",
			wantErr: false,
		},
		{
			name:    "Bun input with dangerous chars",
			val:     "const a = 'hello world';",
			command: "/usr/local/bin/bun",
			wantErr: false, // Bun is safe, doesn't contain backticks
		},
		{
			name:    "Deno input with backticks",
			val:     "```",
			command: "/opt/deno",
			wantErr: true, // Deno is safe, but contains backticks
		},
		{
			name:    "Nodejs input without backticks",
			val:     "var x = 1; x++;",
			command: "nodejs",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkBacktickInjection(tt.val, tt.command); (err != nil) != tt.wantErr {
				t.Errorf("checkBacktickInjection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
