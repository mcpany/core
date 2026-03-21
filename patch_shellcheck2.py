# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import re

with open("ui/playwright_test.sh", "r") as f:
    content = f.read()

content = content.replace("  escaped_spec=\"$(printf '%s' \"$selected_spec\" | sed -e 's/[.[\\*^$()+?{}|]/\\&/g')\"", '  escaped_spec="$(printf "%s" "$selected_spec" | sed -e "s/[.[\\\\*^$()+?{}|]/\\\\&/g")"')

with open("ui/playwright_test.sh", "w") as f:
    f.write(content)
