import os
import re

def fix_markdown(filepath):
    if not os.path.exists(filepath):
        print(f"File {filepath} not found.")
        return

    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()

    # Step 0: Standardize line endings and remove trailing whitespace
    content = content.replace('\r\n', '\n')
    content = re.sub(r'[ \t]+$', '', content, flags=re.MULTILINE)

    # Step 1: Normalize Apache License header at the top
    license_header = """<!--
Copyright 2026 Author(s) of MCP Any

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
-->"""

    # Remove any existing license comment block to avoid duplicates
    content = re.sub(r'<!--.*?Licensed under the Apache License.*?-->\s*', '', content, flags=re.DOTALL)
    # Re-insert at the very top
    content = license_header + "\n\n" + content

    # Step 2: Ensure single H1 per file (MD025)
    lines = content.split('\n')
    h1_found = False
    new_lines = []
    for line in lines:
        if line.startswith('# '):
            if h1_found:
                new_lines.append(line.replace('# ', '## '))
            else:
                new_lines.append(line)
                h1_found = True
        else:
            new_lines.append(line)
    content = '\n'.join(new_lines)

    # Step 3: MD022 (Headers) and MD032 (Lists) spacing
    # Clean up excess newlines first
    content = re.sub(r'\n{3,}', '\n\n', content)

    # MD022: Blank lines around headers
    content = re.sub(r'([^\n])\n(#+ )', r'\1\n\n\2', content)
    content = re.sub(r'(#+ [^\n]+)\n([^\n])', r'\1\n\n\2', content)

    # MD032: Blank lines around list blocks
    # Ensure blank line before list if previous is not a list item
    content = re.sub(r'([^\n\*\-\s\d])\n([\*\-] |\d+\. )', r'\1\n\n\2', content)
    # Ensure blank line after list block if next is not a list item
    content = re.sub(r'(\n(?:[\*\-] |\d+\. )[^\n]+)\n([^\n\*\-\s\d])', r'\1\n\n\2', content)

    # Step 5: Encoding (ASCII Only)
    content = content.replace('—', '--').replace('‘', "'").replace('’', "'").replace('“', '"').replace('”', '"').replace('…', '...')
    content = content.encode('ascii', 'ignore').decode('ascii')

    # Step 6: MD047 (Single EOF newline)
    content = content.rstrip() + '\n'

    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)
    print(f"Processed {filepath}")

files = [
    "docs/research/market-sync-2026-06-05.md",
    "docs/02_strategic_vision.md",
    "docs/03_feature_inventory.md",
    "docs/features/design-scpa.md",
    "server/roadmap.md",
    "ui/roadmap.md"
]

for f in files:
    fix_markdown(f)
