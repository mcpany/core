# Coverage Intervention Report

* **Target:** `server/pkg/tool/base.go` and `server/pkg/tool/callable.go`
* **Risk Profile:** These files handle the fundamental, base-level conversion from configured external capabilities ("Tools") into functional execution logic via MCP interfaces. Low coverage here equates to a high risk of unexpected errors failing quietly at the lowest abstraction layers during standard API mapping execution. Complexity arises from their conversion utilities, type checks, schema rendering, execution modes, and caching.
* **New Coverage:**
  - Validated tool parsing capabilities via `newBaseTool` covering edge cases.
  - Successfully verified execution schemas and definitions translation logic (`MCPTool`).
  - Added full test harness for underlying `CallableTool` types wrapping execution hooks.
  - Checked interface implementations covering core streaming vs. non-streaming states (`IsStreaming`, `StreamExecute`).
* **Verification:** `make lint` passes successfully. Tested new changes explicitly via Bazelisk target `//server/pkg/tool/...` which passes cleanly.
