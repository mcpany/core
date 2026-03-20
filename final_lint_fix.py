import os
import re

files = [
    "docs/02_strategic_vision.md",
    "docs/03_feature_inventory.md",
    "server/roadmap.md",
    "ui/roadmap.md",
    "docs/research/market-sync-2026-06-14.md",
    "docs/features/design-hardware-locked-coordination-handshake.md",
    "docs/features/design-sci-interceptor.md"
]

def clean_file(path):
    if not os.path.exists(path):
        return
    with open(path, "rb") as f:
        content = f.read().decode("utf-8", errors="ignore")

    # 1. Replace all non-ASCII with space or nothing
    content = "".join(i if ord(i) < 128 else " " for i in content)

    # 2. Split into lines and strip trailing whitespace
    lines = content.splitlines()
    cleaned_lines = [line.rstrip() for line in lines]

    # 3. Join back and ensure single trailing newline
    new_content = "\n".join(cleaned_lines).strip() + "\n"

    with open(path, "w", encoding="ascii") as f:
        f.write(new_content)

for f in files:
    clean_file(f)
    print(f"Cleaned {f}")
