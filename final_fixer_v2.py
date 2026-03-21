import os
import re

def super_fix(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        lines = f.readlines()

    license_header = [
        "# Copyright 2026 Author(s) of MCP Any\n",
        "# SPDX-License-Identifier: Apache-2.0\n"
    ]

    start_idx = 0
    if len(lines) >= 2 and lines[0].startswith("# Copyright") and lines[1].startswith("# SPDX-License"):
        start_idx = 2

    content_lines = lines[start_idx:]
    processed = []

    for i, line in enumerate(content_lines):
        line = line.rstrip()
        line = line.replace('—', '-').replace('–', '-').replace('\u201c', '"').replace('\u201d', '"').replace('\u2018', "'").replace('\u2019', "'")

        is_heading = line.strip().startswith('#')
        is_list = line.strip().startswith('-') or line.strip().startswith('*') or re.match(r'^\d+\.', line.strip())

        if is_heading:
            if processed and processed[-1].strip() != '':
                processed.append('')
            processed.append(line)
            if i < len(content_lines) - 1 and content_lines[i+1].strip() != '':
                processed.append('')
        elif is_list:
            if processed and processed[-1].strip() != '' and not (processed[-1].strip().startswith('-') or processed[-1].strip().startswith('*') or re.match(r'^\d+\.', processed[-1].strip())):
                processed.append('')
            processed.append(line)
            if i < len(content_lines) - 1:
                next_line = content_lines[i+1].strip()
                next_is_list = next_line.startswith('-') or next_line.startswith('*') or re.match(r'^\d+\.', next_line)
                if next_line != '' and not next_is_list:
                    processed.append('')
        else:
            processed.append(line)

    final = []
    for line in processed:
        if line.strip() == '' and final and final[-1].strip() == '':
            continue
        final.append(line + '\n')

    with open(filepath, 'w', encoding='utf-8') as f:
        f.writelines(license_header)
        f.write('\n')
        f.writelines(final)

files = [
    'docs/02_strategic_vision.md',
    'docs/03_feature_inventory.md',
    'docs/research/market-sync-2026-06-14.md',
    'docs/features/design-hardware-locked-coordination-handshake.md',
    'docs/features/design-sci-interceptor.md',
    'server/roadmap.md',
    'ui/roadmap.md'
]

for f in files:
    if os.path.exists(f):
        super_fix(f)
