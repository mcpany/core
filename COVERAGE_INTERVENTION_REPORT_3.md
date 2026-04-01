# Coverage Intervention Report

* **Target:** `server/pkg/middleware/esb.go`
* **Risk Profile:** The Entangled State Broker (ESB) middleware acts as a high-risk security layer enforcing cryptographic bounds to mission-root intent. It protects against side-channel timing attacks via Temporal Shard Jitter (TSJ). An untracked regression in this logic could expose the adapter to intent drift, ghost-execution exploits, or information leakages via timing discrepancies. It previously had zero test coverage.
* **New Coverage:** New hermetic test coverage now guards the following behaviors:
    * Configuration disablement passthrough.
    * Proper bypassing for non-tool mutation requests (e.g., `mcp.InitializeRequest`).
    * Enforcement of `x-mission-intent` as both a typed context key and a raw string.
    * Enforcement of `x-entanglement-shard` as both a typed context key and a raw string.
    * Proper propagation of execution errors back to the client natively.
* **Verification:** Confirmed that `make test` and `make lint` execute successfully, indicating no regressions in the surrounding codebase or styling violations.
