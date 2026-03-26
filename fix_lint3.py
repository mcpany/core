# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

with open("server/tests/integration/e2e_helpers.go", "r") as f:
    lines = f.readlines()
for i, line in enumerate(lines):
    if "return nil" in line and "if err != nil {" in lines[i-1]:
        # we need to replace return nil with return nil, err when appropriate or return err. Let's look at the signatures:
        if "[]byte, error" in lines[i-5] or "[]byte, error" in lines[i-6] or "[]byte, error" in lines[i-7] or "[]byte, error" in lines[i-8]:
            lines[i] = "\t\treturn nil, err\n"
        elif "*mcp.ListToolsResult, error" in lines[i-5] or "*mcp.ListToolsResult, error" in lines[i-6] or "*mcp.ListToolsResult, error" in lines[i-7] or "*mcp.ListToolsResult, error" in lines[i-8]:
            lines[i] = "\t\treturn nil, err\n"
        elif "*mcp.CallToolResult, error" in lines[i-5] or "*mcp.CallToolResult, error" in lines[i-6] or "*mcp.CallToolResult, error" in lines[i-7] or "*mcp.CallToolResult, error" in lines[i-8]:
            lines[i] = "\t\treturn nil, err\n"
        else:
             lines[i] = "\t\treturn err\n"

    if "var roots []string" in line and "nolint" not in line:
        lines[i] = "\tvar roots []string //nolint:prealloc\n"
with open("server/tests/integration/e2e_helpers.go", "w") as f:
    f.writelines(lines)
