# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import re

with open('server/pkg/app/server_test.go', 'r') as f:
    content = f.read()

# Replace:
# JSONRPCPort:     fmt.Sprintf("%d", port),
# GRPCPort:        "0",
# With:
# JSONRPCPort:     fmt.Sprintf("127.0.0.1:%d", port),
# GRPCPort:        "127.0.0.1:0",

content = re.sub(r'JSONRPCPort:\s*fmt\.Sprintf\("%d", port\),\s*GRPCPort:\s*"0"',
                 'JSONRPCPort:     fmt.Sprintf("127.0.0.1:%d", port),\n\t\t\t\tGRPCPort:        "127.0.0.1:0"',
                 content)

# Replace:
# JSONRPCPort:     "0",
# GRPCPort:        fmt.Sprintf("%d", port),
# With:
# JSONRPCPort:     "127.0.0.1:0",
# GRPCPort:        fmt.Sprintf("127.0.0.1:%d", port),

content = re.sub(r'JSONRPCPort:\s*"0",\s*GRPCPort:\s*fmt\.Sprintf\("%d", port\)',
                 'JSONRPCPort:     "127.0.0.1:0",\n\t\t\t\tGRPCPort:        fmt.Sprintf("127.0.0.1:%d", port)',
                 content)

with open('server/pkg/app/server_test.go', 'w') as f:
    f.write(content)
