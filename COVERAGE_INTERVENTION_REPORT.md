# Coverage Intervention Report

* **Target:** `server/pkg/tool/types.go`
* **Risk Profile:** This file was selected because it contains extremely critical logic for executing arbitrary commands (`CommandTool.Execute`, `LocalCommandTool.Execute`), handling generic API requests (`HTTPTool`, `GRPCTool`), and enforcing security guardrails (`checkUnquotedKeywords`, `checkAwkInjection`, `checkNodePerlPhpInjection`). With very high cyclomatic complexity and various unchecked error and streaming execution paths, it represents significant "Dark Matter" risk for security bugs, command injections, and runtime panics.
* **New Coverage:** The following logic paths are now guarded by comprehensive, table-driven tests:
  - Injection validation handlers (`checkUnquotedKeywords`, `checkAwkInjection`, `checkNodePerlPhpInjection`) now strictly verify behaviors with escaped sequences, quotes, backticks, and common adversarial payloads, fully isolating true vulnerabilities from false positives.
  - Streaming execution paths for `CommandTool` (`StreamExecute`) and context parsing (`IsStreaming`).
  - Safe conversion paths via `MCPTool`.
* **Verification:** `make test` successfully tests the new components alongside all existing legacy tests, with zero negative impact ("Do No Harm" principle verified). Linting is clean.

* **Target:** `server/pkg/pool/pool.go`
* **Risk Profile:** This file was selected because connection pooling is critical for performance and resource management, especially when under high load or when connections fail dynamically. Complex states, closures, locks, and channel operations present high risk of deadlocks, panics, and resource leaks. Many edge cases (such as pools closing while threads are waiting to get a client) were untested.
* **New Coverage:** The following logic paths are now guarded by explicit tests:
  - Error handling when spinning up initial clients.
  - Logic to handle getting items from a pool that gets closed concurrently.
  - Returning items to a pool that is closed asynchronously.
  - Nil factory clients checks.
* **Verification:** `bazelisk test //server/...` passes and successfully tests the new coverage cases alongside all existing legacy tests, with zero negative impact ("Do No Harm" principle verified). Linting is clean.
