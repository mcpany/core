# Coverage Intervention Report

* **Target:** `server/pkg/upstream/mcp/streamable_http.go`
* **Risk Profile:** The function `buildSafeEnv` in this file is responsible for sanitizing environment variables before passing them to external tools/plugins running locally. If this code fails to correctly drop untrusted or secret variables (like API keys) while preserving necessary system variables (`PATH`, `HOME`), it poses a high risk of secret leakage, privilege escalation, and lateral movement. Prior to this intervention, this specific function and its edge cases lacked adequate coverage.
* **New Coverage:** Added comprehensive Google-style Table-Driven tests (`TestBuildSafeEnv`) in `server/pkg/upstream/mcp/streamable_http_test.go`. The tests now explicitly assert the behavior of:
  - Base case (empty inputs).
  - Explicit inclusion of only allowed system environment variables.
  - Correct filtering and redaction of non-allowed/secret environment variables.
  - Safe merging and precedence rules when user-resolved variables override or append to system variables.
* **Verification:** `make test`, `make lint` and the complete test suite via `bazelisk test //server/pkg/upstream/mcp:mcp_test` passed cleanly without any regressions.