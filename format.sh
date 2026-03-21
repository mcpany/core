#!/usr/bin/env bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

find server/pkg -name "*.go" | grep -v "_test.go" | while read -r file; do
    gofmt -w "$file"
done
