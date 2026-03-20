import sys
content = open("server/pkg/tool/command_injection_repro_test.go").read()

if "TestMyPythonQuotedExecfileInjection" not in content:
    content += '''
func TestMyPythonQuotedExecfileInjection(t *testing.T) {
	val := "execfile('hack.py')"
	err := checkForShellInjection(val, "\\"{{val}}\\"", "{{val}}", "python", false)
	if err == nil {
		t.Fatalf("Vulnerability confirmed: execfile passed validation.")
	} else {
        t.Logf("Blocked with error: %v", err)
    }
}

func TestJSFunctionKeywordsInjection(t *testing.T) {
	val := "setTimeout('console.log(1)')"
	err := checkForShellInjection(val, "\\"{{val}}\\"", "{{val}}", "node", false)
	if err == nil {
		t.Fatalf("Vulnerability confirmed: setTimeout passed validation.")
	} else {
        t.Logf("Blocked with error: %v", err)
    }
}

func TestPHPKeywordsInjection(t *testing.T) {
    val := "passthru('ls')"
	err := checkForShellInjection(val, "\\"{{val}}\\"", "{{val}}", "php", false)
	if err == nil {
		t.Fatalf("Vulnerability confirmed: passthru passed validation.")
	} else {
        t.Logf("Blocked with error: %v", err)
    }
}
'''
open("server/pkg/tool/command_injection_repro_test.go", "w").write(content)
