# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

.PHONY: docker-lint docker-test k8s-e2e lint test prepare build

prepare:
	go install github.com/bazelbuild/bazelisk@latest

build:
	bazelisk build //...

lint:
	bazelisk run //:lint

test:
	bazelisk test //...

docker-lint: lint

docker-test: test

k8s-e2e:
	$(MAKE) -C k8s test
