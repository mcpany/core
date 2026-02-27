# Coverage Intervention Report

**Target:** `server/pkg/upstream/graphql` (specifically `graphql_lists_test.go`)

**Risk Profile:**
This area was selected because it handles the construction of GraphQL queries, a critical part of the data fetching layer. The logic involves complex recursion to handle nested types (Lists of Non-Nulls, Non-Nulls of Lists, etc.). Errors here would result in malformed queries sent to upstream services, causing data retrieval failures. The combination of high complexity (recursive type parsing) and low prior coverage made it a high-risk candidate ("Dark Matter").

**New Coverage:**
The new test suite `TestGraphQLUpstream_Lists` guards the following logic paths:
1.  **Nested List Handling:** Verifies that deeply nested combinations of `LIST` and `NON_NULL` types are correctly parsed and represented in the generated GraphQL query variables.
    *   Example: `[[[String!]!]!]!`
2.  **Argument Generation:** Ensures that arguments are correctly formatted and attached to the query fields.
3.  **Variable Definitions:** Confirms that the correct variable definitions (e.g., `$var: [[[String!]!]!]!`) are generated for the operation.
4.  **Response Parsing:** (Implicitly guarded by the successful execution of the request against the mock server).

**Verification:**
*   `go test ./server/pkg/upstream/graphql/...` passed cleanly.
*   `make lint` (via CI) is expected to pass with the new configuration.
