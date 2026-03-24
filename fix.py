# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import re

with open('ui/playwright_test.sh', 'r') as f:
    content = f.read()

content = content.replace('export PATH="$(dirname "$node_bin"):$PATH"', 'export PATH\nPATH="$(dirname "$node_bin"):$PATH"')
content = content.replace(r'''escaped_spec="$(printf '%s' "$selected_spec" | sed -e 's/[.[\*^$()+?{}|]/\\&/g')"''', r'''escaped_spec="$(printf "%s" "$selected_spec" | sed -e "s/[.[\*^\$()+?{}|]/\\\\&/g")"''')

with open('ui/playwright_test.sh', 'w') as f:
    f.write(content)
