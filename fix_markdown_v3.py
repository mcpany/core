import os
import re
import textwrap

def wrap_text(text, width=80):
    lines = text.split('\n')
    wrapped_lines = []
    in_code_block = False
    in_table = False

    for line in lines:
        if line.strip().startswith('```'):
            in_code_block = not in_code_block
            wrapped_lines.append(line)
            continue

        if in_code_block:
            wrapped_lines.append(line)
            continue

        if line.strip().startswith('|'):
            in_table = True
            # Ensure tables have leading/trailing pipes if they are supposed to
            l = line.strip()
            if not l.startswith('|'): l = '| ' + l
            if not l.endswith('|'): l = l + ' |'
            wrapped_lines.append(l)
            continue
        elif in_table and not line.strip().startswith('|'):
            in_table = False

        if line.strip().startswith('#') or not line.strip():
            wrapped_lines.append(line)
            continue

        # Check for list markers
        match = re.match(r'^(\s*[-*+]\s+)(.*)', line)
        if match:
            marker, content = match.groups()
            indent = len(marker)
            # Don't wrap if it's a list item that's very short or looks like a header (unlikely here)
            wrapped = textwrap.fill(content, width=width, initial_indent='', subsequent_indent=' ' * indent)
            wrapped_lines.append(marker + wrapped)
        else:
            wrapped_lines.append(textwrap.fill(line, width=width))

    return '\n'.join(wrapped_lines)

def fix_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    # MD047: End with newline
    if not content.endswith('\n'):
        content += '\n'

    # Fix MD030 (Space after list markers)
    content = re.sub(r'^(\s*[-*+])\s{2,}', r'\1 ', content, flags=re.MULTILINE)

    # MD022: Headers should be surrounded by blank lines
    content = re.sub(r'([^\n])\n(#{1,6} )', r'\1\n\n\2', content)
    content = re.sub(r'(#{1,6} .*)\n([^\n])', r'\1\n\n\2', content)

    # MD032: Lists should be surrounded by blank lines
    content = re.sub(r'([^\n])\n([-*+] )', r'\1\n\n\2', content)
    content = re.sub(r'([-*+] .*)\n([^\n])', r'\1\n\n\2', content)

    # Disable MD024 for files with intentional duplicates
    if "Evolution:" in content and "<!-- markdownlint-disable MD024 -->" not in content:
        content = "<!-- markdownlint-disable MD024 -->\n" + content

    # Apply wrapping
    content = wrap_text(content)

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
