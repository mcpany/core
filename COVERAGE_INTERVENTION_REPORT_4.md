# Coverage Intervention Report

* **Target:** `server/pkg/tool/schema_sanitizer.go`
* **Risk Profile:** This module modifies generic JSON schemas in place and deep-copies data passed across boundaries for execution context. Improperly constrained recursion when parsing arbitrarily deep schema bounds exposes the backend to Denial of Service (DoS) via stack overflow or uncontrolled resource exhaustion. Prior coverage failed to assert edge conditions associated with cycle limits, implicit type injection under definitions, nested structures in combinators (oneOf, etc.), and items parameters. Due to its strategic placement directly processing inbound untrusted schemas, it's considered High Risk.
* **New Coverage:** New hermetic test coverage now explicitly guards the following behaviors:
    * Cycle detection and recursion depth limitations (`TestSanitizeJSONSchema_RecursionLimit`, `TestSanitizeJSONSchema_DeepCopyRecursionLimit`).
    * Deep-nested property transformations (`TestSanitizeSchema_Recursive`).
    * Array items and definition traversing validation (`TestSanitizeJSONSchema_TableDriven`).
    * Correct combinator (`oneOf`, `anyOf`, `allOf`) extraction and missing type bindings.
    * Safety behavior around raw scalars, booleans passed as schemas, and nil structures.
* **Verification:** Confirmed that `./bazelisk test //server/pkg/tool/...` executes successfully, meaning the tests pass properly without regressions to existing ones.
