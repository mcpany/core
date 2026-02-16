package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnalyzeQuoteContext(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		template    string
		placeholder string
		wantLevel   int
	}{
		{"Unquoted Simple", "echo {{msg}}", "{{msg}}", 0},
		{"Double Quoted", "echo \"{{msg}}\"", "{{msg}}", 1},
		{"Single Quoted", "echo '{{msg}}'", "{{msg}}", 2},
		{"Backticked", "echo `{{msg}}`", "{{msg}}", 3},
		{"Double Quoted with Escaped Quote", "echo \"foo \\\" {{msg}} \\\" bar\"", "{{msg}}", 1},
		// Standard shell does not support escaping single quotes inside single quotes.
		// 'foo \' closes the string at the second quote. The \ is a literal outside quotes.
		// {{msg}} is therefore unquoted.
		{"Single Quoted with Escaped Quote", "echo 'foo \\' {{msg}} \\' bar'", "{{msg}}", 0},
		{"Mixed Quotes Double Inside Single", "echo 'foo \"{{msg}}\" bar'", "{{msg}}", 2},
		{"Mixed Quotes Single Inside Double", "echo \"foo '{{msg}}' bar\"", "{{msg}}", 1},
		{"Complex Nested", "bash -c \"python -c 'print(\\\"{{msg}}\\\")'\"", "{{msg}}", 1}, // Outer is double
		{"Multiple Placeholders First", "echo {{msg}} {{other}}", "{{msg}}", 0},
		{"Multiple Placeholders Second", "echo {{msg}} {{other}}", "{{other}}", 0},
		{"Placeholder At Start", "{{msg}}", "{{msg}}", 0},
		{"Placeholder At End", "echo {{msg}}", "{{msg}}", 0},
		{"Empty Template", "", "{{msg}}", 0},
		{"Empty Placeholder", "template", "", 0},
		{"No Placeholder", "echo foo", "{{msg}}", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzeQuoteContext(tt.template, tt.placeholder)
			assert.Equal(t, tt.wantLevel, got)
		})
	}
}

