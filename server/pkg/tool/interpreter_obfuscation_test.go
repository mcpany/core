package tool

import (
	"testing"
)

func TestInterpreterObfuscationBypass(t *testing.T) {
	cases := []struct {
		name string
		val  string
		cmd  string // For future use
	}{
		{"Perl Comment Bypass", "system # \n ('ls')", "perl"},
		{"Perl Line Continuation Bypass", "system \\\n ('ls')", "perl"},
		{"Ruby Comment Bypass", "system # \n ('ls')", "ruby"},
		{"Python Comment Bypass", "os.system # \n ('ls')", "python"},
		{"Node Comment Bypass", "exec // \n ('ls')", "node"},
		{"Node Block Comment Bypass", "exec /* \n */ ('ls')", "node"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := checkInterpreterFunctionCalls(tc.val, tc.cmd); err != nil {
				t.Logf("Blocked as expected: %v", err)
			} else {
				t.Errorf("BYPASS DETECTED: checkInterpreterFunctionCalls passed for %q (cmd: %s)", tc.val, tc.cmd)
			}
		})
	}
}
