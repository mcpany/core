# Coverage Intervention Report

**Target:** `server/pkg/tool/schema_sanitizer.go`

**Risk Profile:**
This file handles JSON schema sanitization, which performs recursive descent on schemas provided by external upstream sources to convert them into normalized `structpb.Struct` protobufs for MCP clients. It previously had low test coverage (44.6% to 66.7%). This code was selected because schema parsing bugs (e.g., stack overflows, unhandled types, or infinite loops) in these utility functions can cause systemic failures or panics across the server when resolving tools, directly impacting the availability of the platform.

**New Coverage:**
The following specific logic paths are now fully guarded, bringing line coverage of `server/pkg/tool/schema_sanitizer.go` to 100%:
- Maximum recursion depth limit handling in both `deepCopyJSON` and `sanitizeJSONSchemaInPlace`.
- Edge cases around nested arrays (`items` mapping and array types).
- Error propagation across complex mapping combinators (`oneOf`, `anyOf`, `allOf`, `definitions`, `$defs`, `additionalProperties`).
- Non-map schema types passed directly to the struct converter.
- Invalid map key conversions causing structural parsing failures.

**Verification:**
- Ran targeted `go test` for the specific package which reported 100.0% coverage for functions `SanitizeJSONSchema`, `sanitizeJSONSchemaInPlace`, and `deepCopyJSON`.
- Ran the full `make test` and `make lint` suite over the entire codebase, confirming no legacy tests were broken, adhering strictly to the "Do No Harm" principle.