func TestInterpreterInjection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		interpreter string // Base command name (e.g., "python", "ruby")
		quoteLevel int    // 0=Unquoted, 1=Double, 2=Single, 3=Backtick
		template   string // Template string (optional, for python checks)
		val        string // User input value
		wantErr    bool   // Expect error?
		errContains string // Expected error substring
	}{
		// --- Python ---
		{"Python Safe", "python", 0, "print('{{msg}}')", "hello", false, ""},
		{"Python F-String Injection", "python", 0, "print(f'{{msg}}')", "{open('file').read()}", true, "python f-string injection"},
		{"Python F-String Safe", "python", 0, "print(f'{{msg}}')", "hello", false, ""},
		{"Python No F-String", "python", 0, "print('{{msg}}')", "{hello}", false, ""}, // Not an f-string context in template

		// --- Ruby ---
		{"Ruby Safe", "ruby", 1, "", "hello", false, ""},
		{"Ruby Interpolation Double Quote", "ruby", 1, "", "#{system('ls')}", true, "ruby interpolation"},
		{"Ruby Interpolation Single Quote", "ruby", 2, "", "#{system('ls')}", false, ""}, // Ruby doesn't interpolate in single quotes
		{"Ruby Interpolation Backtick", "ruby", 3, "", "#{system('ls')}", true, "ruby interpolation"},
		// Ruby Pipe Injection is only checked by the helper at QuoteLevel 1 or 3.
		// At Level 0, it is handled by checkUnquotedInjection (blocked char '|').
		{"Ruby Pipe Injection Double", "ruby", 1, "", "|ls", true, "ruby open injection"},
		{"Ruby Pipe Injection Space Double", "ruby", 1, "", " | ls", true, "ruby open injection"},

		// --- Node/Perl/PHP ---
		{"Node Template Literal", "node", 3, "", "${process.env}", true, "javascript template literal"},
		{"Node Safe", "node", 1, "", "${safe}", false, ""}, // Node only interpolates in backticks (template literals)

		{"Perl Interpolation Double", "perl", 1, "", "${exec}", true, "variable interpolation"},
		{"Perl Interpolation Backtick", "perl", 3, "", "${exec}", true, "variable interpolation"},
		{"Perl QX Injection", "perl", 0, "", "qx/ls/", true, "perl qx execution"},
		{"Perl Array Interpolation", "perl", 1, "", "@{INC}", true, "perl array interpolation"},

		{"PHP Interpolation Double", "php", 1, "", "${exec}", true, "variable interpolation"},
		{"PHP Interpolation Single", "php", 2, "", "${exec}", false, ""}, // PHP doesn't interpolate in single quotes

		// --- Awk ---
		{"Awk Safe", "awk", 0, "", "print", false, ""},
		{"Awk Pipe", "awk", 0, "", "|sh", true, "awk injection"},
		{"Awk Redirect Out", "awk", 0, "", ">file", true, "awk injection"},
		{"Awk Redirect In", "awk", 0, "", "<file", true, "awk injection"},
		{"Awk Getline", "awk", 0, "", "getline", true, "awk injection"},

		// --- SQL ---
		{"SQL Safe", "psql", 0, "", "1", false, ""},
		{"SQL Injection OR", "psql", 0, "", "1 OR 1=1", true, "SQL injection"},
		{"SQL Injection DROP", "psql", 0, "", "DROP TABLE users", true, "SQL injection"},
		{"SQL Comment", "mysql", 0, "", "-- comment", true, "SQL injection"},
		{"SQL Safe Quoted", "psql", 1, "", "OR", false, ""}, // If quoted, keywords are strings

		// --- Tar ---
		{"Tar Safe", "tar", 0, "", "file.tar", false, ""},
		{"Tar Checkpoint", "tar", 0, "", "--checkpoint-action=exec=sh", true, "tar injection"},
		{"Tar To Command", "tar", 0, "", "--to-command=sh", true, "tar injection"}, // Checkpoint check covers exec=/command=

		// --- General Interpreter Function Calls ---
		{"Python System Call", "python", 1, "", "system('ls')", true, "dangerous function call"},
		{"Python Import", "python", 1, "", "__import__('os')", true, "dangerous function call"},
		{"Ruby Eval", "ruby", 1, "", "eval('ls')", true, "dangerous function call"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkInterpreterInjection(tt.val, tt.template, tt.interpreter, tt.quoteLevel)
			// Also check function calls if applicable (logic inside checkForShellInjection calls checkInterpreterFunctionCalls)
			// But checkInterpreterInjection is a standalone helper.
			// Wait, checkInterpreterInjection calls sub-checks.
			// checkInterpreterFunctionCalls is separate in checkForShellInjection.
			// BUT, for the "General Interpreter Function Calls" cases, we might need to invoke checkInterpreterFunctionCalls directly or wrap it.
			// Let's rely on what checkInterpreterInjection covers.
			// Looking at code: checkInterpreterInjection calls specific checks.
			// checkInterpreterFunctionCalls is NOT called by checkInterpreterInjection.
			// So for the last 3 cases, we should expect NO error from checkInterpreterInjection alone,
			// unless we also test checkInterpreterFunctionCalls.

			// Let's split the test or update expectation.
			// Ideally, we test checkForShellInjection which calls both.
			// But since we are testing unit helper, let's just test checkInterpreterInjection logic here.

			if tt.name == "Python System Call" || tt.name == "Python Import" || tt.name == "Ruby Eval" {
				// These are caught by checkInterpreterFunctionCalls, not checkInterpreterInjection
				return
			}

			if tt.wantErr {
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), tt.errContains)
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
		interpreter string
		wantErr     bool
	}{
		{"Safe", "print('hello')", "python", false},
		{"System Call", "system('ls')", "python", true},
		{"Exec Call", "exec('ls')", "python", true},
		{"Import Call", "__import__('os')", "python", true},
		{"Obfuscated System", "s y s t e m ( 'ls' )", "python", true},
		{"Ruby System No Paren", "system 'ls'", "ruby", true},
		{"Safe Word", "systematic", "python", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkInterpreterFunctionCalls(tt.val, tt.interpreter)
			if tt.wantErr {
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), "interpreter injection detected")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckForShellInjection_Complex(t *testing.T) {
	t.Parallel()
	// Integration tests for the main function

	tests := []struct {
		name        string
		val         string
		template    string // Determines quote level
		placeholder string
		command     string
		isShell     bool
		wantErr     bool
		errContains string
	}{
		// --- Shell Context ---
		{"Shell Unquoted Dangerous", "; ls", "echo {{msg}}", "{{msg}}", "sh", true, true, "dangerous character"},
		{"Shell Unquoted Safe", "hello", "echo {{msg}}", "{{msg}}", "sh", true, false, ""},
		{"Shell Single Quoted Dangerous", "foo'bar", "echo '{{msg}}'", "{{msg}}", "sh", true, true, "single quote"},
		{"Shell Double Quoted Dangerous", "foo\"bar", "echo \"{{msg}}\"", "{{msg}}", "sh", true, true, "dangerous character"}, // " is dangerous
		{"Shell Double Quoted Var", "$HOME", "echo \"{{msg}}\"", "{{msg}}", "sh", true, true, "dangerous character"}, // $ is dangerous

		// --- Interpreter Context (Python) ---
		{"Python F-String", "{open('/etc/passwd')}", "python -c \"print(f'{{msg}}')\"", "{{msg}}", "python", false, true, "python f-string"},
		{"Python Function Call", "system('ls')", "python -c \"print('{{msg}}')\"", "{{msg}}", "python", false, true, "dangerous function call"},

		// --- Tar ---
		{"Tar Checkpoint", "--checkpoint-action=exec=sh", "tar cf archive.tar {{files}}", "{{files}}", "tar", false, true, "tar injection"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkForShellInjection(tt.val, tt.template, tt.placeholder, tt.command, tt.isShell)
			if tt.wantErr {
				assert.Error(t, err)
				if err != nil && tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
