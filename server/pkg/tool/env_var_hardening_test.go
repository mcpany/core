// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"
	"github.com/stretchr/testify/require"
)

func TestDangerousEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		envVar   string
		expected bool
	}{
		// Git
		{"GIT_SSH", "GIT_SSH", true},
		{"GIT_SSH_COMMAND", "GIT_SSH_COMMAND", true},
		{"GIT_ASKPASS", "GIT_ASKPASS", true},
		{"GIT_PAGER", "GIT_PAGER", true},
		{"GIT_EDITOR", "GIT_EDITOR", true},
		{"GIT_EXTERNAL_DIFF", "GIT_EXTERNAL_DIFF", true},
		{"GIT_MAN_VIEWER", "GIT_MAN_VIEWER", true},
		{"GIT_SEQUENCE_EDITOR", "GIT_SEQUENCE_EDITOR", true},
		{"GIT_CONFIG_PARAMETERS", "GIT_CONFIG_PARAMETERS", true},
		{"GIT_CONFIG_COUNT", "GIT_CONFIG_COUNT", true},
		{"GIT_CONFIG_KEY_1", "GIT_CONFIG_KEY_1", true},
		{"GIT_CONFIG_VALUE_1", "GIT_CONFIG_VALUE_1", true},

		// Interpreters
		{"PYTHONPATH", "PYTHONPATH", true},
		{"PYTHONSTARTUP", "PYTHONSTARTUP", true},
		{"PYTHONHOME", "PYTHONHOME", true},
		{"PERL5LIB", "PERL5LIB", true},
		{"PERLIB", "PERLIB", true},
		{"PERL5OPT", "PERL5OPT", true},
		{"RUBYLIB", "RUBYLIB", true},
		{"RUBYOPT", "RUBYOPT", true},
		{"NODE_OPTIONS", "NODE_OPTIONS", true},
		{"NODE_PATH", "NODE_PATH", true},
		{"JAVA_TOOL_OPTIONS", "JAVA_TOOL_OPTIONS", true},
		{"JDK_JAVA_OPTIONS", "JDK_JAVA_OPTIONS", true},
		{"_JAVA_OPTIONS", "_JAVA_OPTIONS", true},
		{"R_PROFILE_USER", "R_PROFILE_USER", true},
		{"R_ENVIRON_USER", "R_ENVIRON_USER", true},

		// Shell
		{"BASH_ENV", "BASH_ENV", true},
		{"ENV", "ENV", true},
		{"PS4", "PS4", true},
		{"SHELLOPTS", "SHELLOPTS", true},
		{"PROMPT_COMMAND", "PROMPT_COMMAND", true},
		{"IFS", "IFS", true},

		// Execution & Config Hijacking
		{"GCONV_PATH", "GCONV_PATH", true},
		{"SHELL", "SHELL", true},
		{"HOME", "HOME", true},
		{"XDG_CONFIG_HOME", "XDG_CONFIG_HOME", true},
		{"XDG_DATA_HOME", "XDG_DATA_HOME", true},
		{"XDG_CACHE_HOME", "XDG_CACHE_HOME", true},
		{"PATH", "PATH", true},

		// Dynamic Linker
		{"LD_PRELOAD", "LD_PRELOAD", true},
		{"LD_LIBRARY_PATH", "LD_LIBRARY_PATH", true},
		{"DYLD_LIBRARY_PATH", "DYLD_LIBRARY_PATH", true},
		{"DYLD_INSERT_LIBRARIES", "DYLD_INSERT_LIBRARIES", true},

		// Safe Env Vars
		{"MY_CUSTOM_VAR", "MY_CUSTOM_VAR", false},
		{"USERNAME", "USERNAME", false},
		{"PWD", "PWD", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := isDangerousEnvVar(tc.envVar)
			require.Equal(t, tc.expected, result, "isDangerousEnvVar(%q) expected %v, got %v", tc.envVar, tc.expected, result)
		})
	}
}

func TestLocalCommandTool_PATHHijacking(t *testing.T) {
	// Simple validation to ensure PATH is blocked from injection.
	// Since isDangerousEnvVar blocklist includes PATH, any input attempting to overwrite PATH
	// is skipped and not injected into the final env array passed to executor.
	require.True(t, isDangerousEnvVar("PATH"), "PATH should be considered a dangerous environment variable")
	require.True(t, isDangerousEnvVar("path"), "Lowercase path should be considered a dangerous environment variable")
	require.True(t, isDangerousEnvVar("HOME"), "HOME should be considered a dangerous environment variable")
}
