package main

import (
	"fmt"
	"strings"
	"path/filepath"
)

func main() {
    cmd := "python3"
    // isInterpreter is true for python3
    // isShell is false
    // quoteLevel is 0

    val := "Hello World"

    err := checkForShellInjection(val, "", "", cmd, false)
    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Println("OK")
    }
}

func checkForShellInjection(val string, template string, placeholder string, command string, isShell bool) error {
	// Determine the quoting context of the placeholder in the template
	quoteLevel := analyzeQuoteContext(template, placeholder)

	base := strings.ToLower(filepath.Base(command))
	isWindowsCmd := base == "cmd.exe" || base == "cmd"
	if isWindowsCmd && quoteLevel == 2 {
		quoteLevel = 0
	}

	// Check if the main command is an interpreter
	if isInterpreter(command) {
		if err := checkInterpreterInjection(val, template, base, quoteLevel); err != nil {
			return err
		}
		if quoteLevel == 0 || quoteLevel == 1 || quoteLevel == 2 {
			if err := checkInterpreterFunctionCalls(val, base); err != nil {
				return err
			}
		}
	}

	if err := checkArgumentInterpreterInjection(val, template, base, quoteLevel, isShell); err != nil {
		return err
	}

	if quoteLevel == 3 { // Backticked
		return fmt.Errorf("backticked")
	}

	if quoteLevel == 2 { // Single Quoted
        return fmt.Errorf("single quoted")
	}

	if quoteLevel == 1 { // Double Quoted
        return fmt.Errorf("double quoted")
	}

	return checkUnquotedInjection(val, command, isShell)
}

func checkUnquotedInjection(val, command string, isShell bool) error {
	dangerousChars := ";|&$`(){}!<>\"\n\r\t\v\f*?[]~#%^'\\"
	if isShell {
		dangerousChars += " "
	}

	charsToCheck := dangerousChars
	// For 'env' command, '=' is dangerous as it allows setting arbitrary environment variables
	if filepath.Base(command) == "env" {
		charsToCheck += "="
	}

	if idx := strings.IndexAny(val, charsToCheck); idx != -1 {
		return fmt.Errorf("shell injection detected: value contains dangerous character %q", val[idx])
	}
	return nil
}

func analyzeQuoteContext(template, placeholder string) int {
	if template == "" || placeholder == "" {
		return 0
	}
    return 0
}

func isInterpreter(command string) bool { return true }
func checkInterpreterInjection(val, template, base string, quoteLevel int) error { return nil }
func checkInterpreterFunctionCalls(val, base string) error { return nil }
func checkArgumentInterpreterInjection(val, template, base string, quoteLevel int, isShell bool) error { return nil }
