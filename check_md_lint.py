import sys
import re

def check_md(filename):
    with open(filename, 'rb') as f:
        content = f.read()

    issues = []
    # Check for non-ASCII
    for i, b in enumerate(content):
        if b > 127:
            issues.append(f"Non-ASCII at byte {i}: {b}")
            break # Just one per file

    text = content.decode('utf-8', errors='replace')
    lines = text.splitlines()
    for i, line in enumerate(lines):
        # Line length
        if len(line) > 80:
            # Allow long links or code blocks?
            # But strict rule says 80.
            issues.append(f"Line {i+1} too long: {len(line)}")

        # Trailing whitespace
        if line.endswith(' ') or line.endswith('\t'):
            issues.append(f"Trailing whitespace at line {i+1}")

    # Trailing newline
    if not text.endswith('\n'):
        issues.append("Missing trailing newline")
    elif text.endswith('\n\n'):
        issues.append("Multiple trailing newlines")

    return issues

files = [
    "docs/02_strategic_vision.md",
    "docs/03_feature_inventory.md",
    "docs/features/design-aia-broker.md",
    "docs/features/design-cec.md",
    "docs/features/design-esb.md",
    "docs/features/design-t2t-encryption.md",
    "docs/research/market-sync-2026-06-18.md",
    "server/roadmap.md",
    "ui/roadmap.md"
]

all_ok = True
for f in files:
    errs = check_md(f)
    if errs:
        all_ok = False
        print(f"--- {f} ---")
        for e in errs:
            print(e)
if all_ok:
    print("All Markdown files OK")
