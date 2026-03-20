import os
import re

def fix_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    # MD047: End with newline
    if not content.endswith('\n'):
        content += '\n'

    # Fix MD030 (Space after list markers)
    content = re.sub(r'^(\s*([-*+]|\d+\.))\s{2,}', r'\1 ', content, flags=re.MULTILINE)

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
