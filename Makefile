# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

.PHONY: lint prepare test

prepare:
	cd ui && npm ci

lint:
	cd ui && npm run lint
	npx @bazel/bazelisk run //:lint
	npx @bazel/bazelisk test //ui:lint

test:
	npx @bazel/bazelisk test //...
