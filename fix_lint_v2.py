import os
import re

def fix_markdown(filepath):
    if not os.path.exists(filepath):
        print(f"File {filepath} not found.")
        return

    with open(filepath, 'r', encoding='utf-8') as f:
        lines = f.readlines()

    new_lines = []
    for i, line in enumerate(lines):
        # Fix trailing spaces (MD009)
        line = line.rstrip() + '\n'

        # Replace non-ASCII characters
        line = line.replace('—', '--').replace('‘', "'").replace('’', "'").replace('“', '"').replace('”', '"').replace('…', '...')

        # Ensure blank line before headers (MD022)
        if line.startswith('#') and i > 0 and new_lines[-1].strip() != '':
            new_lines.append('\n')

        new_lines.append(line)

        # Ensure blank line after headers (MD022)
        if line.startswith('#') and i < len(lines) - 1 and lines[i+1].strip() != '' and not lines[i+1].startswith('#'):
             new_lines.append('\n')

    # Re-process for lists (MD032)
    final_lines = []
    for i, line in enumerate(new_lines):
        stripped = line.strip()
        is_list_item = stripped.startswith('- ') or stripped.startswith('* ') or re.match(r'^\d+\. ', stripped)

        if is_list_item:
            # Check previous line
            if i > 0 and final_lines[-1].strip() != '' and not (final_lines[-1].strip().startswith('- ') or final_lines[-1].strip().startswith('* ') or re.match(r'^\d+\. ', final_lines[-1].strip())):
                final_lines.append('\n')

        final_lines.append(line)

        if is_list_item:
            # Check next line
            if i < len(new_lines) - 1 and new_lines[i+1].strip() != '' and not (new_lines[i+1].strip().startswith('- ') or new_lines[i+1].strip().startswith('* ') or re.match(r'^\d+\. ', new_lines[i+1].strip())):
                 final_lines.append('\n')

    content = "".join(final_lines)
    # Collapse multiple blank lines (MD012)
    content = re.sub(r'\n{3,}', '\n\n', content)
    # Ensure single trailing newline (MD047)
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
