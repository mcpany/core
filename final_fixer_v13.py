import os
import re

def super_fix(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        lines = f.readlines()

    # Identify and strip any existing license or empty lines at start
    start_idx = 0
    while start_idx < len(lines):
        line = lines[start_idx].strip()
        if line.startswith("# Copyright") or line.startswith("# SPDX-License") or line.startswith("<!--") or line.startswith("-->") or line == "Copyright 2026 Author(s) of MCP Any" or line == "SPDX-License-Identifier: Apache-2.0" or line == "":
            start_idx += 1
        else:
            break

    content_lines = lines[start_idx:]
    processed = []

    def is_list_line(s):
        s = s.strip()
        return s.startswith('- ') or s.startswith('* ') or re.match(r'^\d+\. ', s)

    def is_heading_line(s):
        return s.strip().startswith('#')

    for i, line in enumerate(content_lines):
        line = line.rstrip()
        line = line.replace('—', '-').replace('–', '-').replace('\u201c', '"').replace('\u201d', '"').replace('\u2018', "'").replace('\u2019', "'")

        # Heading Blank Lines
        if is_heading_line(line):
            if processed and processed[-1].strip() != '':
                processed.append('')
            processed.append(line)
            if i < len(content_lines) - 1 and content_lines[i+1].strip() != '':
                processed.append('')
            continue

        # List formatting (MD030, MD032)
        if is_list_line(line):
            # Normalize space after marker: handle "-   " -> "- "
            match = re.match(r'^(\s*[-*]|\s*\d+\.)\s+(.*)', line)
            if match:
                marker = match.group(1)
                rest = match.group(2)
                line = f"{marker} {rest}"

            # If start of list: needs blank before
            prev_is_list = processed and (processed[-1].strip().startswith('- ') or processed[-1].strip().startswith('* ') or re.match(r'^\d+\. ', processed[-1].strip()))
            if processed and processed[-1].strip() != '' and not prev_is_list:
                processed.append('')
            processed.append(line)
            # If end of list: needs blank after
            if i < len(content_lines) - 1:
                next_line = content_lines[i+1].strip()
                next_is_list = next_line.startswith('- ') or next_line.startswith('* ') or re.match(r'^\d+\. ', next_line)
                if next_line != '' and not next_is_list:
                    processed.append('')
            continue

        processed.append(line)

    final = []
    for line in processed:
        if line.strip() == '' and final and final[-1].strip() == '':
            continue
        final.append(line + '\n')

    # Apache License 2.0 block (standard comment)
    license_block = [
        "<!--\n",
        "Copyright 2026 Author(s) of MCP Any\n",
        "SPDX-License-Identifier: Apache-2.0\n",
        "-->\n",
        "\n"
    ]

    with open(filepath, 'w', encoding='utf-8') as f:
        f.writelines(license_block)
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
