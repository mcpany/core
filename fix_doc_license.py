import os

FILES = [
    "docs/research/market-sync-2026-06-03.md",
    "docs/02_strategic_vision.md",
    "docs/03_feature_inventory.md",
    "docs/features/design-project-level-policy-gate.md",
    "docs/features/design-prompt-path-protection.md",
    "docs/features/design-mssq.md",
    "server/roadmap.md",
    "ui/roadmap.md"
]

LICENSE_HEADER = """<!--
Copyright 2026 Author(s) of MCP Any
SPDX-License-Identifier: Apache-2.0
-->

"""

def fix(filepath):
    if not os.path.exists(filepath):
        return
    with open(filepath, 'r') as f:
        lines = f.readlines()

    # Remove existing license header if any
    start = -1
    end = -1
    for i, line in enumerate(lines):
        if line.strip() == "<!--":
            start = i
        if line.strip() == "-->" and start != -1:
            end = i
            break

    if start != -1 and end != -1:
        content = "".join(lines[end+1:]).lstrip()
    else:
        content = "".join(lines).lstrip()

    final_content = LICENSE_HEADER + content
    final_content = final_content.rstrip() + "\n"

    with open(filepath, 'w') as f:
        f.write(final_content)
    print(f"Fixed {filepath}")

for f in FILES:
    fix(f)
