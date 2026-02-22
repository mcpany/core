
.PHONY: test

test:
	@if ! docker image inspect mcpany/http-echo-server:latest >/dev/null 2>&1; then \
		echo "Building mcpany/http-echo-server:latest..."; \
		$(MAKE) -C .. build-http-echo-docker; \
	fi
	E2E=true PLAYWRIGHT_GREP_INVERT="Screenshots" go test -v -count=1 ./operator/tests/... -timeout 20m

test-full:
	E2E=true go test -v github.com/mcpany/core/server/tests/integration -run TestK8sFullStack -timeout 20m
