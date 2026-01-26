package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCmdExeSingleQuoteVulnerability(t *testing.T) {
	// Case 1: cmd.exe with single quotes - Dangerous input
	// cmd.exe treats single quotes as literals, not grouping characters.
	// Therefore, characters like '&' are interpreted as operators even inside single quotes.

	val := "& calc.exe"
	arg := "echo '{{INPUT}}'"
	placeholder := "{{INPUT}}"
	command := "cmd.exe"

	err := checkForShellInjection(val, arg, placeholder, command)

	assert.Error(t, err, "Should detect shell injection for cmd.exe even inside single quotes")
	assert.Contains(t, err.Error(), "shell injection detected")
}

func TestCmdExeCaseInsensitivity(t *testing.T) {
	// Verify that the check is case-insensitive (e.g. CMD.EXE)
	val := "& calc.exe"
	arg := "echo '{{INPUT}}'"
	placeholder := "{{INPUT}}"
	command := "CMD.EXE"

	err := checkForShellInjection(val, arg, placeholder, command)
	assert.Error(t, err, "Should detect shell injection for CMD.EXE (mixed case)")
}

func TestCmdExePathHandling(t *testing.T) {
	// Verify that full paths are handled correctly
	val := "& calc.exe"
	arg := "echo '{{INPUT}}'"
	placeholder := "{{INPUT}}"

	tests := []string{
		`C:\Windows\System32\cmd.exe`,
		`/usr/bin/cmd.exe`, // Wine/emulator scenario
		`..\cmd.exe`,
	}

	for _, cmd := range tests {
		err := checkForShellInjection(val, arg, placeholder, cmd)
		assert.Error(t, err, "Should detect shell injection for path: "+cmd)
	}
}

func TestCmdExeSafeInput(t *testing.T) {
	val := "hello"
	arg := "echo '{{INPUT}}'"
	placeholder := "{{INPUT}}"
	command := "cmd.exe"

	err := checkForShellInjection(val, arg, placeholder, command)
	assert.NoError(t, err, "Should allow safe input")
}

func TestShSingleQuoteSafe(t *testing.T) {
	// bash/sh respects single quotes, so this should be allowed
	// (assuming the shell is POSIX compliant)
	val := "& echo vulnerable"
	arg := "echo '{{INPUT}}'"
	placeholder := "{{INPUT}}"
	command := "bash"

	err := checkForShellInjection(val, arg, placeholder, command)
	assert.NoError(t, err, "Should allow & inside single quotes for bash")
}
