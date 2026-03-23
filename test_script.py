# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import re

content = """
// DynamicResource implements the Resource interface for resources that are
// fetched dynamically by executing a tool.
type DynamicResource struct {
"""

decls = re.finditer(r'(?:^\s*//.*$\n)*^type\s+([A-Z]\w*)', content, re.MULTILINE)
for m in decls:
    print(repr(m.group(0)))
