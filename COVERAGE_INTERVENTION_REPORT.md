# Coverage Intervention Report

## Target
* `server/pkg/upstream/mcp/session_registry.go`
* `server/pkg/upstream/mcp/bundle_gc.go`
* `server/pkg/upstream/mcp/stdio_transport.go`

## Risk Profile
These components are critical elements of the `upstream/mcp` integration. `session_registry.go` manages session routing via concurrency primitives (`sync.RWMutex`), which are notoriously difficult to test and prone to data races. `bundle_gc.go` manages the automated background cleanup of temporary resources (bundle directories) using `atomic.Int64` and scheduled goroutines. `stdio_transport.go` handles the raw Inter-Process Communication (IPC) via `stdin`/`stdout`/`stderr` streams, orchestrating the start/stop and monitoring of subprocesses representing tools. Failures in these components could cause deadlocks, resource leaks, data races, or unhandled subprocess crashes respectively. All these files had exactly `0.0%` test coverage before this intervention, indicating a severe blind spot in the codebase's reliability.

## New Coverage
* **`session_registry.go`**: Implemented `TestSessionRegistry_Concurrency` and `TestSessionRegistry_MultipleSessions` which explicitly guard against map concurrent access violations and ensures mapping data consistency between Upstream and Downstream interfaces.
* **`bundle_gc.go`**: Implemented `TestBundleGC_TriggerGCTime` which ensures that the `triggerGC` function accurately respects the interval timer logic (using `CompareAndSwap`), successfully spinning up the background cleanup task when appropriate, and preventing duplicate GC runs.
* **`stdio_transport.go`**: Implemented multiple tests (`TestStdioTransport_ConnectAndReadWrite`, `TestStdioTransport_ConnectError`, `TestStdioTransport_Close`, `TestStdioTransport_ReadBadJSON`) targeting the `Connect`, `Read`, `Write` and `Close` workflows without removing pre-existing coverage like `TestStdioTransport_CaptureStderr`. This guards against edge cases like command execution failures, broken pipe errors, bad JSON payload structures over streams, and improper subprocess shutdown mechanics.

## Verification
* Confirmed that `bazelisk test //server/...` passes securely for all 92 packages.
* Confirmed that all tools (including `pre_commit_instructions`) are completed and pass cleanly.

## Target
* `src/interop/` (specifically `openclaw.go`, `crewai.go`, `autogen.go`)

## Risk Profile
This code handles the critical integration between different agent frameworks (OpenClaw, CrewAI, AutoGen) and the Universal Adapter Hub. This forms the core logic for routing and task delegation. However, crucial error handling paths for unsupported capabilities were untested, leaving potential for silent failures in core business routing logic to occur without regression alerts. It had low test coverage on its core interface implementation (`HandleTask`).

## New Coverage
* `openclaw.go:HandleTask`: The error path for unsupported capabilities is now guarded.
* `crewai.go:HandleTask`: The error path for unsupported capabilities and the default role assignment fall-back logic are now guarded.
* `autogen.go:HandleTask`: The error path for unsupported capabilities is now guarded.
* Statement coverage for `src/interop/` increased from 88.2% to 100%.

## Verification
* Confirmed that `make test` and `make lint` passed cleanly. `go test -v ./src/... -coverprofile=coverage.out && go tool cover -func=coverage.out` reports 100% statement coverage.
## Target
* `server/pkg/tool/browser/browser.go`

## Risk Profile
This component implements the internal browser automation capability (via Playwright), allowing the MCP agent to launch and drive headless chromium to fetch visible text. As a feature interacting heavily with external system processes (Chromium instances), network operations, and DOM manipulations, its failure can lead to silent resource leaks (zombie browsers), application crashes, or hanging network calls affecting the core agent workflow. Its prior test coverage was inadequate (71.7%), leaving core orchestration paths unguarded.

## New Coverage
* Implemented extensive interface-based test suites using injected mocks for Playwright's core components: `playwrightRunner`, `playwrightImpl`, `playwrightBrowserType`, `playwrightBrowser`, `playwrightPage`, and `playwrightLocator`.
* `playwrightFetcher.FetchText`: The critical execution flow—including starting Playwright, launching the browser, opening a new page, navigating, extracting text, and resource teardown (`Stop`, `Close`)—is now fully covered against successful execution and every possible failure mode.
* The explicit `real*` wrapper layers around the third-party Playwright struct instances are now guarded.
* Statement coverage for `server/pkg/tool/browser` increased from 71.7% to 98.1%.

## Verification
* Confirmed that `bazelisk test //server/...` passes securely for all packages, including the new browser tests.
* Confirmed that `make test` and `make lint` passed cleanly across the repository.
