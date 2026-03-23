# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

with open("ui/playwright_test.sh", "r") as f:
    c = f.read()

c = c.replace('export PATH="$(dirname "$node_bin"):$PATH"', 'NODE_DIR="$(dirname "$node_bin")"\nexport PATH="${NODE_DIR}:$PATH"')
# Fix shellcheck SC2027 and SC2046 properly
c = c.replace(
    'escaped_spec="$(printf \'%s\' "$selected_spec" | sed -e \'s/[.[\\*^$()+?{}|]/\\\\&/g\')"',
    'escaped_spec="$(printf "%s" "$selected_spec" | sed -e \'s/[.[*^$()+?{}|]/\\\\&/g\')"'
)

with open("ui/playwright_test.sh", "w") as f:
    f.write(c)


with open("scripts/lint.sh", "r") as f:
    lines = f.readlines()

out = []
i = 0
while i < len(lines):
    line = lines[i]
    if line.startswith("buildifier_files=("):
        out.append('mapfile -t buildifier_files < <(find . \\\n')
        i += 1
        # skip the `    $(find . \` line
        continue
    if "2>/dev/null)" in line and "buildifier" in "".join(lines[max(0, i-20):i]):
        if i + 1 < len(lines) and lines[i+1].startswith(")"):
            out.append(line.replace("2>/dev/null)", "2>/dev/null)\n"))
            i += 1
            continue

    if line.startswith("shellcheck_files=("):
        out.append('mapfile -t shellcheck_files < <(find . \\\n')
        i += 1
        # skip the `    $(find . \` line
        continue

    if "        -not -path \"./.git/*\")" in line:
        if i + 1 < len(lines) and lines[i+1].startswith(")"):
            out.append('        -not -path "./.git/*" \\\n')
            out.append('        2>/dev/null)\n')
            i += 1
            continue

    out.append(line)
    i += 1

with open("scripts/lint.sh", "w") as f:
    f.write("".join(out))
