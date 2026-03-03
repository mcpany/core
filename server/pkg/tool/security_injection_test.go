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

func TestCheckGdbInjection(t *testing.T) {
	tests := []struct {
		name       string
		val        string
		base       string
		quoteLevel int
		wantErr    bool
	}{
		{
			name:       "Not GDB base",
			val:        "shell whoami",
			base:       "bash",
			quoteLevel: 0,
			wantErr:    false,
		},
		{
			name:       "Safe GDB input",
			val:        "print var",
			base:       "gdb",
			quoteLevel: 0,
			wantErr:    false,
		},
		{
			name:       "Single quoted GDB input",
			val:        "shell whoami",
			base:       "gdb",
			quoteLevel: 2,
			wantErr:    false, // Single quoted is considered literal
		},
		{
			name:       "Empty value",
			val:        "  ",
			base:       "gdb",
			quoteLevel: 0,
			wantErr:    false,
		},
		{
			name:       "Dangerous command: shell",
			val:        "shell ls -la",
			base:       "gdb",
			quoteLevel: 0,
			wantErr:    true,
		},
		{
			name:       "Dangerous command: SYSTEM",
			val:        "SYSTEM whoami",
			base:       "gdb",
			quoteLevel: 1, // Double quoted might still be command
			wantErr:    true,
		},
		{
			name:       "Dangerous command: pipe",
			val:        "pipe bash",
			base:       "gdb",
			quoteLevel: 0,
			wantErr:    true,
		},
		{
			name:       "Dangerous command: make",
			val:        "make all",
			base:       "gdb",
			quoteLevel: 0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkGdbInjection(tt.val, tt.base, tt.quoteLevel)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkGdbInjection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckJqInjection(t *testing.T) {
	tests := []struct {
		name       string
		val        string
		base       string
		quoteLevel int
		wantErr    bool
	}{
		{
			name:       "Not JQ base",
			val:        "env",
			base:       "python",
			quoteLevel: 0,
			wantErr:    false,
		},
		{
			name:       "Inside string literal (Double Quotes)",
			val:        "env",
			base:       "jq",
			quoteLevel: 1,
			wantErr:    false,
		},
		{
			name:       "Safe jq command",
			val:        ".foo | .bar",
			base:       "jq",
			quoteLevel: 0,
			wantErr:    false,
		},
		{
			name:       "Dangerous keyword: env",
			val:        ".foo | env",
			base:       "jq",
			quoteLevel: 0,
			wantErr:    true,
		},
		{
			name:       "Dangerous keyword: input",
			val:        "input",
			base:       "jq",
			quoteLevel: 2, // Single quoted (checked anyway)
			wantErr:    true,
		},
		{
			name:       "Dangerous keyword: import",
			val:        "import \"foo\"",
			base:       "jq",
			quoteLevel: 0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkJqInjection(tt.val, tt.base, tt.quoteLevel)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkJqInjection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckTarInjection(t *testing.T) {
	tests := []struct {
		name    string
		val     string
		base    string
		wantErr bool
	}{
		{
			name:    "Not tar base",
			val:     "exec=bash",
			base:    "gzip",
			wantErr: false,
		},
		{
			name:    "Safe tar flag",
			val:     "--verbose",
			base:    "tar",
			wantErr: false,
		},
		{
			name:    "Execution directive: exec=",
			val:     "--checkpoint-action=exec=bash",
			base:    "tar",
			wantErr: true,
		},
		{
			name:    "Execution directive: command=",
			val:     "--to-command=bash",
			base:    "gtar",
			wantErr: true,
		},
		{
			name:    "Keyword: checkpoint-action",
			val:     "--checkpoint-action",
			base:    "bsdtar",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkTarInjection(tt.val, tt.base)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkTarInjection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
