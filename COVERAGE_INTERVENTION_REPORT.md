# Coverage Intervention Report

*   **Target:** `server/pkg/tool/types.go` (Security Functions: `validateSafePathAndInjection`, `checkForShellInjection`, `fastJSONNumber`)
*   **Risk Profile:**
    *   **High Risk:** This code executes arbitrary commands and interprets user inputs in shell and interpreter contexts (Python, Ruby, Perl, etc.).
    *   **Criticality:** Vulnerabilities here lead to RCE (Remote Code Execution) or SSRF (Server-Side Request Forgery).
    *   **Reason for selection:** The file handles complex "Dark Matter" logic for sanitizing inputs against injection attacks across multiple languages and protocols. Existing tests were sparse or relied on happy-path integration tests.
*   **New Coverage:**
    *   **SSRF Protection:** Added tests for blocked schemes (`file://`, `gopher://`, `ftp://`), dangerous localhost/metadata access (`127.0.0.1`, `169.254.169.254`), and correct handling of URL parsing edge cases.
    *   **Command Injection:** Added comprehensive table-driven tests for:
        *   **Polyglot Injections:** Null bytes, newlines, tabs.
        *   **Interpreter-Specific Vectors:** Python (`__import__`, `exec`, `f-strings`), Ruby (`system`, backticks, interpolation), Perl (`qx`, `system`).
        *   **Nested Injection:** Shell wrapping interpreter calls (e.g., `sh -c "python ..."`).
        *   **Whitespace Bypass:** Leading spaces to evade filters.
    *   **Argument Injection:** Tests for flag injection (`-rf`, `--option`).
    *   **Precision Safety:** Verified `fastJSONNumber` correctly handles large integers without precision loss.
*   **Verification:**
    *   `make gen` run successfully to ensure protobuf artifacts.
    *   `go test -v ./server/pkg/tool/...` passed clean.
    *   Verified that new tests fail if security checks are disabled (demonstrated by initial failures when mock was too permissive).
