# Skipped Tests Manifest

## Backend (Go)

### `server/pkg/app/server_test.go`
*   `TestGRPC_RegError`: Skipped ("startGrpcServer no longer handles registration callbacks")
*   `TestGRPC_Panic`: Skipped ("startGrpcServer no longer handles registration callbacks")
*   `TestStartGrpcServer_PanicHandling`: Skipped ("registration is external")
*   `TestStartGrpcServer_PanicInRegistrationRecovers`: Skipped ("registration is external")

### `server/pkg/upstream/mcp/docker_transport_test.go`
*   `TestDockerTransport_Connect_Integration`: Skipped ("Docker socket not accessible") - *FIXED in previous step*

### `server/pkg/bus/redis/bus_test.go`
*   Various: Skipped ("Redis is not available") - *Environment limitation*

### `server/tests/integration/deployment_test.go`
*   `TestDockerCompose`: Skipped ("heavy integration test... flaky in CI")
*   `TestHelmChart`: Skipped ("helm command not found")

## Frontend (UI)

*   No explicit `it.skip` or `test.skip` found in `ui/` source files (excluding node_modules).
*   However, `ui/src/tests/components/analytics-dashboard.test.tsx` and others use `mockResolvedValue` which we are replacing with real data integration.
*   The "skipped tests" mentioned in the prompt might refer to tests that were commented out or not run, or the `services.spec.ts` which I have now enabled and refactored.
