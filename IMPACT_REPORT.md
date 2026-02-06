# Coverage Intervention: LLM Client

## Target
`server/pkg/llm/client.go`

## Risk Profile
*   **High Risk**: This component is the primary interface for interacting with Large Language Models (LLMs), a core capability of the "MCP Any" platform.
*   **Dependency**: It is a direct dependency of the `SmartRecoveryMiddleware`, which relies on it to automatically fix tool execution errors. A failure here breaks the self-healing capabilities of the server.
*   **Complexity**: The code handles network requests, JSON marshaling/unmarshaling, API key management, and error propagation.
*   **Pre-existing Coverage**: 0% (Dark Matter). No tests existed for this package.

## New Coverage
We have implemented a robust test suite in `server/pkg/llm/client_test.go` that covers the following scenarios:

1.  **Happy Path**: Verifies that a valid request is correctly marshaled (headers, body, URL) and a valid response is correctly parsed.
2.  **Network Error**: Verifies behavior when the LLM provider is unreachable (e.g., connection refused).
3.  **API Error (500)**: Verifies that non-200 status codes from the provider are correctly propagated as errors.
4.  **Invalid JSON**: Verifies resilience against malformed JSON responses from the provider.
5.  **OpenAI Error Field**: Verifies handling of logical API errors (where status is 200 but the body contains an error object).
6.  **Empty Choices**: Verifies handling of valid but empty responses (no choices returned).
7.  **Context Cancellation**: Verifies that the client respects context cancellation (timeouts/cancellations).

## Verification
Ran `go test -v ./server/pkg/llm/...` and `go test -v ./server/pkg/middleware/...`.

```
=== RUN   TestChatCompletion
=== RUN   TestChatCompletion/HappyPath
=== RUN   TestChatCompletion/NetworkError
=== RUN   TestChatCompletion/APIError_500
=== RUN   TestChatCompletion/InvalidJSON
=== RUN   TestChatCompletion/OpenAIErrorField
=== RUN   TestChatCompletion/EmptyChoices
=== RUN   TestChatCompletion/ContextCancellation
--- PASS: TestChatCompletion (0.02s)
PASS
ok  	github.com/mcpany/core/server/pkg/llm	0.030s
```

All regression tests in `server/pkg/middleware` passed successfully.
