# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

with open("server/tests/integration/e2e_helpers.go", "r") as f:
    lines = f.readlines()
for i, line in enumerate(lines):
    if "return nil" in line and lines[i-1].strip() == "if err != nil {":
        if "var roots []string" not in lines[i-1]:
            lines[i] = "\t\treturn err\n"
    if "var roots []string" in line:
        lines[i] = "\tvar roots []string //nolint:prealloc\n"
with open("server/tests/integration/e2e_helpers.go", "w") as f:
    f.writelines(lines)
