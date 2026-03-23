# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

with open("ui/playwright_test.sh", "r") as f:
    c = f.read()

c = c.replace('export PATH="$(dirname "$node_bin"):$PATH"', 'NODE_DIR="$(dirname "$node_bin")"\nexport PATH="${NODE_DIR}:$PATH"')
c = c.replace("escaped_spec=\"$(printf '%s' \\\"$selected_spec\\\" | sed -e 's/[.[\\*^$()+?{}|]/\\\\&/g')\"", "escaped_spec=\"$(printf '%s' \\\"$selected_spec\\\" | sed -e 's/[.[\\\\*^\\\\$()+?{}|]/\\\\\\\\&/g')\"")

with open("ui/playwright_test.sh", "w") as f:
    f.write(c)


with open("scripts/lint.sh", "r") as f:
    c = f.read()

c = c.replace("buildifier_files=(\n    $(find .\n", "mapfile -t buildifier_files < <(find .\n")
c = c.replace("shellcheck_files=(\n    $(find .\n", "mapfile -t shellcheck_files < <(find .\n")
c = c.replace("        2>/dev/null)\n)", "        2>/dev/null)")
c = c.replace("        -not -path \"./.git/*\")\n)", "        -not -path \"./.git/*\"\n        2>/dev/null)")

with open("scripts/lint.sh", "w") as f:
    f.write(c)
