import os
import re

MD_LICENSE = re.compile(r"<!--\nCopyright 2026 Author\(s\) of MCP Any\nSPDX-License-Identifier: Apache-2.0\n-->")

def check_md(filepath):
    with open(filepath, 'r') as f:
        content = f.read()
    if not MD_LICENSE.match(content):
        print(f"MISSING OR INVALID LICENSE: {filepath}")
        return False
    if content.rstrip() + "\n" != content:
        print(f"INVALID WHITESPACE/NEWLINE: {filepath}")
        return False
    return True

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
        check_md(f)
