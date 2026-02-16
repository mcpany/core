package tool

import (
	"testing"
)

func TestPHPInjectionSecurity(t *testing.T) {
	vectors := []struct {
		input    string
		name     string
		language string
	}{
		{"passthru('ls')", "passthru", "php"},
		{"shell_exec('ls')", "shell_exec", "php"},
		{"proc_open('ls', [], $pipes)", "proc_open", "php"},
		{"pcntl_exec('/bin/ls')", "pcntl_exec", "php"},
		{"assert('system(\"ls\")')", "assert", "php"},
		{"include('file.php')", "include", "php"},
		{"include_once('file.php')", "include_once", "php"},
		{"require_once('file.php')", "require_once", "php"},
		{"dl('extension.so')", "dl", "php"},
		// Ruby/Perl vectors
		{"syscall(123)", "syscall", "ruby"},
		{"load('file.rb')", "load", "ruby"},
		{"syscall(123)", "syscall", "perl"},
	}

	for _, v := range vectors {
		t.Run(v.name+"_"+v.language, func(t *testing.T) {
			err := checkInterpreterFunctionCalls(v.input, v.language)
			if err == nil {
				t.Errorf("Expected blocked for %s in %s, but was allowed", v.input, v.language)
			} else {
				t.Logf("Correctly blocked: %s (%v)", v.input, err)
			}
		})
	}

	// Verify false positives (e.g. load in Python should be allowed)
	t.Run("python_load", func(t *testing.T) {
		input := "json.load(f)"
		err := checkInterpreterFunctionCalls(input, "python")
		if err != nil {
			t.Errorf("Expected allowed for %s in python, but was blocked: %v", input, err)
		}
	})
}
