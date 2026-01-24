
.PHONY: test

test:
ifneq ($(SKIP_E2E),true)
	E2E=true go test -v -count=1 ./operator/tests/... -timeout 20m
else
	@echo "Skipping K8s tests (SKIP_E2E is true)"
endif

test-full:
	E2E=true go test -v github.com/mcpany/core/server/tests/integration -run TestK8sFullStack -timeout 20m
