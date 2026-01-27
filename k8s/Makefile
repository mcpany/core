
.PHONY: test

test:
	E2E=true go test -v -count=1 ./operator/tests/... -timeout 20m

test-full:
	E2E=true go test -v github.com/mcpany/core/server/tests/integration -run TestK8sFullStack -timeout 20m
