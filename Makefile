# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

.PHONY: lint prepare test

prepare:

lint:
	cd ui && npm run lint

test:
	npx @bazel/bazelisk test //...
