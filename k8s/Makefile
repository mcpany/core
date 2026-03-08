
.PHONY: test build-images

build-images:
	bazelisk run --config=local //server/cmd/server:server_tarball
	bazelisk run --config=local //server/tests/integration/cmd/mocks/http_echo_server:http_echo_server_tarball

test: build-images
	E2E=true PLAYWRIGHT_GREP_INVERT="Screenshots" go test -v -count=1 ./operator/tests/... -timeout 20m

test-full:
	E2E=true go test -v github.com/mcpany/core/server/tests/integration -run TestK8sFullStack -timeout 20m
