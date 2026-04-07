# Coverage Intervention Impact Report

## Target
`server/pkg/app/api_hitl.go`

## Risk Profile
This component contains the HTTP handlers and global state (`HITLState`) for Human-in-the-Loop (HITL) manual interventions and approvals (such as `database.drop_table` or `aws.terminate_instance`). It has high cyclomatic complexity (cc=9) and manages critical authorization logic without *any* corresponding tests, which signifies a high risk to core business authorization paths in the event of unintended regressions.

## New Coverage
Added `server/pkg/app/api_hitl_test.go` to comprehensively cover:
- State initialization via `newHITLState()` and its effect on `globalHITLState`.
- `GET /hitl/approvals`: Verification that pending approvals are correctly fetched.
- `POST /hitl/approvals/{id}`:
  - Validations that a validly formed action (e.g., `approved` or `rejected`) is correctly processed.
  - Confirms that after processing, the item is removed from the pending state.
  - Validates correct response codes and format on `MethodNotAllowed` (e.g. GET instead of POST) and `BadRequest` (e.g. invalid JSON) error states.

## Verification
- Test code matches testing structure within `server/pkg/app/` using `httptest` and `testify/assert`.
- Ran `bazelisk test //server/pkg/app:app_test` and verified passing tests and successful integration within the suite.
- Code matches production readiness criteria.
