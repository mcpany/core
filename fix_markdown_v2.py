import os
import re

def fix_content(content):
    content = content.replace('—', '-')
    content = content.replace('–', '-')
    content = content.replace('\u201c', '"')
    content = content.replace('\u201d', '"')
    content = content.replace('\u2018', "'")
    content = content.replace('\u2019', "'")

    lines = content.split('\n')
    new_lines = []

    for i, line in enumerate(lines):
        stripped = line.strip()
        # Rule: Blank line before and after list items IF it's a new list or followed by text
        # But MD032 is usually about lists being surrounded by blank lines as a block.

        if line.startswith('#'):
            if new_lines and new_lines[-1] != '':
                new_lines.append('')
            new_lines.append(line)
            if i < len(lines) - 1 and lines[i+1].strip() != '' and not lines[i+1].startswith('#'):
                new_lines.append('')
        elif stripped.startswith('-') or stripped.startswith('*') or re.match(r'^\d+\.', stripped):
            # Start of a list
            if new_lines and new_lines[-1] != '' and not (new_lines[-1].strip().startswith('-') or new_lines[-1].strip().startswith('*') or re.match(r'^\d+\.', new_lines[-1].strip())):
                new_lines.append('')
            new_lines.append(line)
            # End of a list (if next line is not a list item)
            if i < len(lines) - 1 and lines[i+1].strip() != '' and not (lines[i+1].strip().startswith('-') or lines[i+1].strip().startswith('*') or re.match(r'^\d+\.', lines[i+1].strip())):
                new_lines.append('')
        else:
            new_lines.append(line)

    # Collapse multiple blank lines
    final_lines = []
    for line in new_lines:
        line = line.rstrip()
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
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
        fixed = fix_content(content)
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(fixed)
