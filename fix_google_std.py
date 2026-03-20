import os
import re

def fix_file(filepath):
    with open(filepath, 'r') as f:
        lines = f.readlines()

    new_lines = []
    for line in lines:
        # MD030: Fix space after list markers (standardize to 1 space)
        line = re.sub(r'^(\s*[-*+])\s{2,}', r'\1 ', line)
        line = re.sub(r'^(\s*\d+\.)\s{2,}', r'\1 ', line)
        new_lines.append(line)

    content = "".join(new_lines)

    # MD047: End with newline
    if not content.endswith('\n'):
        content += '\n'

    # Ensure blank lines around headings (MD022) and lists (MD032)
    # This is complex to do perfectly with regex, so we'll do common cases.

    # Heading spacing
    content = re.sub(r'([^\n])\n(#{1,6} )', r'\1\n\n\2', content)
    content = re.sub(r'(#{1,6} .*)\n([^\n])', r'\1\n\n\2', content)

    # List spacing
    content = re.sub(r'([^\n])\n([-*+] )', r'\1\n\n\2', content)
    content = re.sub(r'([-*+] .*)\n([^\n])', r'\1\n\n\2', content)

    with open(filepath, 'w') as f:
        f.write(content)

files = [
    'docs/research/market-sync-2026-06-02.md',
    'docs/02_strategic_vision.md',
    'docs/03_feature_inventory.md',
    'docs/features/design-avq-hub.md',
    'server/roadmap.md',
    'ui/roadmap.md'
]

for f in files:
    if os.path.exists(f):
        fix_file(f)
