import re
import os

def fix_file(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        lines = f.readlines()

    new_lines = []
    seen_headings = {}

    for line in lines:
        line = line.rstrip()

        # Handle Heading
        if line.startswith('#'):
            level = len(line) - len(line.lstrip('#'))
            title = line.strip('#').strip()
            if title in seen_headings:
                seen_headings[title] += 1
                line = f"{'#' * level} {title} (v{seen_headings[title]})"
            else:
                seen_headings[title] = 1

            if new_lines and new_lines[-1].strip() != '':
                new_lines.append('')
            new_lines.append(line)
            new_lines.append('')
            continue

        new_lines.append(line)

    content = "\n".join(new_lines)
    # Deduplicate blank lines
    content = re.sub(r'\n{3,}', '\n\n', content)

    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content.strip() + "\n")

files = ['docs/03_feature_inventory.md', 'server/roadmap.md', 'ui/roadmap.md']
for f in files:
    fix_file(f)
