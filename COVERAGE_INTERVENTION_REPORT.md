# Coverage Intervention Report

* **Target:** `server/pkg/app/api.go` (`handleSecretDetail` and `handleSecretReveal`)
* **Risk Profile:** This code was selected because it is a critical high-risk component that handles secret creation, retrieval, and revealing. It processes highly sensitive data directly. The `handleSecretDetail` function also features a high cyclomatic complexity combined with very low initial test coverage (30.8%), making it prime "Dark Matter" for intervention. The `handleSecretReveal` function had 0.0% coverage.
* **New Coverage:** The test coverage for `handleSecretDetail` was increased from 30.8% to 73.1%, and for `handleSecretReveal` from 0.0% to 80.0%. Specifically, the following logic paths are now guarded:
  - `GET` requests to retrieve existing and non-existent secrets, verifying correct redaction of the value payload.
  - Edge cases such as missing IDs.
  - Invalid HTTP verbs returning `405 Method Not Allowed`.
  - `DELETE` requests to remove existing secrets.
  - `PUT` requests to create new secrets and update existing ones, verifying that ID logic and defaulting Name to ID behaves correctly.
  - `POST` requests to the `/reveal` sub-path to successfully reveal secrets that exist, or return `404 Not Found` if the secret does not exist, and `400 Bad Request` if the ID is missing.
* **Verification:** `cd server && make test` and `make lint` ran cleanly without regressing existing tests.
