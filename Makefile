# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

.PHONY: docker-lint docker-test k8s-e2e lint test prepare build

prepare:
	@mkdir -p build/env/bin
	@GOBIN=$(CURDIR)/build/env/bin go install github.com/bazelbuild/bazelisk@latest
	@ln -sf bazelisk $(CURDIR)/build/env/bin/bazel
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(CURDIR)/build/env/bin v1.64.5

build:
	@bazelisk build //...

lint:
	@cd server && pre-commit run --all-files
	@bazelisk run //:lint

test:
	@bazelisk test //...

docker-lint: lint

docker-test: test

k8s-e2e:
	@$(MAKE) -C k8s test
