# Coverage Intervention Report

## Target
`server/pkg/upstream/mcp/docker_transport.go`

## Risk Profile
This component was selected because `DockerTransport` handles critical I/O communication with Docker containers. The `Read` method involves complex JSON-RPC message parsing, error handling, and state management (e.g., buffering stderr for error reporting). The initial coverage for `Read` was only ~36.8%, leaving significant logic paths untested and prone to regression during refactoring or updates.

## New Coverage
A new test file `server/pkg/upstream/mcp/docker_transport_coverage_test.go` was created to target the low-coverage areas.

**Key Logic Paths Guarded:**
1.  **EOF Handling with Stderr:** Verified that when the container stream closes unexpectedly, any buffered stderr output is correctly returned in the error message.
2.  **ID Type Handling:** Verified support for string, number, and null IDs in both Requests and Responses, ensuring compatibility with various JSON-RPC clients/servers.
3.  **Malformed JSON:** Verified robust error handling for truncated or syntactically invalid JSON messages.
4.  **Header Unmarshalling:** Verified error handling when the message structure doesn't match the expected JSON-RPC header format.

**Metrics:**
*   **Old Coverage (Read):** ~36.8%
*   **New Coverage (Read):** 84.2% (based on local coverage profile analysis)

## Verification
*   **Unit Tests:** All tests in `server/pkg/upstream/mcp` passed, including the new coverage tests.
    *   Command: `go test -v ./server/pkg/upstream/mcp/...`
    *   Result: PASS
*   **Linting:** `make lint` was run. Existing lint errors in `pkg/tool/types.go` (unrelated to changes) persist, but the new code is clean and adheres to project standards.
