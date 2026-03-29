# Security Hardening: Enforcing Granular Scopes and Resource Limits

## Description
This PR addresses critical vulnerabilities surrounding the lack of proper granular scoping in our gRPC middleware and unrestrained resource execution inside Docker containers.

## Fixes Implemented
1. **AuthMiddleware `PromptsGet` Bypass Fix:** Corrected a logic flaw where `PromptsGet` failed to extract a fallback `serviceID` appropriately, allowing authentication bypass if only the global API key was provided.
2. **gRPC Middleware Scope Enforcement:** The previous gRPC interceptors (`grpcUnaryInterceptor` and `grpcStreamInterceptor`) allowed connections via simple IP allow-listing but entirely bypassed the AuthManager's `Authenticate` routine. They now explicitly enforce `AuthManager.Authenticate` checks, ensuring requests respect proper tokens, service definitions, and granular scopes.
3. **Container Resource Constraints:** Implemented hard resource limits in `hostConfig` for Docker execution (`dockerExecutor.Execute` and `dockerExecutor.ExecuteWithStdIO`), limiting containers to 512MB of RAM, 1 CPU core, and 1024 PIDs max to prevent resource exhaustion attacks via unbounded tool execution. Added bounds directly to `io.Copy` in Docker logging outputs to protect the system.

## Verification
- Wrote and passed an integration test suite `TestGRPCMiddleware_EnforceGranularScopes` targeting `server/tests/security/grpc_middleware_security_test.go` confirming 403 Forbidden expectations on unauthenticated/unscoped gRPC calls.
- Resolved unrelated `TestSuggestFix_Fuzzy` regression for testing integrity.
- `bazelisk test //...` verified fully green across all unit and integration environments.
