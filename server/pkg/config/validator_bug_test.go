package config

import (
	"testing"
    "github.com/stretchr/testify/assert"
)

func TestValidateStdioArgs_PythonDashC(t *testing.T) {
	// Case: python -c "print(1.5)"
    // "print(1.5)" has extension ".5)" so it is treated as a file.
	err := validateStdioArgs("python", []string{"-c", "print(1.5)"}, "")
	assert.NoError(t, err)
}

func TestValidateStdioArgs_PythonDashC_NoExt(t *testing.T) {
    // Case: python -c "print(1)"
    // "print(1)" has no extension. Should be ignored.
    err := validateStdioArgs("python", []string{"-c", "print(1)"}, "")
    assert.NoError(t, err)
}

func TestValidateStdioArgs_NodeEval(t *testing.T) {
    // node -e "console.log('hello.world')"
    // "console.log('hello.world')" has extension ".world')"
    err := validateStdioArgs("node", []string{"-e", "console.log('hello.world')"}, "")
    assert.NoError(t, err)
}

func TestValidateStdioArgs_BashDashC(t *testing.T) {
    // bash -c "echo hello.world"
    err := validateStdioArgs("bash", []string{"-c", "echo hello.world"}, "")
    assert.NoError(t, err)
}
