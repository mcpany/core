import os

LICENSE_HEADER = """<!--
Copyright 2026 Author(s) of MCP Any
SPDX-License-Identifier: Apache-2.0
-->
"""

def sanitize(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    # Ensure license header
    if not content.startswith("<!--"):
        content = LICENSE_HEADER + "\n" + content

    # Strip trailing whitespace and ensure exactly one trailing newline
    content = content.rstrip() + "\n"

    with open(filepath, 'w') as f:
        f.write(content)

files = [
    "docs/research/market-sync-2026-06-03.md",
    "docs/02_strategic_vision.md",
    "docs/03_feature_inventory.md",
    "docs/features/design-project-level-policy-gate.md",
    "docs/features/design-prompt-path-protection.md",
    "docs/features/design-mssq.md",
    "server/roadmap.md",
    "ui/roadmap.md"
]

for f in files:
    if os.path.exists(f):
        sanitize(f)
        print(f"Sanitized {f}")
    else:
        print(f"Skipping {f} (not found)")
