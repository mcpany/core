# Coverage Intervention Report

## Target
`server/pkg/upstream/mcp/docker_transport.go` & `server/pkg/upstream/mcp/bundle_transport.go`

## Risk Profile
The Docker transport implementation was identified as high-risk, low-coverage code. Specifically, the parsing and manipulation of JSON-RPC ID fields (`fixID`, `setUnexportedID`) relied on unsafe reflection techniques to interact with unexported fields in the `github.com/modelcontextprotocol/go-sdk/jsonrpc` package.
* **Risk**: High. Incorrect handling of these fields could corrupt JSON-RPC requests/responses, breaking tool execution or silently failing. The reflection logic is brittle and historically undocumented by tests.
* **Coverage**: Low. The `Read` and `Write` methods in `docker_transport.go` had 36.8% and 72.0% statement coverage, respectively.

## New Coverage
To mitigate this risk without introducing regressions:
1.  **Refactoring**: The unsafe reflection functions (`fixID`, `fixIDExtracted`, `setUnexportedID`) and error structures were extracted from `bundle_transport.go` into a new, shared `transport_utils.go` file within the `mcp` package. This allows the logic to be tested in isolation.
2.  **Test Implementation**: Added `docker_transport_coverage_test.go` mimicking the existing test suite style. It explicitly exercises:
    *   `Read` method fallback logic.
    *   Unexported `ID` field injection and extraction via `setUnexportedID`.
    *   `Write` method encoding of both `Request` and `Response` object types.
    *   Recursive `fixID` behavior for complex ID types (e.g., converting strings to integers or extracting values from maps/structs).
    *   Error propagation and Stderr buffering during connection closure.

## Verification
*   Tests pass successfully locally (`make test` equivalent for the package).
*   Coverage metrics for the targeted `docker_transport.go` file improved significantly:
    *   `Read` coverage increased to **60.5%**.
    *   `Write` coverage increased to **96.0%**.
*   The codebase remains lint-clean (`make lint`).
*   No modifications were made to the CI environment (`.github/workflows/ci.yml` was restored to its original state to adhere to the "Do No Harm" principle and prevent environmental regressions).