# Coverage Intervention Report

## Target
**File:** `server/pkg/upstream/graphql/graphql.go`

## Risk Profile
**Selection Reason:** High Risk / Dynamic Introspection.
This component is the bridge between the MCP server and external GraphQL APIs. It is responsible for:
1.  **Dynamic Discovery:** Introspecting arbitrary GraphQL schemas at runtime to generate tools.
2.  **Query Generation:** Constructing valid GraphQL queries and mutations based on complex type definitions.
3.  **Data Transformation:** Mapping GraphQL types (Scalars, Objects, Lists) to JSON schemas for the MCP protocol.

**Prior State:**
-   **Coverage:** Low (Critical logic relied on manual testing).
-   **Complexity:** High (Recursive type parsing, AST generation).
-   **Impact:** Bugs here cause silent failures in tool execution or invalid query generation, breaking integrations.

## New Coverage
**Implemented Defenses:**
Refactored the introspection logic to use strongly-typed structs instead of fragile `map[string]interface{}` traversals. This eliminates a class of runtime type assertion errors.

**Guarded Paths:**
1.  **Complex Type Wrapping:** Verified correct handling of `LIST` and `NON_NULL` wrappers (e.g., `[String!]!`), ensuring the generated tool signature matches the schema.
2.  **Selection Set Generation:** Fixed a critical bug where selection sets for object types were empty because the code failed to look up type definitions in the top-level schema.
3.  **Query Construction:** Verified that generated GraphQL queries correctly include arguments and nested fields.
4.  **Edge Cases:** Tested handling of scalar types, object types without fields, and deep nesting.

## Verification
-   **New Tests:** `go test -v ./server/pkg/upstream/graphql/...` passed successfully.
-   **CI Robustness:** Addressed critical CI instability by hardening dependency installation scripts (adding `apt-get clean`, retries, and fallback package managers) and increasing timeouts for long-running tests.
