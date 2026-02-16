// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
		// 1. Unquoted Context (Level 0)
		{
			name:        "Unquoted Safe",
			val:         "safe_value",
			template:    "echo {{val}}",
			placeholder: "{{val}}",
			command:     "bash",
			isShell:     true,
			wantErr:     false,
		},
		{
			name:        "Unquoted Semicolon",
			val:         "val; rm -rf /",
			template:    "echo {{val}}",
			placeholder: "{{val}}",
			command:     "bash",
			isShell:     true,
			wantErr:     true,
			errContains: "shell injection detected",
		},
		{
			name:        "Unquoted Pipe",
			val:         "val | ls",
			template:    "echo {{val}}",
			placeholder: "{{val}}",
			command:     "bash",
			isShell:     true,
			wantErr:     true,
			errContains: "shell injection detected",
		},
		{
			name:        "Unquoted Backtick",
			val:         "`ls`",
			template:    "echo {{val}}",
			placeholder: "{{val}}",
			command:     "bash",
			isShell:     true,
			wantErr:     true,
			errContains: "shell injection detected",
		},
		{
			name:        "Unquoted Space",
			val:         "arg1 arg2",
			template:    "echo {{val}}",
			placeholder: "{{val}}",
			command:     "bash",
			isShell:     true,
			wantErr:     true,
			errContains: "shell injection detected",
		},

		// 2. Double Quoted Context (Level 1)
		{
			name:        "Double Quoted Safe",
			val:         "safe value",
			template:    "echo \"{{val}}\"",
			placeholder: "{{val}}",
			command:     "bash",
			isShell:     true,
			wantErr:     false,
		},
		{
			name:        "Double Quoted Backtick",
			val:         "`ls`",
			template:    "echo \"{{val}}\"",
			placeholder: "{{val}}",
			command:     "bash",
			isShell:     true,
			wantErr:     true,
			errContains: "shell injection detected",
		},
		{
			name:        "Double Quoted Dollar",
			val:         "$(ls)",
			template:    "echo \"{{val}}\"",
			placeholder: "{{val}}",
			command:     "bash",
			isShell:     true,
			wantErr:     true,
			errContains: "shell injection detected",
		},
		{
			name:        "Double Quoted Quote Breakout",
			val:         "\" ; rm -rf /",
			template:    "echo \"{{val}}\"",
			placeholder: "{{val}}",
			command:     "bash",
			isShell:     true,
			wantErr:     true,
			errContains: "shell injection detected",
		},

		// 3. Single Quoted Context (Level 2)
		{
			name:        "Single Quoted Safe",
			val:         "safe value $ \\", // Removed backtick as it is blocked
			template:    "echo '{{val}}'",
			placeholder: "{{val}}",
			command:     "bash",
			isShell:     true,
			wantErr:     false,
		},
		{
			name:        "Single Quoted Quote Breakout",
			val:         "' ; rm -rf /",
			template:    "echo '{{val}}'",
			placeholder: "{{val}}",
			command:     "bash",
			isShell:     true,
			wantErr:     true,
			errContains: "shell injection detected",
		},

		// 4. Backticked Context (Level 3) - Usually dangerous
		{
			name:        "Backticked Semicolon",
			val:         "val; rm -rf /",
			template:    "echo `{{val}}`",
			placeholder: "{{val}}",
			command:     "bash",
			isShell:     true,
			wantErr:     true,
			errContains: "shell injection detected",
		},

		// 5. Interpreter (Python)
		{
			name:        "Python F-String Injection",
			val:         "{__import__('os').system('ls')}",
			template:    "python -c \"print(f'{{val}}')\"",
			placeholder: "{{val}}",
			command:     "python",
			isShell:     false,
			wantErr:     true,
			errContains: "python f-string injection detected",
		},
		{
			name:        "Python Dangerous Function",
			val:         "__import__('os')",
			template:    "python -c '{{val}}'",
			placeholder: "{{val}}",
			command:     "python",
			isShell:     false,
			wantErr:     true,
			errContains: "interpreter injection detected",
		},

		// 6. Interpreter (Ruby)
		{
			name:        "Ruby Interpolation",
			val:         "#{system('ls')}",
			template:    "ruby -e \"puts '{{val}}'\"",
			placeholder: "{{val}}",
			command:     "ruby",
			isShell:     false,
			wantErr:     true,
			errContains: "ruby interpolation injection detected",
		},

		// 7. Interpreter (Node)
		{
			name:        "Node Template Literal",
			val:         "${process.exit()}",
			template:    "node -e `console.log(\\`{{val}}\\`)`",
			placeholder: "{{val}}",
			command:     "node",
			isShell:     false,
			wantErr:     true,
			errContains: "javascript template literal injection detected",
		},

		// 8. Interpreter (Perl)
		{
			name:        "Perl QX Unquoted",
			val:         "qx/ls/",
			template:    "perl -e {{val}}",
			placeholder: "{{val}}",
			command:     "perl",
			isShell:     false,
			wantErr:     true,
			errContains: "shell injection detected: perl qx execution",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkForShellInjection(tt.val, tt.template, tt.placeholder, tt.command, tt.isShell)
			if tt.wantErr {
				if assert.Error(t, err) {
					if tt.errContains != "" {
						assert.Contains(t, err.Error(), tt.errContains)
					}
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckTarInjection(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		val         string
		base        string
		wantErr     bool
		errContains string
	}{
		{
			name:    "Safe Tar",
			val:     "archive.tar",
			base:    "tar",
			wantErr: false,
		},
		{
			name:        "Tar Exec Directive",
			val:         "--checkpoint-action=exec=sh",
			base:        "tar",
			wantErr:     true,
			errContains: "tar injection detected",
		},
		{
			name:        "Tar Command Directive",
			val:         "--to-command=sh",
			base:        "tar",
			wantErr:     true,
			errContains: "tar injection detected",
		},
		{
			name:    "Not Tar",
			val:     "exec=sh",
			base:    "ls",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkTarInjection(tt.val, tt.base)
			if tt.wantErr {
				if assert.Error(t, err) {
					if tt.errContains != "" {
						assert.Contains(t, err.Error(), tt.errContains)
					}
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckSQLInjection(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		val         string
		base        string
		quoteLevel  int
		wantErr     bool
		errContains string
	}{
		{
			name:       "Safe SQL",
			val:        "valid_table",
			base:       "psql",
			quoteLevel: 0,
			wantErr:    false,
		},
		{
			name:        "SQL Injection OR",
			val:         "1 OR 1=1",
			base:        "psql",
			quoteLevel:  0,
			wantErr:     true,
			errContains: "SQL injection detected",
		},
		{
			name:        "SQL Injection DROP",
			val:         "users; DROP TABLE users",
			base:        "mysql",
			quoteLevel:  0,
			wantErr:     true,
			errContains: "SQL injection detected",
		},
		{
			name:        "SQL Comment",
			val:         "admin' --",
			base:        "sqlite3",
			quoteLevel:  0,
			wantErr:     true,
			errContains: "SQL injection detected",
		},
		{
			name:       "Quoted SQL (Safe-ish)",
			val:        "1 OR 1=1",
			base:       "psql",
			quoteLevel: 1, // Double quoted
			wantErr:    false,
		},
		{
			name:       "Not SQL Tool",
			val:        "DROP TABLE",
			base:       "echo",
			quoteLevel: 0,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkSQLInjection(tt.val, tt.base, tt.quoteLevel)
			if tt.wantErr {
				if assert.Error(t, err) {
					if tt.errContains != "" {
						assert.Contains(t, err.Error(), tt.errContains)
					}
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckInterpreterFunctionCalls(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		val         string
		lang        string
		wantErr     bool
		errContains string
	}{
		{
			name:    "Safe Code",
			val:     "print('hello')",
			lang:    "python",
			wantErr: false,
		},
		{
			name:        "System Call",
			val:         "system('ls')",
			lang:    "python",
			wantErr:     true,
			errContains: "interpreter injection detected",
		},
		{
			name:        "Exec Call",
			val:         "exec('ls')",
			lang:    "node",
			wantErr:     true,
			errContains: "interpreter injection detected",
		},
		{
			name:        "Spawn Call",
			val:         "require('child_process').spawn('ls')",
			lang:    "node",
			wantErr:     true,
			errContains: "interpreter injection detected",
		},
		{
			name:        "Eval Call",
			val:         "eval('code')",
			lang:    "php",
			wantErr:     true,
			errContains: "interpreter injection detected",
		},
		{
			name:        "Obfuscated System Call",
			val:         "system ( 'ls' )",
			lang:    "python",
			wantErr:     true,
			errContains: "interpreter injection detected",
		},
		{
			name:        "Ruby System No Parens",
			val:         "system 'ls'",
			lang:    "ruby",
			wantErr:     true,
			errContains: "interpreter injection detected",
		},
		{
			name:        "Import Check",
			val:         "__import__('os')",
			lang:    "python",
			wantErr:     true,
			errContains: "interpreter injection detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkInterpreterFunctionCalls(tt.val, tt.lang)
			if tt.wantErr {
				if assert.Error(t, err) {
					if tt.errContains != "" {
						assert.Contains(t, err.Error(), tt.errContains)
					}
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStripInterpreterComments(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		val  string
		lang string
		want string
	}{
		{
			name: "Python Comment",
			val:  "print('hi') # comment",
			lang: "python",
			want: "print('hi') ",
		},
		{
			name: "JS Line Comment",
			val:  "console.log('hi') // comment",
			lang: "node",
			want: "console.log('hi') ",
		},
		{
			name: "JS Block Comment",
			val:  "console.log('hi') /* comment */",
			lang: "node",
			want: "console.log('hi') ",
		},
		{
			name: "String with Hash",
			val:  "print('# not a comment')",
			lang: "python",
			want: "print('# not a comment')",
		},
		{
			name: "String with Slash",
			val:  "print('// not a comment')",
			lang: "node",
			want: "print('// not a comment')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripInterpreterComments(tt.val, tt.lang)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAnalyzeQuoteContext(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		template    string
		placeholder string
		wantLevel   int
	}{
		{
			name:        "Unquoted",
			template:    "echo {{val}}",
			placeholder: "{{val}}",
			wantLevel:   0,
		},
		{
			name:        "Double Quoted",
			template:    "echo \"{{val}}\"",
			placeholder: "{{val}}",
			wantLevel:   1,
		},
		{
			name:        "Single Quoted",
			template:    "echo '{{val}}'",
			placeholder: "{{val}}",
			wantLevel:   2,
		},
		{
			name:        "Backticked",
			template:    "echo `{{val}}`",
			placeholder: "{{val}}",
			wantLevel:   3,
		},
		{
			name:        "Mixed Double",
			template:    "echo \"prefix {{val}} suffix\"",
			placeholder: "{{val}}",
			wantLevel:   1,
		},
		{
			name:        "Escaped Quote",
			template:    "echo \"pre \\\" {{val}}\"",
			placeholder: "{{val}}",
			wantLevel:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzeQuoteContext(tt.template, tt.placeholder)
			assert.Equal(t, tt.wantLevel, got)
		})
	}
}
