# Coverage Intervention Report

* **Target:** `server/pkg/middleware/esb.go`
* **Risk Profile:** This module contains the Entangled State Broker logic which guarantees requests are cryptographically bound to a mission-root intent. It injects Temporal Shard Jitter (TSJ) to neutralize side-channel timing attacks. Testing this logic is high-risk since bypassing it or having it fail open could allow security vulnerabilities around state mutation in tools.
* **New Coverage:**
  * Validated that a disabled ESB configuration bypasses checking appropriately.
  * Verified that non-tool-call requests bypass the checks properly.
  * Explicitly enforced error generation for missing `x-mission-intent` metadata.
  * Explicitly enforced error generation for missing `x-entanglement-shard` metadata.
  * Allowed and tested fallback mechanisms that read string values directly from context if typed context values are unavailable.
  * Tested the Happy Path ensuring TSJ injection triggers and delays the execution by at least the minimal threshold (5ms).
* **Verification:** `make test` equivalents (via `bazelisk test //...`) ran cleanly.
