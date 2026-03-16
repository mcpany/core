# Coverage Intervention Report

* **Target:** `server/pkg/app/api.go` (specifically `handleSecretDetail` and `handleSecretReveal`)
* **Risk Profile:** This module contains critical authorization and data access logic dealing directly with the system's Secret Management routes (creation, revealing, and deleting of credentials). Given the cyclomatic complexity of routing via the same handler function based on methods, the risk is extremely high. The original coverage for `handleSecretReveal` was 0%, meaning critical credential exposure was untested.
* **New Coverage:**
  * `handleSecretReveal`: Tests now cover the happy path (returning a secret successfully), method not allowed (validating HTTP method), and secret not found. Coverage is now 80.0%.
  * `handleSecretDetail`: Added a new PUT test ensuring the actual values in the store reflect API inputs and handles malformed JSON successfully. Increased `handleSecretDetail` coverage from 30.8% to 61.5%.
* **Verification:** `go test ./pkg/app/...` passes locally without regressions, with `server/pkg/app` coverage jumping from 68.9% to 69.6%.
