# Coverage Intervention Report

## Target: `server/pkg/tool/schema_sanitizer.go`

### Risk Profile
This component was selected because it is a **High Risk** area responsible for sanitizing untrusted JSON schemas from tools. A failure here could lead to security vulnerabilities or service instability.
- **Vulnerability:** The original implementation was vulnerable to **Stack Overflow** due to infinite recursion on circular references.
- **Incompleteness:** It failed to sanitize critical JSON schema keywords like `items` (array form), `additionalProperties`, `oneOf`, `anyOf`, `allOf`, and `$defs`.
- **Low Coverage:** Existing tests only covered basic `properties` recursion and `items` (map form).

### New Coverage
The following logic paths are now guarded by the new test suite `server/pkg/tool/schema_sanitizer_coverage_test.go`:
1.  **Circular References:** Verified that the sanitizer now prunes deep recursion (depth > 500) instead of crashing.
2.  **Items Array:** Verified recursive sanitization of `items` when it is an array of schemas (tuple validation).
3.  **AdditionalProperties:** Verified recursive sanitization of `additionalProperties` schema.
4.  **OneOf/AnyOf/AllOf:** Verified recursive sanitization of these combination keywords.
5.  **Definitions:** Verified recursive sanitization of `$defs` and `definitions`.
6.  **Invalid Types:** Verified graceful handling of non-standard types in the schema map.

### Verification
- **Unit Tests:** `go test ./server/pkg/tool/ -run TestSanitizeJSONSchema_Coverage` passed.
- **Regression Tests:** `go test ./server/pkg/tool/...` passed, ensuring no regressions in the tool package.
- **Lint/Build:** The code is clean and compiles successfully.
