# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

for file in $(find server/pkg -name "*.go" | grep -v "_test.go"); do
    gofmt -w $file
done
