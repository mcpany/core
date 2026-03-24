# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

prepare:
	@echo "Preparation complete"

lint:
	echo "Running lint on all commits"
	bazel run //:lint

test:
	bazel test //...

test-fast:
	bazel test //... --test_tag_filters=-e2e,-integration,-slow

test-public-api:
	bazel test //tests/...

build-docker:
	bazel build //server:docker
	@if [ "$(PUSH)" = "true" ]; then \
		bazel run //server:docker_push; \
	fi

build-http-echo-docker:
	bazel build //server/tests/integration/cmd/mocks/http_echo_server:docker
	@if [ "$(PUSH)" = "true" ]; then \
		bazel run //server/tests/integration/cmd/mocks/http_echo_server:docker_push; \
	fi

build-e2e-timeserver-docker:
	bazel build //server/tests/integration/cmd/mocks/python_e2e_time_server:docker
	@if [ "$(PUSH)" = "true" ]; then \
		bazel run //server/tests/integration/cmd/mocks/python_e2e_time_server:docker_push; \
	fi

build-cowsay-docker:
	bazel build //server/tests/integration/cmd/mocks/python_cowsay_server:docker
	@if [ "$(PUSH)" = "true" ]; then \
		bazel run //server/tests/integration/cmd/mocks/python_cowsay_server:docker_push; \
	fi

build-everything-docker: build-docker build-http-echo-docker build-e2e-timeserver-docker build-cowsay-docker

e2e-parallel:
	bazel test //ui:playwright --test_tag_filters=e2e --jobs=4

e2e-sequential:
	bazel test //ui:playwright --test_tag_filters=e2e --jobs=1
