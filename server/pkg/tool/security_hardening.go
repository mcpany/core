// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"path/filepath"
	"strings"
)

// IdentifyDangerousInterpreterArgs returns a set of indices in args that are
// considered "code" arguments for known interpreters.
// Substituting into these arguments allows code injection vulnerabilities.
func IdentifyDangerousInterpreterArgs(command string, args []string) map[int]bool {
	dangerous := make(map[int]bool)
	base := strings.ToLower(filepath.Base(command))

	isPython := strings.HasPrefix(base, "python")
	isRuby := strings.HasPrefix(base, "ruby")
	isPerl := strings.HasPrefix(base, "perl")
	isNode := strings.HasPrefix(base, "node") || base == "nodejs" || base == "bun" || base == "deno"
	isPhp := strings.HasPrefix(base, "php")
	isLua := strings.HasPrefix(base, "lua")
	isShell := base == "sh" || base == "bash" || base == "zsh" || base == "dash" || base == "ash" || base == "ksh" || base == "csh" || base == "tcsh" || base == "fish" || base == "pwsh" || base == "powershell" || base == "cmd" || base == "cmd.exe"
	isAwk := strings.HasPrefix(base, "awk") || strings.HasPrefix(base, "gawk") || strings.HasPrefix(base, "nawk") || strings.HasPrefix(base, "mawk")
	isSed := base == "sed" || base == "gsed"

	// Iterate args to find execution flags
	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Check for combined flags (e.g. -cCode)
		if len(arg) > 2 && strings.HasPrefix(arg, "-") {
			// Python, Bash, Sh: -c
			if (isPython || isShell) && strings.HasPrefix(arg, "-c") {
				dangerous[i] = true
				continue
			}
			// Ruby, Perl, Lua: -e
			if (isRuby || isPerl || isLua) && strings.HasPrefix(arg, "-e") {
				dangerous[i] = true
				continue
			}
			// Node: -e, -p
			if isNode && (strings.HasPrefix(arg, "-e") || strings.HasPrefix(arg, "-p")) {
				dangerous[i] = true
				continue
			}
			// PHP: -r
			if isPhp && strings.HasPrefix(arg, "-r") {
				dangerous[i] = true
				continue
			}
		}

		// Check for separate flags
		if arg == "-c" {
			if isPython || isShell {
				if i+1 < len(args) {
					dangerous[i+1] = true
				}
			}
		} else if arg == "-e" {
			if isRuby || isPerl || isLua || isNode || isSed {
				if i+1 < len(args) {
					dangerous[i+1] = true
				}
			}
		} else if arg == "-E" {
			if isPerl {
				if i+1 < len(args) {
					dangerous[i+1] = true
				}
			}
		} else if arg == "-p" || arg == "--print" {
			if isNode {
				if i+1 < len(args) {
					dangerous[i+1] = true
				}
			}
		} else if arg == "--eval" {
			if isNode {
				if i+1 < len(args) {
					dangerous[i+1] = true
				}
			}
		} else if arg == "-r" {
			if isPhp {
				if i+1 < len(args) {
					dangerous[i+1] = true
				}
			}
		}
	}

	if isAwk || isSed {
		// Find first non-flag argument
		for i := 0; i < len(args); i++ {
			arg := args[i]
			if strings.HasPrefix(arg, "-") {
				// If it takes an argument, skip next
				// sed -e script (handled above)
				// awk -f file
				if arg == "-f" || arg == "--file" {
					i++
				}
				continue
			}
			// First non-flag is script
			// For sed, if -e was used, subsequent non-flags are input files.
			// But if no -e used, first non-flag is script.
			// How to know if -e was used?
			// We can track it.

			// Simplified: If we haven't marked any dangerous args yet (via -e),
			// then this is likely the script.
			alreadyMarked := false
			for j := 0; j < i; j++ {
				if dangerous[j] {
					alreadyMarked = true
					break
				}
			}

			if !alreadyMarked {
				dangerous[i] = true
			}
			break
		}
	}

	return dangerous
}
