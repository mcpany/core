import os
import re

def fix_content(content):
    # Rule 1: Replace non-ASCII em-dashes and other weird chars
    content = content.replace('—', '-')
    content = content.replace('–', '-')
    content = content.replace('\u201c', '"')
    content = content.replace('\u201d', '"')
    content = content.replace('\u2018', "'")
    content = content.replace('\u2019', "'")

    # Rule 2: Ensure headings are surrounded by blank lines
    lines = content.split('\n')
    new_lines = []
    for i, line in enumerate(lines):
        if line.startswith('#'):
            if i > 0 and new_lines and new_lines[-1].strip() != '':
                new_lines.append('')
            new_lines.append(line)
            if i < len(lines) - 1 and lines[i+1].strip() != '':
                new_lines.append('')
        else:
            new_lines.append(line)

    # Rule 3: Strip trailing whitespace
    new_lines = [line.rstrip() for line in new_lines]

    # Rule 4: Ensure exactly one blank line between blocks (collapse multiples)
    final_lines = []
    for line in new_lines:
        if line == '' and final_lines and final_lines[-1] == '':
            continue
        final_lines.append(line)

    return '\n'.join(final_lines)

files = [
    'docs/02_strategic_vision.md',
    'docs/03_feature_inventory.md',
    'docs/research/market-sync-2026-06-14.md',
    'docs/features/design-hardware-locked-coordination-handshake.md',
    'docs/features/design-sci-interceptor.md',
    'server/roadmap.md',
    'ui/roadmap.md'
]

for filepath in files:
    if os.path.exists(filepath):
        print(f"Fixing {filepath}...")
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
        fixed = fix_content(content)
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(fixed)
