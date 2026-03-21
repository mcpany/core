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

    # Step 1: Ensure Apache License header is at the top in a comment block
    # Check if license exists. If not, add it.
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
-->

"""
    if "Licensed under the Apache License" not in content:
        content = license_header + content
    else:
        # If it's there but not at the top, or incorrectly formatted, we might need to be careful.
        # For now, if it's there, assume it's okay but fix spacing.
        pass

    # Step 2: Ensure single H1 per file (MD025)
    # The roadmaps have multiple H1s (Mission Statement, Core Pillars, etc are H2).
    # Vision and Feature Inventory use #, ##.
    # Check if license header is causing MD025 (it shouldn't as it's a comment).

    # Step 3: Ensure blank lines around headers (MD022)
    # This regex ensures at least one blank line before any # except at start of string.
    content = re.sub(r'([^\n])\n(#+ )', r'\1\n\n\2', content)
    # Ensure at least one blank line after any # line.
    content = re.sub(r'((?:^|\n)#+ [^\n]+)\n([^\n])', r'\1\n\n\2', content)

    # Step 4: Ensure blank lines around lists (MD032)
    # Blank line before list
    content = re.sub(r'([^\n])\n([*-] )', r'\1\n\n\2', content)
    # Blank line after list
    content = re.sub(r'(\n[*-] [^\n]+)\n([^\n*-])', r'\1\n\n\2', content)

    # Step 5: Collapse multiple blank lines (MD012)
    content = re.sub(r'\n{3,}', '\n\n', content)

    # Step 6: Non-ASCII characters removal
    content = content.replace('—', '--').replace('‘', "'").replace('’', "'").replace('“', '"').replace('”', '"').replace('…', '...')
    content = content.encode('ascii', 'ignore').decode('ascii')

    # Step 7: Final newline (MD047)
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
