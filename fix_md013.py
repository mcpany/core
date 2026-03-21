import os
import textwrap

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

def fix_file(path):
    if not os.path.exists(path): return
    with open(path, 'r', encoding='ascii') as f:
        lines = f.readlines()

    new_lines = []
    for line in lines:
        if len(line) > 80:
            if line.startswith('#'):
                # Don't wrap headers
                new_lines.append(line)
            elif line.strip().startswith('- ') or line.strip().startswith('* '):
                # Wrap list items
                indent = line[:line.find('- ') if '- ' in line else line.find('* ')]
                bullet = '- ' if '- ' in line else '* '
                content = line.strip()[2:]
                wrapped = textwrap.fill(content, width=80, initial_indent=indent+bullet, subsequent_indent=indent+'  ')
                new_lines.append(wrapped + '\n')
            elif '|' in line:
                # Don't wrap tables
                new_lines.append(line)
            else:
                # Wrap regular text
                new_lines.append(textwrap.fill(line.strip(), width=80) + '\n')
        else:
            new_lines.append(line)

    with open(path, 'w', encoding='ascii') as f:
        f.writelines(new_lines)

for f in files:
    fix_file(f)
