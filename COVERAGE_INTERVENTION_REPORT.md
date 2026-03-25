# Coverage Intervention Report

* **Target:** `server/pkg/tool/types.go`
* **Risk Profile:** This file was selected because it contains extremely critical logic for executing arbitrary commands (`CommandTool.Execute`, `LocalCommandTool.Execute`), handling generic API requests (`HTTPTool`, `GRPCTool`), and enforcing security guardrails (`checkUnquotedKeywords`, `checkAwkInjection`, `checkNodePerlPhpInjection`). With very high cyclomatic complexity and various unchecked error and streaming execution paths, it represents significant "Dark Matter" risk for security bugs, command injections, and runtime panics.
* **New Coverage:** The following logic paths are now guarded by comprehensive, table-driven tests:
  - Injection validation handlers (`checkUnquotedKeywords`, `checkAwkInjection`, `checkNodePerlPhpInjection`) now strictly verify behaviors with escaped sequences, quotes, backticks, and common adversarial payloads, fully isolating true vulnerabilities from false positives.
  - Streaming execution paths for `CommandTool` (`StreamExecute`) and context parsing (`IsStreaming`).
  - Safe conversion paths via `MCPTool`.
* **Verification:** `make test` successfully tests the new components alongside all existing legacy tests, with zero negative impact ("Do No Harm" principle verified). Linting is clean.
