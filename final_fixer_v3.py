import os
import re

def super_fix(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        lines = f.readlines()

    # 1. Identify existing license
    license_header = [
        "# Copyright 2026 Author(s) of MCP Any\n",
        "# SPDX-License-Identifier: Apache-2.0\n"
    ]

    start_idx = 0
    # Check if file starts with a license
    if len(lines) >= 2 and lines[0].startswith("# Copyright") and lines[1].startswith("# SPDX-License"):
        start_idx = 2
        # Check if there is a blank line after license
        if len(lines) > 2 and lines[2].strip() == '':
            start_idx = 3

    content_lines = lines[start_idx:]

    # 2. Process content
    processed = []

    def is_list_line(s):
        s = s.strip()
        return s.startswith('- ') or s.startswith('* ') or re.match(r'^\d+\. ', s)

    def is_heading_line(s):
        return s.strip().startswith('#')

    for i, line in enumerate(content_lines):
        line = line.rstrip()
        # Non-ASCII replacement
        line = line.replace('—', '-').replace('–', '-').replace('\u201c', '"').replace('\u201d', '"').replace('\u2018', "'").replace('\u2019', "'")

        # Heading Blank Lines
        if is_heading_line(line):
            if processed and processed[-1].strip() != '':
                processed.append('')
            processed.append(line)
            # Add blank after heading if next line is not blank
            if i < len(content_lines) - 1 and content_lines[i+1].strip() != '':
                processed.append('')
            continue

        # List Blank Lines (MD032)
        if is_list_line(line):
            # If start of list: needs blank before
            if processed and processed[-1].strip() != '' and not is_list_line(processed[-1]):
                processed.append('')
            processed.append(line)
            # If end of list: needs blank after
            if i < len(content_lines) - 1:
                next_line = content_lines[i+1].strip()
                if next_line != '' and not is_list_line(next_line):
                    processed.append('')
            continue

        processed.append(line)

    # 3. Dedup blank lines and strip trailing whitespace again
    final = []
    for line in processed:
        if line.strip() == '' and final and final[-1].strip() == '':
            continue
        final.append(line + '\n')

    # 4. Write back with EXACTLY one license
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
        print(f"Super-fixing {f}...")
        super_fix(f)
