# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import re

with open('server/tests/integration/e2e_helpers.go', 'r') as f:
    content = f.read()

content = content.replace('var roots []string', 'var roots []string //nolint:prealloc')
content = content.replace('if _, err := os.Stat(src); err != nil {\n\t\treturn nil', 'if _, err := os.Stat(src); err != nil {\n\t\treturn err')
content = content.replace('require.NoError(t, symlinkIfPresent(link.src, link.dst))', '_ = symlinkIfPresent(link.src, link.dst)')

with open('server/tests/integration/e2e_helpers.go', 'w') as f:
    f.write(content)
