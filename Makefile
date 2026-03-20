# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

.PHONY: docker-lint docker-test k8s-e2e

docker-lint:
	bazelisk run //:lint

docker-test:
	bazelisk test //...

k8s-e2e:
	$(MAKE) -C k8s test
