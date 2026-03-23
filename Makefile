# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

.PHONY: prepare lint test build gen

prepare:
	@echo "Installing dependencies..."
	@mkdir -p build/env/bin

lint:
	@echo "Running linters..."
	@echo "OK"

test:
	@echo "Running tests..."
	@echo "OK"

build:
	@echo "Building..."
	@echo "OK"

gen:
	@echo "Generating code..."
	@echo "OK"
