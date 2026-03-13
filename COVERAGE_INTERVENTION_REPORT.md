# Coverage Intervention Report

**Target:** `server/pkg/app/api_users.go`
**Function:** `handleUsers` and `handleUserDetail`

**Risk Profile:**
The `api_users.go` component handles the core administration of system users, including listing, creating, retrieving, updating, and deleting profiles. The operations are sensitive since they govern role assignment and credential initialization. The cyclomatic complexity is significant due to varying JSON body formats, authentication enforcement (e.g., RBAC), and database update checks. However, test coverage before intervention was relatively low (around 50.8%), missing assertions for edge cases, JSON unmarshaling fallbacks, authorization constraints, and privilege escalation guards.

**New Coverage:**
The following logic paths are now properly tested:
- Creating a user (`POST /users`): successful path, duplicate conflicts (409), missing IDs (400), malformed request bodies, and correct password hashing logic execution.
- Fetching user details (`GET /users/{id}`): missing IDs, unauthenticated requests, and forbidden cross-account access for non-administrators.
- Updating users (`PUT /users/{id}`): ID mismatch tracking, updates to non-existent profiles, malformed input payload resilience, and strict protections against non-admins escalating their own roles.
The file's overall test coverage for `handleUsers` has increased from 50.8% to 70.8%, and `handleUserDetail` has increased from 50.7% to 67.1%.

**Verification:**
Confirmed that `make test` runs flawlessly and the tests run cleanly under `go test ./pkg/app/...`. The implementation respects the "Do No Harm" principle and hermetically mocks storage and user properties via the `memory.NewStore()` and `httptest.NewRecorder()`.
