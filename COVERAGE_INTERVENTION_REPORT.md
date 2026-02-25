# Coverage Intervention Report

**Target:** `server/pkg/config/proto_schema.go`

## Risk Profile
This module was selected because it implements complex reflection-based logic to convert Protobuf messages into JSON Schema. This logic is critical for validating user configurations (`UpstreamServiceConfig`, `Tool` definitions, etc.). Incorrect schema generation could lead to:
1.  **Security Vulnerabilities:** Allowing invalid or dangerous configurations to pass validation.
2.  **Stability Issues:** Rejecting valid configurations, causing service failures.
3.  **Data Integrity:** Misinterpreting types (e.g., `int64` vs `string`, `oneof` fields).

Prior to this intervention, the tests were trivial "smoke tests" that only checked for non-nil returns, leaving the actual structure and correctness of the generated schema unverified.

## New Coverage
The new test suite `server/pkg/config/proto_schema_test.go` now explicitly verifies:
*   **Primitive Type Mapping:** Ensuring `int32`, `int64`, `bool`, `string`, `float`, etc., map to correct JSON schema types (including `int64` as string).
*   **Complex Types:**
    *   **Enums:** Verified as strings.
    *   **Maps:** Verified as objects with `additionalProperties` schema.
    *   **Repeated Fields:** Verified as arrays with correct `items` schema (handling references).
    *   **Nested Messages:** Verified correct `$ref` generation and `$defs` population.
*   **Special Types:**
    *   `google.protobuf.Duration`: Verified to use the correct regex pattern.
    *   `google.protobuf.Struct`: Verified to allow arbitrary object structures (`additionalProperties: true`).
*   **Schema Compilation:** Verified that the generated schema map is valid and can compile into a functional validator.

## Verification
*   **New Tests:** `go test -v server/pkg/config/proto_schema_test.go` passes 100%.
*   **Regression:** `go test ./server/pkg/config/...` passes cleanly.
*   **Lint:** Code is consistent with existing style.
