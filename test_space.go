package main

import (
	"fmt"
	"strings"
	"path/filepath"
)

func isInterpreter(command string) bool { return true }
func checkInterpreterInjection(val, template, base string, quoteLevel int) error { return nil }
func checkInterpreterFunctionCalls(val, base string) error { return nil }
func checkArgumentInterpreterInjection(val, template, base string, quoteLevel int, isShell bool) error { return nil }
func isSafeBacktickLanguage(command string) bool { return false }

func checkUnquotedInjection(val, command string, isShell bool) error {
	dangerousChars := ";|&$`(){}!<>\"\n\r\t\v\f*?[]~#%^'\\"
	if isShell {
		dangerousChars += " "
	}

	charsToCheck := dangerousChars
	if filepath.Base(command) == "env" {
		charsToCheck += "="
	}

	if idx := strings.IndexAny(val, charsToCheck); idx != -1 {
		return fmt.Errorf("shell injection detected: value contains dangerous character %q", val[idx])
	}
	return nil
}

func analyzeQuoteContext(template, placeholder string) int {
	return 0
}

func checkForShellInjection(val string, template string, placeholder string, command string, isShell bool) error {
	quoteLevel := analyzeQuoteContext(template, placeholder)

	base := strings.ToLower(filepath.Base(command))
	isWindowsCmd := base == "cmd.exe" || base == "cmd"
	if isWindowsCmd && quoteLevel == 2 {
		quoteLevel = 0
	}

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

	if quoteLevel == 3 {
        return fmt.Errorf("backtick")
	}

	if quoteLevel == 2 {
		if strings.Contains(val, "'") {
			return fmt.Errorf("shell injection detected: value contains single quote which breaks out of single-quoted argument")
		}
		if !isShell && strings.Contains(val, "\\") {
			return fmt.Errorf("interpreter injection detected: value contains backslash inside single-quoted argument")
		}
		if strings.Contains(val, "`") {
			return fmt.Errorf("shell injection detected: value contains backtick inside single-quoted argument (potential interpreter abuse)")
		}
		return nil
	}

	if quoteLevel == 1 {
		if idx := strings.IndexAny(val, "\"$`\\%"); idx != -1 {
			return fmt.Errorf("shell injection detected: value contains dangerous character %q inside double-quoted argument", val[idx])
		}
		return nil
	}

	return checkUnquotedInjection(val, command, isShell)
}

func main() {
    err := checkForShellInjection("Hello World", "{{arg}}", "{{arg}}", "python3", false)
    fmt.Println(err)
}
