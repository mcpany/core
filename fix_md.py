import os
import sys

def fix_file(path):
    with open(path, "rb") as f:
        content = f.read().decode("ascii", errors="ignore")

    lines = content.splitlines()
    fixed_lines = [line.rstrip() for line in lines]

    # Ensure a single trailing newline
    new_content = "\n".join(fixed_lines).strip() + "\n"

    with open(path, "w", encoding="ascii") as f:
        f.write(new_content)

files = [
    "docs/02_strategic_vision.md",
    "docs/03_feature_inventory.md",
    "server/roadmap.md",
    "ui/roadmap.md",
    "docs/research/market-sync-2026-06-14.md",
    "docs/features/design-hardware-locked-coordination-handshake.md",
    "docs/features/design-sci-interceptor.md"
]

for f in files:
    if os.path.exists(f):
        fix_file(f)
        print(f"Fixed {f}")
