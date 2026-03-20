import os
import re

def fix_markdown(filepath):
    if not os.path.exists(filepath):
        return

    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()

    # 1. No tabs
    content = content.replace('\t', '    ')

    # 2. No trailing whitespace
    lines = [line.rstrip() for line in content.splitlines()]

    # 3. Ensure blank line before and after headers
    # (Simplified: ensure exactly one blank line before headers (except at start) and exactly one after)
    new_lines = []
    for i, line in enumerate(lines):
        if line.startswith('#'):
            if i > 0 and new_lines[-1] != '':
                new_lines.append('')
            new_lines.append(line)
            # We will handle the blank line after in the next step or here
        else:
            new_lines.append(line)

    # 4. Ensure blank line after headers
    lines = new_lines
    new_lines = []
    for i, line in enumerate(lines):
        new_lines.append(line)
        if line.startswith('#') and i < len(lines) - 1 and lines[i+1] != '':
            new_lines.append('')

    # 5. Ensure blank line before and after horizontal rules
    lines = new_lines
    new_lines = []
    for i, line in enumerate(lines):
        if line == '---':
            if i > 0 and new_lines[-1] != '':
                new_lines.append('')
            new_lines.append(line)
            if i < len(lines) - 1 and lines[i+1] != '':
                new_lines.append('')
        else:
            new_lines.append(line)

    # 6. Ensure blank line before and after lists
    # This is trickier. Let's just ensure blank line before a list if the previous line isn't a header or list item.
    lines = new_lines
    new_lines = []
    for i, line in enumerate(lines):
        is_list_item = re.match(r'^\s*([-*+]|\d+\.)\s', line)
        if is_list_item:
            if i > 0 and new_lines[-1] != '' and not re.match(r'^\s*([-*+]|\d+\.)\s', new_lines[-1]) and not new_lines[-1].startswith('#'):
                new_lines.append('')
        new_lines.append(line)

    # 7. Deduplicate blank lines
    lines = new_lines
    new_lines = []
    for line in lines:
        if line == '':
            if not new_lines or new_lines[-1] != '':
                new_lines.append(line)
        else:
            new_lines.append(line)

    # 8. Exactly one newline at EOF
    content = '\n'.join(new_lines).strip() + '\n'

    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)

files_to_fix = [
    'docs/02_strategic_vision.md',
    'docs/03_feature_inventory.md',
    'docs/features/design-aia-broker.md',
    'docs/features/design-cec.md',
    'docs/features/design-esb.md',
    'docs/features/design-t2t-encryption.md',
    'docs/research/market-sync-2026-06-18.md',
    'server/roadmap.md',
    'ui/roadmap.md'
]

for f in files_to_fix:
    print(f"Fixing {f}...")
    fix_markdown(f)
