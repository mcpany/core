# Coverage Intervention Report

## Target
**File:** `server/pkg/util/redact_fast.go`

## Risk Profile
This file was selected for intervention because it represents "Dark Matter" in the codebase:
*   **High Complexity:** It implements a zero-allocation, streaming JSON parser manually (700 LOC). It handles complex logic like skipping whitespace/comments, unescaping strings, and buffer management.
*   **Security Critical:** Its purpose is to redact sensitive information (PII, credentials) from logs. Failure here means leaking user secrets.
*   **Low Coverage:** Initial analysis showed 0 lines of test code specifically dedicated to this file (though some indirect or fragmented coverage existed). It lacked a comprehensive, structured test suite defining its contract.

## New Coverage
A new test suite `server/pkg/util/redact_fast_test.go` was created, implementing robust Table-Driven Tests.
Specific logic paths now guarded include:
*   **Happy Path:** Standard JSON objects with known sensitive keys (`password`, `token`, etc.).
*   **Recursion/Nesting:** Deeply nested objects and arrays to ensure the parser correctly maintains state/depth.
*   **Edge Cases:** Handling of `null`, `true`, `false`, numbers, empty objects `{}`, and empty arrays `[]`.
*   **Robustness:** Verification that malformed JSON (unclosed strings/braces, garbage input) does not crash the server and fails safe (redacting potential secrets even in broken JSON).
*   **Performance/Buffer Management:** Large input tests to verify lazy allocation and buffer resizing logic.
*   **Escaping:** Correct handling of escaped quotes `\"` and backslashes in keys and values.

## Verification
*   **Unit Tests:** `go test -v github.com/mcpany/core/server/pkg/util -run TestRedactJSONFast` passed successfully.
*   **Regression:** The full package tests passed, ensuring no interference with existing `redact` logic.
