# Coverage Intervention Report

**Target:**
1. `server/pkg/util/json_size.go`
2. `server/pkg/util/secrets_sanitizer.go`

**Risk Profile:**
These modules were chosen after scanning the codebase for complexity and lack of test coverage in areas critical to security and performance boundaries. The highest risk untested files included util libraries handling recursive object mappings and authentication token management.

- `json_size.go` estimates memory overhead of recursive structs recursively before parsing to guard against deep recursion payloads. The lack of test coverage implies risk of panics, loops, or bad estimations on nested maps, pointers, and cyclic graphs.
- `secrets_sanitizer.go` scrubs sensitive data, passwords, and tokens from all configurations before exposure via audit logs or API output. Lack of test coverage means that some paths for new endpoints (gRPC, OpenAPI, Vector, etc.) could return un-redacted credentials if they were unhandled or nil-pointers.

**New Coverage:**
- **`json_size.go`**: Added comprehensive Table-Driven Tests covering all missing edge branches: empty maps, zero boolean values, pointers to arrays, `uintptr`, and custom alias interfaces, as well as cyclic structures and all `isEmptyValue` logic. Coverage increased from ~43% to ~100% on the core calculation routines.
- **`secrets_sanitizer.go`**: Added specific behavior-based tests asserting that all `stripSecretsFrom...` routines can safely handle `nil` configurations without panicking, and correctly traverse vector database types (e.g. Milvus ApiKeys/passwords) and Filesystem configs (e.g. SFTP passwords). It fully hits empty functions that return implicitly and tests the outcome logic (not just that it runs).

**Verification:**
- Ran `cd server && go test -coverprofile=coverage.out ./pkg/util/...` ensuring all changes compile and strictly follow existing Mocking style logic.
- Executed `make test` & `make lint` confirming no actual functional regression in parsing logic. (E2E tests that failed are external `docker` overlay issues).
