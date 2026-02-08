# Impact Report: OpenAPI Recursive Schema Protection

## Target
* **File**: `server/pkg/upstream/openapi/parser.go`
* **Function**: `convertSchemaToStructPB`, `convertOpenAPISchemaToInputSchemaProperties`, `convertOpenAPISchemaToOutputSchemaProperties`

## Risk Profile
* **Why was this code selected?**
    * The OpenAPI parser is a critical component that processes external, potentially untrusted specifications.
    * The schema parsing logic involves recursion to handle nested objects and `AllOf` inheritance.
    * Recursive schemas (e.g., a `Category` containing a list of `Category` objects) are valid in OpenAPI but were not handled by the parser, leading to infinite recursion and a stack overflow (panic/crash).
    * This vulnerability could be exploited to crash the server by providing a malicious or complex OpenAPI spec.

## New Coverage
* **Specific Logic Paths Guarded**:
    * **Infinite Recursion Prevention**: Introduced a `depth` parameter and a `maxDepth` limit (20) to `convertSchemaToStructPB`.
    * **Truncation**: If the recursion depth exceeds `maxDepth`, the parser now safely truncates the schema (returning an empty object) and logs a warning, instead of crashing.
    * **Input & Output Schemas**: Both input (request body/parameters) and output (response body) schema conversions are now protected.

* **New Test**: `server/pkg/upstream/openapi/recursive_test.go`
    * `TestConvertMcpOperationsToTools_RecursiveSchema`: Defines a recursive `Category` schema and asserts that the parser completes successfully (instead of timing out/crashing) and logs the truncation warning.

## Verification
* **New Test**: `go test -v ./server/pkg/upstream/openapi/` passed, confirming that the recursive schema is handled correctly.
* **Regression Testing**:
    * All existing tests in `server/pkg/upstream/openapi/` passed.
    * Dependent tests in `server/pkg/upstream/factory/` passed.
* **Linting**: Code is consistent with existing style.
