import os
import textwrap
import re

files = [
    'docs/02_strategic_vision.md',
    'docs/03_feature_inventory.md',
    'docs/features/design-aia-broker.md',
    'docs/features/design-cec.md',
    'docs/features/design-esb.md',
    'docs/features/design-t2t-encryption.md',
    'docs/research/market-sync-2026-06-18.md',
    'server/roadmap.md',
    'ui/roadmap.md',
    'REPORT.md'
]

def fix_line_length(text, width=80):
    lines = []
    for line in text.split('\n'):
        if len(line) <= width or line.startswith('#') or '|' in line or line.startswith('```') or line.startswith('    '):
            lines.append(line)
        elif line.strip().startswith('- ') or line.strip().startswith('* '):
            indent = line[:line.find('- ') if '- ' in line else line.find('* ')]
            bullet = '- ' if '- ' in line else '* '
            content = line.strip()[2:]
            wrapped = textwrap.fill(content, width=width, initial_indent=indent+bullet, subsequent_indent=indent+'  ')
            lines.append(wrapped)
        else:
            lines.append(textwrap.fill(line, width=width))
    return '\n'.join(lines)

def clean(path):
    if not os.path.exists(path): return
    with open(path, 'r', encoding='ascii', errors='ignore') as f:
        content = f.read()

    # 1. Strip non-ASCII
    content = "".join(i for i in content if ord(i) < 128)

    # 2. Fix spacing
    content = re.sub(r'\n{3,}', '\n\n', content)
    content = re.sub(r'^(#+)([^# ])', r'\1 \2', content, flags=re.MULTILINE)

    # 3. Fix line lengths
    content = fix_line_length(content)

    with open(path, 'w', encoding='ascii') as f:
        f.write(content.strip() + '\n')

for f in files:
    clean(f)
