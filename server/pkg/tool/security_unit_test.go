// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckSQLInjection(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		val        string
		base       string
		quoteLevel int
		wantErr    bool
	}{
		// Safe Cases
		{name: "safe simple", val: "id", base: "psql", quoteLevel: 0, wantErr: false},
		{name: "safe quoted", val: "DROP TABLE", base: "psql", quoteLevel: 1, wantErr: false}, // Quoted is safe from injection in this context (usually)
		{name: "not sql tool", val: "UNION SELECT", base: "grep", quoteLevel: 0, wantErr: false},

		// SQL Keywords
		{name: "UNION", val: "UNION", base: "psql", quoteLevel: 0, wantErr: true},
		{name: "SELECT", val: "SELECT", base: "mysql", quoteLevel: 0, wantErr: true},
		{name: "DROP TABLE", val: "DROP TABLE users", base: "sqlite3", quoteLevel: 0, wantErr: true},
		{name: "OR 1=1", val: "1 OR 1=1", base: "psql", quoteLevel: 0, wantErr: true},
		{name: "AND 1=1", val: "1 AND 1=1", base: "psql", quoteLevel: 0, wantErr: true},
		{name: "comment", val: "--", base: "psql", quoteLevel: 0, wantErr: true},
		{name: "comment with text", val: "admin --", base: "psql", quoteLevel: 0, wantErr: true},

		// Boundary Checks
		{name: "keyword middle", val: "foo UNION bar", base: "psql", quoteLevel: 0, wantErr: true},
		{name: "keyword start", val: "UNION bar", base: "psql", quoteLevel: 0, wantErr: true},
		{name: "keyword end", val: "foo UNION", base: "psql", quoteLevel: 0, wantErr: true},
		{name: "keyword part of word", val: "REUNION", base: "psql", quoteLevel: 0, wantErr: false},
		{name: "keyword part of word 2", val: "UNIONIZED", base: "psql", quoteLevel: 0, wantErr: false},

		// Case Insensitivity
		{name: "lower case", val: "union select", base: "psql", quoteLevel: 0, wantErr: true},
		{name: "mixed case", val: "uNiOn SeLeCt", base: "psql", quoteLevel: 0, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkSQLInjection(tt.val, tt.base, tt.quoteLevel)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "SQL injection detected")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckInterpreterInjection(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		val        string
		template   string
		base       string
		quoteLevel int
		wantErr    bool
	}{
		// Python
		{name: "python f-string safe", val: "foo", template: "print(f'{foo}')", base: "python", quoteLevel: 2, wantErr: false},
		{name: "python f-string injection", val: "{__import__('os').system('ls')}", template: "print(f'{foo}')", base: "python", quoteLevel: 2, wantErr: true},
		{name: "python no f-string", val: "{foo}", template: "print('{foo}')", base: "python", quoteLevel: 2, wantErr: false},

		// Ruby
		{name: "ruby safe", val: "foo", base: "ruby", quoteLevel: 1, wantErr: false},
		{name: "ruby interpolation", val: "#{system('ls')}", base: "ruby", quoteLevel: 1, wantErr: true},
		{name: "ruby open pipe", val: "|ls", base: "ruby", quoteLevel: 1, wantErr: true},
		{name: "ruby safe pipe", val: "foo|bar", base: "ruby", quoteLevel: 1, wantErr: false}, // Leading pipe is key

		// Node/JS
		{name: "node safe", val: "foo", base: "node", quoteLevel: 3, wantErr: false},
		{name: "node template injection", val: "${process.exit(1)}", base: "node", quoteLevel: 3, wantErr: true},

		// Perl
		{name: "perl qx", val: "qx/ls/", base: "perl", quoteLevel: 1, wantErr: true},
		{name: "perl array", val: "@{foo}", base: "perl", quoteLevel: 1, wantErr: true},

		// Awk
		{name: "awk safe", val: "print", base: "awk", quoteLevel: 0, wantErr: false},
		{name: "awk pipe", val: "|sh", base: "awk", quoteLevel: 0, wantErr: true},
		{name: "awk redirect", val: ">file", base: "awk", quoteLevel: 0, wantErr: true},
		{name: "awk getline", val: "getline", base: "awk", quoteLevel: 0, wantErr: true},

		// Tar
		{name: "tar safe", val: "archive.tar", base: "tar", quoteLevel: 0, wantErr: false},
		{name: "tar exec", val: "--to-command=sh", base: "tar", quoteLevel: 0, wantErr: true},
		{name: "tar checkpoint", val: "--checkpoint-action=exec=sh", base: "tar", quoteLevel: 0, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkInterpreterInjection(tt.val, tt.template, tt.base, tt.quoteLevel)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckShellInjection(t *testing.T) {
	t.Parallel()
	t.Run("checkBacktickInjection", func(t *testing.T) {
		tests := []struct {
			name    string
			val     string
			command string
			wantErr bool
		}{
			{name: "safe", val: "foo", command: "bash", wantErr: false},
			{name: "backtick injection", val: "`ls`", command: "bash", wantErr: true},
			{name: "semicolon", val: "; ls", command: "bash", wantErr: true},
			{name: "pipe", val: "| ls", command: "bash", wantErr: true},
			{name: "safe node", val: "foo", command: "node", wantErr: false},
			{name: "safe node backtick", val: "template literal", command: "node", wantErr: false}, // Node allows backticks if it's template literal context (caller ensures this)
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := checkBacktickInjection(tt.val, tt.command)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("checkUnquotedInjection", func(t *testing.T) {
		tests := []struct {
			name    string
			val     string
			command string
			isShell bool
			wantErr bool
		}{
			{name: "safe", val: "foo", command: "ls", isShell: false, wantErr: false},
			{name: "semicolon", val: ";ls", command: "ls", isShell: false, wantErr: true},
			{name: "space in shell", val: "foo bar", command: "bash", isShell: true, wantErr: true},
			{name: "space not in shell", val: "foo bar", command: "grep", isShell: false, wantErr: false}, // Exec handles splitting
			{name: "env var assignment", val: "FOO=bar", command: "env", isShell: false, wantErr: true},
			{name: "quote injection", val: "foo\"bar", command: "ls", isShell: false, wantErr: true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := checkUnquotedInjection(tt.val, tt.command, tt.isShell)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

func TestCheckForDangerousSchemes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		val     string
		wantErr bool
	}{
		{name: "safe http", val: "http://example.com", wantErr: false},
		{name: "safe https", val: "https://example.com", wantErr: false},
		{name: "safe text", val: "just some text", wantErr: false},
		{name: "file scheme", val: "file:///etc/passwd", wantErr: true},
		{name: "gopher scheme", val: "gopher://bad.com", wantErr: true},
		{name: "php scheme", val: "php://filter", wantErr: true},
		{name: "mvg scheme", val: "mvg:payload", wantErr: true},
		{name: "label scheme", val: "label:@/etc/passwd", wantErr: true},
		{name: "git ext", val: "ext::sh -c touch%20/tmp/pwn", wantErr: true},
		{name: "case insensitive", val: "FiLe:///ETC/PASSWD", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkForDangerousSchemes(tt.val)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "dangerous scheme detected")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStripInterpreterComments(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		language string
		want     string
	}{
		{name: "bash hash comment", input: "ls # comment", language: "bash", want: "ls "},
		{name: "python hash comment", input: "import os # comment", language: "python", want: "import os "},
		{name: "c slash comment", input: "int main() // comment", language: "c", want: "int main() "},
		{name: "c block comment", input: "int /* comment */ main()", language: "c", want: "int  main()"},
		{name: "mixed comments", input: "ls # comment\nls // comment", language: "php", want: "ls \nls "},
		{name: "quoted hash", input: "echo '# not a comment'", language: "bash", want: "echo '# not a comment'"},
		{name: "quoted slash", input: "echo '// not a comment'", language: "c", want: "echo '// not a comment'"},
		{name: "escaped quote", input: "echo \"escaped \\\" quote # comment\"", language: "bash", want: "echo \"escaped \\\" quote # comment\""},
		{name: "nested block", input: "/* /* nested? */ */", language: "c", want: " */"}, // C comments don't nest usually
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripInterpreterComments(tt.input, tt.language)
			assert.Equal(t, tt.want, got)
		})
	}
}
