// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"path/filepath"
	"strings"
)

type interpreterConfig struct {
	execFlags         map[string]bool
	combinedFlags     []string
	scriptIsFirstArg  bool
	skipFlags         map[string]bool
}

// IdentifyDangerousInterpreterArgs returns a set of indices in args that are
// considered "code" arguments for known interpreters.
// Substituting into these arguments allows code injection vulnerabilities.
func IdentifyDangerousInterpreterArgs(command string, args []string) map[int]bool {
	dangerous := make(map[int]bool)
	base := strings.ToLower(filepath.Base(command))

	// Define interpreter configurations
	interpreters := map[string]interpreterConfig{
		"python": {
			execFlags:     map[string]bool{"-c": true},
			combinedFlags: []string{"-c"},
		},
		"ruby": {
			execFlags:     map[string]bool{"-e": true},
			combinedFlags: []string{"-e"},
		},
		"perl": {
			execFlags:     map[string]bool{"-e": true, "-E": true},
			combinedFlags: []string{"-e"},
		},
		"node": {
			execFlags:     map[string]bool{"-e": true, "--eval": true, "-p": true, "--print": true},
			combinedFlags: []string{"-e", "-p"},
		},
		"php": {
			execFlags:     map[string]bool{"-r": true},
			combinedFlags: []string{"-r"},
		},
		"lua": {
			execFlags:     map[string]bool{"-e": true},
			combinedFlags: []string{"-e"},
		},
		"shell": { // sh, bash, etc.
			execFlags:     map[string]bool{"-c": true},
			combinedFlags: []string{"-c"},
		},
		"awk": {
			scriptIsFirstArg: true,
			skipFlags:        map[string]bool{"-f": true, "--file": true},
		},
		"sed": {
			execFlags:        map[string]bool{"-e": true},
			scriptIsFirstArg: true,
			skipFlags:        map[string]bool{"-f": true, "--file": true},
		},
	}

	// Determine which config to use
	var config interpreterConfig
	found := false

	switch {
	case strings.HasPrefix(base, "python"):
		config = interpreters["python"]
		found = true
	case strings.HasPrefix(base, "ruby"):
		config = interpreters["ruby"]
		found = true
	case strings.HasPrefix(base, "perl"):
		config = interpreters["perl"]
		found = true
	case strings.HasPrefix(base, "node") || base == "nodejs" || base == "bun" || base == "deno":
		config = interpreters["node"]
		found = true
	case strings.HasPrefix(base, "php"):
		config = interpreters["php"]
		found = true
	case strings.HasPrefix(base, "lua"):
		config = interpreters["lua"]
		found = true
	case isShell(base):
		config = interpreters["shell"]
		found = true
	case strings.HasPrefix(base, "awk") || strings.HasPrefix(base, "gawk") || strings.HasPrefix(base, "nawk") || strings.HasPrefix(base, "mawk"):
		config = interpreters["awk"]
		found = true
	case base == "sed" || base == "gsed":
		config = interpreters["sed"]
		found = true
	}

	if !found {
		return dangerous
	}

	// Check for flags
	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Check separate flags
		if config.execFlags != nil && config.execFlags[arg] {
			if i+1 < len(args) {
				dangerous[i+1] = true
			}
			continue
		}

		// Check combined flags
		for _, prefix := range config.combinedFlags {
			if len(arg) > len(prefix) && strings.HasPrefix(arg, prefix) {
				dangerous[i] = true
				break
			}
		}
	}

	// Check for script argument (first non-flag) if applicable
	if config.scriptIsFirstArg {
		// Find first non-flag argument
		for i := 0; i < len(args); i++ {
			arg := args[i]
			if strings.HasPrefix(arg, "-") {
				// If it takes an argument, skip next
				if config.skipFlags != nil && config.skipFlags[arg] {
					i++
				}
				continue
			}

			// Only mark if we haven't already marked this index as dangerous via a flag (e.g. sed -e script)
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

func isShell(base string) bool {
	switch base {
	case "sh", "bash", "zsh", "dash", "ash", "ksh", "csh", "tcsh", "fish", "pwsh", "powershell", "cmd", "cmd.exe":
		return true
	}
	return false
}
