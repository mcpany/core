# Coverage Intervention Report

* **Target:** `server/pkg/upstream/mcp/streamable_http.go`
* **Risk Profile:** This file handles core logic for connecting to an upstream MCP service via HTTP. It uses custom http roundtripper wrappers to provide authentication securely to backend models or downstream clients. If the authentication, parsing, and stream multiplexing configurations have logic errors or regressions, requests might silently fail, stream parsing could panic, or unauthenticated downstream requests might be incorrectly passed through. This directly impacts both system resilience and security isolation.
* **New Coverage:**
  * Unauthenticated scenarios (`authenticatedRoundTripper` behaviors on invalid/nil states, `StreamableHTTP.RoundTrip` initialization with no client).
  * The integration flow inside `createAndRegisterMCPItemsFromStreamableHTTP` with invalid service addresses (`httpAddress` validation), missing addresses, and authentication creation failures (`auth.NewUpstreamAuthenticator` failures on invalid types).
  * Testing `mergeStructs` recursive deep copy capability.
* **Verification:** `bazelisk test //server/pkg/...` passed fully successfully, confirming that standard integration behavior remains intact while new scenarios properly exercise the fallback boundaries.
