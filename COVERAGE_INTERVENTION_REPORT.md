# Coverage Intervention Report

**Target:** `server/pkg/tool/types.go` (specifically security logic for interpreted languages)

**Risk Profile:**
This code was selected because it handles the execution of arbitrary commands via the `CommandTool`. Specifically, it implements critical security checks to prevent Remote Code Execution (RCE) vulnerabilities when interacting with interpreters like PHP, Expect, and Lua. While Python, Ruby, and Perl had existing coverage, these other interpreters were "Dark Matter"—complex security logic protecting against RCE but lacking specific test cases. A failure in this logic would allow an attacker to bypass restrictions and execute arbitrary code on the server.

**New Coverage:**
The following logic paths in `checkInterpreterFunctionCalls`, `checkNodePerlPhpInjection`, and `checkUnquotedKeywords` are now guarded:

*   **PHP Injection Protection:**
    *   Blocking `system('...')`, `exec('...')` calls.
    *   Blocking backtick execution `` `...` ``.
    *   Blocking variable interpolation `${...}`.
    *   Blocking unquoted keywords (strict mode for PHP).
*   **Expect Injection Protection:**
    *   Blocking `spawn` command (unquoted keyword).
    *   Blocking `system` command.
*   **Lua Injection Protection:**
    *   Blocking `os.execute` (via `os` object access).
    *   Blocking `io.popen` (via `popen` keyword).
    *   Blocking `require` (unquoted keyword).

**Verification:**
*   **New Tests:** Created `server/pkg/tool/interpreter_extra_security_test.go` with 10 robust test cases.
*   **Results:** `go test -v ./server/pkg/tool -run TestInterpreterExtraSecurity` passed successfully.
*   **Regressions:** Verified that existing tests in the package pass (`go test -v ./server/pkg/tool/...`).
