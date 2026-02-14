package tool

import (
	"testing"
)

func TestPythonInjectionBypass(t *testing.T) {
	// Negative Tests (Must Block)
	base := "python"

	payloads := []string{
		`import subprocess; subprocess.call(["id"])`,
		`import os; os.system('id')`,
		`import os; os.execl('/bin/sh', 'sh')`,
		`import subprocess; subprocess.call(['id'])`,
		`__import__('os').system('id')`,
		`exec('print(1)')`,
		`eval('1+1')`,
		`os.system('id')`,
		`open('/etc/passwd')`,
		`read()`, // Contextual, read( matches
	}

	for _, p := range payloads {
		err := checkInterpreterFunctionCalls(p, base)
		if err == nil {
			t.Errorf("VULNERABILITY: Payload allowed: %s", p)
		} else {
			t.Logf("Correctly blocked: %s -> %v", p, err)
		}
	}

	// Positive Tests (Must Allow)
	// These contain keywords but not in dangerous context
	safePayloads := []string{
		`I read a book`,
		`System status`,
		`Please write to me`,
		`The host is down`, // host contains os? Regex \b matches word boundary.
		`cost`,
		`subprocess_id`, // \b won't match subprocess if followed by _? No, _ is word char.
		`import_data`, // _ is word char.
	}

	for _, p := range safePayloads {
		err := checkInterpreterFunctionCalls(p, base)
		if err != nil {
			t.Errorf("FALSE POSITIVE: Payload blocked: %s -> %v", p, err)
		} else {
			t.Logf("Correctly allowed: %s", p)
		}
	}
}
