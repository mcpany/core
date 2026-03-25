# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

.PHONY: prepare lint test build run dev clean format update
.DEFAULT_GOAL := help

prepare:
	@echo "Installing Bazelisk..."
	@sudo curl -L https://github.com/bazelbuild/bazelisk/releases/download/v1.19.0/bazelisk-linux-amd64 -o /usr/local/bin/bazelisk
	@sudo chmod +x /usr/local/bin/bazelisk
	@mkdir -p build/env/bin
	@mkdir -p build/.cache

lint:
	python3 server/tools/check_ts_doc.py ui/src
	bazelisk run //:lint
	bazelisk test //ui:typecheck //ui:lint

test:
	bazelisk test //...

build:
	bazelisk build //...
